from __future__ import annotations

import asyncio
import json
import logging
import os
import re
import urllib.request
import zipfile
from datetime import datetime, timezone
from pathlib import Path

from storage import RustFS, safe_filename


BLOCKED_TERMS = {
    "色情", "成人视频", "破解软件", "crack", "keygen", "木马", "勒索软件", "stealer",
    "盗号", "窃取数据", "儿童色情", "毒品", "枪支教程",
}
DEFAULT_TYPES = {"pdf", "epub", "mobi", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "zip", "txt", "csv", "rtf", "md", "json", "xml", "odt", "ods", "odp", "jpg", "jpeg", "png", "webp"}
BLOCKED_ARCHIVE_EXTENSIONS = {".exe", ".dll", ".scr", ".bat", ".cmd", ".com", ".msi", ".ps1", ".vbs", ".js", ".jar", ".apk"}
LOGGER = logging.getLogger(__name__)


def request_ai_json(config: dict, system_prompt: str, payload: dict) -> dict | None:
    base_url = config["base_url"].rstrip("/")
    api_key = config["api_key"]
    model = config["model"]
    if not base_url or not api_key or not model:
        return None
    body = json.dumps({
        "model": model,
        "temperature": 0,
        "response_format": {"type": "json_object"},
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": json.dumps(payload, ensure_ascii=False)},
        ],
    }, ensure_ascii=False).encode()
    request = urllib.request.Request(f"{base_url}/chat/completions", data=body, headers={"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(request, timeout=12) as response:
            result = json.load(response)
        parsed = json.loads(result["choices"][0]["message"]["content"])
        return parsed if isinstance(parsed, dict) else None
    except Exception as exc:
        LOGGER.warning("AI request failed: %s", exc)
        return None


def rank_sources(keyword: str, dialogs: list[dict]) -> list[dict]:
    query = keyword.casefold().strip()
    tokens = set(re.findall(r"[a-z0-9]+|[\u4e00-\u9fff]{2,}", query))

    def score(dialog: dict) -> tuple[int, object, str]:
        text = f"{dialog.get('name') or ''} {dialog.get('username') or ''}".casefold()
        relevance = int(query in text) * 100 + sum(int(token in text) * 10 for token in tokens)
        return relevance, dialog.get("last_message_at") or datetime.min.replace(tzinfo=timezone.utc), text

    return sorted(dialogs, key=score, reverse=True)


class ResourceService:
    def __init__(self, telegram, db, storage: RustFS):
        self.telegram = telegram
        self.db = db
        self.storage = storage
        self.queue: asyncio.Queue[int] = asyncio.Queue()
        self.workers: list[asyncio.Task] = []
        self.concurrency = min(max(int(os.getenv("TG_RESOURCE_CONCURRENCY", "1")), 1), 3)
        self.max_file_size = int(os.getenv("TG_RESOURCE_MAX_FILE_MB", "200")) * 1024 * 1024
        configured_tmp = os.getenv("TELEGRAM_TMP_ROOT")
        module_path = Path(__file__).resolve()
        default_tmp = module_path.parents[2] / "tmp" / "telegram" if len(module_path.parents) > 2 else Path.cwd() / "tmp" / "telegram"
        self.tmp_root = Path(configured_tmp or default_tmp).resolve()

    def ai_settings(self) -> dict[str, str]:
        return {
            "base_url": (self.db.setting("ai_base_url") or os.getenv("TG_RESOURCE_AI_BASE_URL", "")).strip(),
            "api_key": (self.db.setting("ai_api_key") or os.getenv("TG_RESOURCE_AI_API_KEY", "")).strip(),
            "model": (self.db.setting("ai_model") or os.getenv("TG_RESOURCE_AI_MODEL", "")).strip(),
        }

    async def probe_ai(self, config: dict[str, str]) -> bool:
        result = await asyncio.to_thread(request_ai_json, config, "仅返回 JSON：{\"ok\":true}。", {"action": "connection_test"})
        return result is not None

    async def start(self) -> None:
        self.tmp_root.mkdir(parents=True, exist_ok=True)
        self.workers = [asyncio.create_task(self._worker()) for _ in range(self.concurrency)]
        for task_id in self.db.queued_resource_tasks():
            self.enqueue(task_id)

    async def stop(self) -> None:
        for worker in self.workers:
            worker.cancel()
        await asyncio.gather(*self.workers, return_exceptions=True)

    def enqueue(self, task_id: int) -> None:
        self.queue.put_nowait(task_id)

    async def plan_sources(self, keyword: str, source_mode: str, selected_ids: list[int]) -> tuple[list[int], str]:
        dialogs = self.db.resource_dialogs()
        if source_mode == "manual":
            return selected_ids, "manual"
        if not dialogs:
            raise ValueError("没有可检索的公开正常群组或频道，请先完成扫描")
        if source_mode == "all":
            return [item["id"] for item in dialogs], "all"
        ranked = rank_sources(keyword, dialogs)
        ai_ids = await asyncio.to_thread(self._ai_source_ids, keyword, ranked)
        if ai_ids is None:
            return [item["id"] for item in ranked], "metadata"
        by_id = {item["id"]: item for item in ranked}
        ordered = list(dict.fromkeys(source_id for source_id in ai_ids if source_id in by_id))
        selected = set(ordered)
        ordered.extend(item["id"] for item in ranked if item["id"] not in selected)
        return ordered, "ai"

    def _ai_source_ids(self, keyword: str, dialogs: list[dict]) -> list[int] | None:
        sources = [{key: item.get(key) for key in ("id", "name", "username", "dialog_type")} for item in dialogs]
        result = request_ai_json(self.ai_settings(), "根据来源名称判断与检索主题的相关性，返回 JSON：{\"source_ids\":[按相关性降序的来源ID]}。不要编造ID。", {"keyword": keyword, "sources": sources})
        if result is None:
            return None
        source_ids = []
        for value in result.get("source_ids", []):
            try:
                source_ids.append(int(value))
            except (TypeError, ValueError):
                continue
        return source_ids

    async def refresh_adult_dialogs(self) -> str:
        dialogs = self.db.adult_classification_dialogs()
        if not dialogs:
            self.db.save_adult_dialog_ids([])
            return "ai"
        prompt = "仅根据群组或频道的名称、用户名判断是否可能主要提供露骨色情或成人性内容。保守识别，不要把普通两性、健康或教育内容误判。返回 JSON：{\"adult_dialog_ids\":[来源ID]}，不要编造ID。"
        result = await asyncio.to_thread(request_ai_json, self.ai_settings(), prompt, {"dialogs": dialogs})
        if result is None:
            return "cached"
        valid_ids = {dialog["id"] for dialog in dialogs}
        adult_ids = []
        for value in result.get("adult_dialog_ids", []):
            try:
                dialog_id = int(value)
            except (TypeError, ValueError):
                continue
            if dialog_id in valid_ids and dialog_id not in adult_ids:
                adult_ids.append(dialog_id)
        self.db.save_adult_dialog_ids(adult_ids)
        return "ai"

    def cancel(self, task_id: int) -> None:
        task = self.db.resource_task(task_id)
        if not task or task["status"] not in {"queued", "running"}:
            raise ValueError("任务不存在或当前状态不能取消")
        if task["status"] == "queued":
            self.db.update_resource_task(task_id, status="cancelled", cancel_requested=True, completed_at=datetime.now(timezone.utc))
        else:
            self.db.update_resource_task(task_id, cancel_requested=True)

    def retry(self, task_id: int) -> None:
        task = self.db.resource_task(task_id)
        if not task or task["status"] not in {"failed", "cancelled"}:
            raise ValueError("仅失败或已取消任务可以重试")
        found, completed = self.db.prepare_resource_retry(task_id)
        self.db.update_resource_task(task_id, status="queued", files_found=found, files_completed=completed, error_message=None, completed_at=None, cancel_requested=False)
        self.enqueue(task_id)

    async def _worker(self) -> None:
        while True:
            task_id = await self.queue.get()
            try:
                await self._run(task_id)
            finally:
                self.queue.task_done()

    async def _run(self, task_id: int) -> None:
        task = self.db.resource_task(task_id)
        if not task or task["status"] != "queued":
            return
        self.db.update_resource_task(task_id, status="running", started_at=datetime.now(timezone.utc), error_message=None)
        try:
            if not self.telegram.authorized:
                await self.telegram.connect()
            if not self.telegram.authorized:
                raise RuntimeError("Telegram 未登录")
            await self.telegram._ensure_connected()
            completed = task["files_completed"]
            found = task["files_found"]
            failed = 0
            for dialog in task["sources"]:
                if dialog["status"] != "normal" or dialog["resource_compliance_status"] != "allowed" or not dialog["username"]:
                    continue
                if self._cancelled(task_id):
                    self.db.update_resource_task(task_id, status="cancelled", completed_at=datetime.now(timezone.utc))
                    return
                entity = await self.telegram.call(self.telegram.client.get_entity, dialog["telegram_id"])
                async for message in self.telegram.client.iter_messages(entity, search=task["keyword"], limit=max(task["max_downloads"] * 20, 100)):
                    if completed >= task["max_downloads"]:
                        break
                    if self._cancelled(task_id):
                        self.db.update_resource_task(task_id, status="cancelled", completed_at=datetime.now(timezone.utc))
                        return
                    meta = self._match(message, task)
                    if not meta:
                        continue
                    record = self.db.claim_resource_file(task_id, dialog, message.id, meta["file_id"], meta["filename"], meta["size"], meta["mime_type"], message.raw_text or "")
                    if not record:
                        continue
                    found += 1
                    self.db.update_resource_task(task_id, files_found=found)
                    if await self._download(task_id, record, message):
                        completed += 1
                        self.db.update_resource_task(task_id, files_completed=completed)
                    else:
                        failed += 1
                if completed >= task["max_downloads"]:
                    break
            self.db.update_resource_task(task_id, status="failed" if failed else "completed", error_message=f"{failed} 个文件下载或上传失败" if failed else None, completed_at=datetime.now(timezone.utc))
        except asyncio.CancelledError:
            raise
        except Exception as exc:
            self.db.update_resource_task(task_id, status="failed", error_message=str(exc)[:2000], completed_at=datetime.now(timezone.utc))

    def _match(self, message, task: dict) -> dict | None:
        document = getattr(message, "document", None)
        file = getattr(message, "file", None)
        filename = getattr(file, "name", None)
        if not document or not filename:
            return None
        cleaned = safe_filename(filename)
        extension = cleaned.rsplit(".", 1)[-1].lower() if "." in cleaned else ""
        haystack = f"{cleaned}\n{message.raw_text or ''}".lower()
        if task["keyword"].lower() not in haystack or task["file_types"] and extension not in set(task["file_types"]):
            return None
        if any(term in haystack for term in BLOCKED_TERMS):
            return None
        size = int(getattr(file, "size", 0) or 0)
        if size <= 0 or size > self.max_file_size:
            return None
        return {"filename": cleaned, "size": size, "mime_type": getattr(file, "mime_type", None), "file_id": str(document.id)}

    async def _download(self, task_id: int, record: dict, message) -> bool:
        task_dir = self.tmp_root / str(task_id)
        task_dir.mkdir(parents=True, exist_ok=True)
        path = task_dir / f"{record['id']}-{safe_filename(record['original_filename'])}"
        self.db.update_resource_file(record["id"], download_status="downloading", attempts=record["attempts"] + 1)
        try:
            result = await self.telegram.call(self.telegram.client.download_media, message, file=str(path))
            if not result or not path.is_file():
                raise RuntimeError("Telegram 下载未生成文件")
            if path.stat().st_size > self.max_file_size:
                raise RuntimeError("文件超过大小限制")
            self._validate_download(path, record["file_type"])
            now = datetime.now(timezone.utc)
            object_key = f"telegram/{now:%Y}/{now:%m}/{task_id}/{record['id']}-{safe_filename(record['original_filename'])}"
            self.db.update_resource_file(record["id"], bytes_downloaded=path.stat().st_size, download_status="success", upload_status="uploading")
            await asyncio.to_thread(self.storage.upload, path, object_key, record["mime_type"])
            self.db.update_resource_file(record["id"], upload_status="success", object_key=object_key, completed_at=now, error_message=None)
            return True
        except Exception as exc:
            self.db.update_resource_file(record["id"], download_status="failed", upload_status="failed", error_message=str(exc)[:2000], completed_at=datetime.now(timezone.utc))
            return False
        finally:
            path.unlink(missing_ok=True)
            try:
                task_dir.rmdir()
            except OSError:
                pass

    def _validate_download(self, path: Path, file_type: str) -> None:
        if file_type != "zip":
            return
        if not zipfile.is_zipfile(path):
            raise RuntimeError("ZIP 文件格式无效")
        with zipfile.ZipFile(path) as archive:
            entries = archive.infolist()
            if len(entries) > 10_000 or sum(item.file_size for item in entries) > self.max_file_size * 5:
                raise RuntimeError("ZIP 内容超过安全限制")
            if any(item.flag_bits & 1 or Path(item.filename).suffix.lower() in BLOCKED_ARCHIVE_EXTENSIONS for item in entries):
                raise RuntimeError("ZIP 包含加密内容或明显危险文件")

    def _cancelled(self, task_id: int) -> bool:
        task = self.db.resource_task(task_id)
        return not task or task["cancel_requested"]
