from __future__ import annotations

import asyncio
import base64
import binascii
import hashlib
import http.client
import json
import logging
import os
import re
import shutil
import tempfile
import urllib.request
from urllib.parse import urlsplit
import zipfile
from datetime import datetime, timedelta, timezone
from pathlib import Path

from telethon import errors

from storage import RustFS, safe_filename


BLOCKED_TERMS = {
    "色情", "成人视频", "破解软件", "crack", "keygen", "木马", "勒索软件", "stealer",
    "盗号", "窃取数据", "儿童色情", "毒品", "枪支教程",
}
FILE_TYPE_DICTIONARY = {
    "pdf": {"label": "PDF", "category": "document", "preview": "pdf"},
    "md": {"label": "Markdown", "category": "document", "preview": "text"},
    "txt": {"label": "文本", "category": "document", "preview": "text"},
    "csv": {"label": "CSV", "category": "document", "preview": "text"},
    "json": {"label": "JSON", "category": "document", "preview": "text"},
    "xml": {"label": "XML", "category": "document", "preview": "text"},
    "epub": {"label": "EPUB", "category": "document", "preview": None},
    "mobi": {"label": "MOBI", "category": "document", "preview": None},
    "doc": {"label": "Word 97-2003", "category": "document", "preview": None},
    "docx": {"label": "Word", "category": "document", "preview": None},
    "xls": {"label": "Excel 97-2003", "category": "document", "preview": None},
    "xlsx": {"label": "Excel", "category": "document", "preview": None},
    "ppt": {"label": "PowerPoint 97-2003", "category": "document", "preview": None},
    "pptx": {"label": "PowerPoint", "category": "document", "preview": None},
    "rtf": {"label": "RTF", "category": "document", "preview": None},
    "odt": {"label": "OpenDocument 文本", "category": "document", "preview": None},
    "ods": {"label": "OpenDocument 表格", "category": "document", "preview": None},
    "odp": {"label": "OpenDocument 演示", "category": "document", "preview": None},
    "jpg": {"label": "JPEG 图片", "category": "image", "preview": "image"},
    "jpeg": {"label": "JPEG 图片", "category": "image", "preview": "image"},
    "png": {"label": "PNG 图片", "category": "image", "preview": "image"},
    "webp": {"label": "WebP 图片", "category": "image", "preview": "image"},
    "mp4": {"label": "MP4 视频", "category": "video", "preview": "video"},
    "webm": {"label": "WebM 视频", "category": "video", "preview": "video"},
    "mov": {"label": "MOV 视频", "category": "video", "preview": "video"},
    "m4v": {"label": "M4V 视频", "category": "video", "preview": "video"},
    "mp3": {"label": "MP3 音频", "category": "audio", "preview": "audio"},
    "m4a": {"label": "M4A 音频", "category": "audio", "preview": "audio"},
    "wav": {"label": "WAV 音频", "category": "audio", "preview": "audio"},
    "ogg": {"label": "OGG 音频", "category": "audio", "preview": "audio"},
    "zip": {"label": "ZIP 压缩包", "category": "archive", "preview": None},
}
DEFAULT_TYPES = set(FILE_TYPE_DICTIONARY)
BLOCKED_ARCHIVE_EXTENSIONS = {".exe", ".dll", ".scr", ".bat", ".cmd", ".com", ".msi", ".ps1", ".vbs", ".js", ".jar", ".apk"}
LOGGER = logging.getLogger(__name__)
CANDIDATE_TTL = timedelta(hours=24)
DOWNLOAD_TIMEOUT_SECONDS = 1800
TELEGRAM_SOURCE_PATTERN = re.compile(r"(?:https?://)?t\.me/([A-Za-z][A-Za-z0-9_]{4,})", re.IGNORECASE)


class ResourceCapabilityError(Exception):
    def __init__(self, code: str, message: str, retryable: bool = False):
        super().__init__(message)
        self.code = code
        self.retryable = retryable


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


def is_interactive_search_group(dialog: dict) -> bool:
    return bool(dialog.get("is_group")) and "搜" in (dialog.get("name") or "")


def message_source_usernames(message) -> set[str]:
    values = [getattr(message, "raw_text", "") or ""]
    for entity in getattr(message, "entities", None) or []:
        values.append(getattr(entity, "url", "") or "")
    for row in getattr(message, "buttons", None) or []:
        for button in row if isinstance(row, (list, tuple)) else [row]:
            values.append(getattr(button, "url", "") or "")
    return {match.group(1).casefold() for value in values for match in TELEGRAM_SOURCE_PATTERN.finditer(value)}


def _decode_base64(value: str) -> bytes:
    return base64.urlsafe_b64decode(value + "=" * (-len(value) % 4))


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _put_file(upload_url: str, path: Path, mime_type: str | None) -> None:
    target = urlsplit(upload_url)
    if target.scheme not in {"http", "https"} or not target.hostname:
        raise ResourceCapabilityError("delivery_failed", "上传地址无效")
    connection_type = http.client.HTTPSConnection if target.scheme == "https" else http.client.HTTPConnection
    connection = connection_type(target.hostname, target.port, timeout=30)
    request_path = target.path + (f"?{target.query}" if target.query else "")
    try:
        connection.putrequest("PUT", request_path)
        connection.putheader("Content-Length", str(path.stat().st_size))
        connection.putheader("Content-Type", mime_type or "application/octet-stream")
        connection.endheaders()
        with path.open("rb") as source:
            for chunk in iter(lambda: source.read(1024 * 1024), b""):
                connection.send(chunk)
        response = connection.getresponse()
        response.read()
        if response.status < 200 or response.status >= 300:
            raise ResourceCapabilityError("delivery_failed", "资源上传失败", retryable=response.status >= 500)
    finally:
        connection.close()


class ResourceService:
    def __init__(self, telegram, db, storage: RustFS):
        self.telegram = telegram
        self.db = db
        self.storage = storage
        self.queue: asyncio.Queue[int] = asyncio.Queue()
        self.workers: list[asyncio.Task] = []
        self.active_downloads: dict[int, asyncio.Task] = {}
        self.acquisitions: dict[int, asyncio.Task] = {}
        self.interactive_search_lock = asyncio.Lock()
        self.interactive_search_wait = min(max(float(os.getenv("TG_INTERACTIVE_SEARCH_WAIT_SECONDS", "8")), 1), 30)
        self.concurrency = min(max(int(os.getenv("TG_RESOURCE_CONCURRENCY", "1")), 1), 3)
        self.slots = asyncio.Semaphore(self.concurrency)
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
        for task in self.acquisitions.values():
            task.cancel()
        await asyncio.gather(*self.acquisitions.values(), return_exceptions=True)

    def enqueue(self, task_id: int) -> None:
        self.queue.put_nowait(task_id)

    async def submit_acquisition(self, candidate_id: str, resource_task_id: int, idempotency_key: str, delivery_token: str) -> dict:
        if resource_task_id < 1 or not idempotency_key.strip() or not delivery_token:
            raise ResourceCapabilityError("invalid_argument", "资源任务参数不完整")
        await self.inspect_candidate(candidate_id)
        task, created = self.db.create_acquisition_task(resource_task_id, candidate_id, idempotency_key.strip())
        if created:
            worker = asyncio.create_task(self._acquire_when_ready(task["id"], candidate_id, resource_task_id, delivery_token))
            self.acquisitions[task["id"]] = worker
            worker.add_done_callback(lambda _worker, task_id=task["id"]: self.acquisitions.pop(task_id, None))
        return task

    async def _acquire_when_ready(self, task_id, candidate_id, resource_task_id, delivery_token):
        async with self.slots:
            await self._acquire(task_id, candidate_id, resource_task_id, delivery_token)

    def cancel_acquisition(self, task_id: int) -> None:
        task = self.db.acquisition_task(task_id)
        if not task or task["state"] not in {"queued", "running"}:
            raise ValueError("任务不存在或当前状态不能取消")
        worker = self.acquisitions.get(task_id)
        if worker:
            worker.cancel()
        self.db.update_acquisition_task(task_id, state="cancelled", completed_at=datetime.now(timezone.utc))

    async def _acquire(self, task_id: int, candidate_id: str, resource_task_id: int, delivery_token: str) -> None:
        task_dir = self.tmp_root / f"acquire-{task_id}"
        path = None
        self.db.update_acquisition_task(task_id, state="running", stage="downloading", started_at=datetime.now(timezone.utc))
        await asyncio.to_thread(self._notify_backend, f"/xy/internal/resource-tasks/{resource_task_id}/progress", delivery_token, {
            "stage": "downloading", "progressPercent": 0, "progressDone": 0, "progressTotal": 0, "progressUnit": "bytes",
        })
        try:
            candidate = self._decode_candidate(candidate_id)
            dialog = self.db.dialog(candidate["source_id"])
            if not dialog:
                raise ResourceCapabilityError("source_not_found", "资源来源不存在")
            await self.telegram._ensure_connected()
            entity = await self._dialog_entity(dialog)
            message = await self.telegram.call(self.telegram.client.get_messages, entity, ids=candidate["message_id"])
            metadata = self._match(message, {"keyword": "", "file_types": []}) if message else None
            if not metadata or metadata["file_id"] != candidate["file_id"]:
                raise ResourceCapabilityError("resource_changed", "资源不存在或已发生变化")

            task_dir.mkdir(parents=True, exist_ok=True)
            path = task_dir / safe_filename(metadata["filename"])
            downloaded = await self.telegram.call(self.telegram.client.download_media, message, file=str(path))
            if not downloaded or not path.is_file():
                raise ResourceCapabilityError("delivery_failed", "资源下载失败", retryable=True)
            size = path.stat().st_size
            if size != metadata["size"] or size > self.max_file_size:
                raise ResourceCapabilityError("resource_changed", "资源大小已发生变化")
            self._validate_download(path, path.suffix.lower().lstrip("."))
            checksum = await asyncio.to_thread(_sha256_file, path)

            self.db.update_acquisition_task(task_id, stage="waiting_upload", bytes_completed=size, bytes_total=size)
            await asyncio.to_thread(self._notify_backend, f"/xy/internal/resource-tasks/{resource_task_id}/progress", delivery_token, {
                "stage": "waiting_upload", "progressPercent": 70, "progressDone": size, "progressTotal": size, "progressUnit": "bytes",
            })
            upload = await asyncio.to_thread(self._backend_json, f"/xy/internal/resource-tasks/{resource_task_id}/upload", delivery_token, None)
            self.db.update_acquisition_task(task_id, stage="uploading")
            await asyncio.to_thread(self._notify_backend, f"/xy/internal/resource-tasks/{resource_task_id}/progress", delivery_token, {
                "stage": "uploading", "progressPercent": 80, "progressDone": size, "progressTotal": size, "progressUnit": "bytes",
            })
            await asyncio.to_thread(_put_file, upload["uploadUrl"], path, metadata["mime_type"])
            completed = await asyncio.to_thread(
                self._backend_json,
                f"/xy/internal/resource-tasks/{resource_task_id}/complete",
                delivery_token,
                {"size": size, "sha256": checksum},
            )
            self.db.update_acquisition_task(
                task_id, state="completed", stage="completed", result_json=json.dumps(completed, ensure_ascii=False),
                completed_at=datetime.now(timezone.utc), error_code=None, error_message=None,
            )
        except asyncio.CancelledError:
            self.db.update_acquisition_task(task_id, state="cancelled", completed_at=datetime.now(timezone.utc))
            raise
        except ResourceCapabilityError as exc:
            self.db.update_acquisition_task(task_id, state="failed", error_code=exc.code, error_message=str(exc), completed_at=datetime.now(timezone.utc))
            await asyncio.to_thread(self._notify_backend, f"/xy/internal/resource-tasks/{resource_task_id}/fail", delivery_token, {
                "errorCode": exc.code, "errorMessage": "资源获取失败",
            })
        except Exception as exc:
            LOGGER.exception("Resource acquisition failed for task %s", task_id)
            self.db.update_acquisition_task(task_id, state="failed", error_code="delivery_failed", error_message=str(exc)[:2000], completed_at=datetime.now(timezone.utc))
            await asyncio.to_thread(self._notify_backend, f"/xy/internal/resource-tasks/{resource_task_id}/fail", delivery_token, {
                "errorCode": "delivery_failed", "errorMessage": "资源获取失败",
            })
        finally:
            if path:
                path.unlink(missing_ok=True)
            try:
                task_dir.rmdir()
            except OSError:
                pass

    def _backend_json(self, path: str, delivery_token: str, payload: dict | None) -> dict:
        base_url = os.getenv("XIANYU_INTERNAL_URL", "http://server:8888").rstrip("/")
        body = json.dumps(payload, ensure_ascii=False).encode() if payload is not None else b"{}"
        request = urllib.request.Request(
            base_url + path,
            data=body,
            headers={"Content-Type": "application/json", "X-Delivery-Token": delivery_token},
            method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=10) as response:
                result = json.load(response)
        except Exception as exc:
            raise ResourceCapabilityError("delivery_grant_failed", "无法获取资源交付授权", retryable=True) from exc
        if result.get("code") != 0 or not isinstance(result.get("data"), dict):
            raise ResourceCapabilityError("delivery_failed", "资源交付服务拒绝请求")
        return result["data"]

    def _notify_backend(self, path: str, delivery_token: str, payload: dict) -> None:
        try:
            self._backend_json(path, delivery_token, payload)
        except Exception:
            LOGGER.warning("Failed to report resource task state to backend: %s", path)

    async def search_candidates(self, query: str, file_types: list[str] | None, limit: int, source_ids: list[int] | None = None) -> list[dict]:
        query = query.strip()
        cleaned_types = list(dict.fromkeys(value.lower().lstrip(".").strip() for value in (file_types or [])))
        if len(query) < 2 or len(query) > 200:
            raise ResourceCapabilityError("invalid_argument", "检索关键词长度必须在 2 到 200 个字符之间")
        if len(cleaned_types) > len(DEFAULT_TYPES) or any(value not in DEFAULT_TYPES for value in cleaned_types):
            raise ResourceCapabilityError("unsupported_file_type", "文件类型不在允许范围内")
        if limit < 1 or limit > 100:
            raise ResourceCapabilityError("invalid_argument", "结果数量必须在 1 到 100 之间")
        if not self.telegram.authorized:
            raise ResourceCapabilityError("authentication_required", "当前能力需要先完成认证")

        dialogs = self.db.resource_dialogs()
        if source_ids:
            requested = set(source_ids)
            dialogs = [dialog for dialog in dialogs if dialog["id"] in requested]
            if {dialog["id"] for dialog in dialogs} != requested:
                raise ResourceCapabilityError("source_not_found", "检索来源不存在或当前不可用")
        if not dialogs:
            raise ResourceCapabilityError("source_unavailable", "当前没有可检索来源")

        await self.telegram._ensure_connected()
        candidates = []
        failed_sources = 0
        ranked = rank_sources(query, dialogs)
        dialogs_by_username = {dialog["username"].casefold(): dialog for dialog in dialogs if dialog.get("username")}
        interactive_queue = [dialog for dialog in ranked if is_interactive_search_group(dialog)][:3]
        interactive_ids = set()
        discovered = []
        while interactive_queue and len(interactive_ids) < 3:
            dialog = interactive_queue.pop(0)
            if dialog["id"] in interactive_ids:
                continue
            interactive_ids.add(dialog["id"])
            try:
                found, usernames = await self._search_interactive_group(dialog, query, cleaned_types, limit - len(candidates))
                candidates.extend(found)
                for username in usernames:
                    linked = dialogs_by_username.get(username)
                    if not linked or linked["id"] in interactive_ids:
                        continue
                    discovered.append(linked)
                    if is_interactive_search_group(linked) and linked not in interactive_queue:
                        interactive_queue.append(linked)
                if len(candidates) >= limit:
                    return candidates
            except Exception as exc:
                if isinstance(exc, errors.FloodWaitError):
                    await self.telegram.record_flood_wait(exc)
                    raise
                failed_sources += 1
                LOGGER.warning("Interactive resource search skipped source %s: %s", dialog["id"], exc)
        history_candidates = discovered + ranked
        history_sources = history_candidates if source_ids else [dialog for dialog in history_candidates if dialog["id"] not in interactive_ids][:5]
        history_sources = list({dialog["id"]: dialog for dialog in history_sources}.values())
        # shortcut: bound automatic history fan-out to avoid Telegram flood waits; make it configurable when operators need broader searches.
        for dialog in history_sources:
            if dialog["id"] in interactive_ids:
                continue
            try:
                entity = await self._dialog_entity(dialog)
                async for message in self.telegram.client.iter_messages(entity, search=query, limit=max(limit * 5, 100)):
                    metadata = self._match(message, {"keyword": query, "file_types": cleaned_types})
                    if not metadata:
                        continue
                    candidates.append(self._candidate(dialog, message, metadata))
                    if len(candidates) >= limit:
                        return candidates
            except Exception as exc:
                if isinstance(exc, errors.FloodWaitError):
                    await self.telegram.record_flood_wait(exc)
                    raise
                failed_sources += 1
                LOGGER.warning("Resource search skipped source %s: %s", dialog["id"], exc)
        if failed_sources == len(dialogs):
            raise ResourceCapabilityError("source_unavailable", "检索来源当前不可用", retryable=True)
        return candidates

    async def _search_interactive_group(self, dialog: dict, query: str, file_types: list[str], limit: int) -> tuple[list[dict], set[str]]:
        if limit < 1:
            return [], set()
        lock = getattr(self, "interactive_search_lock", None)
        if lock is None:
            lock = self.interactive_search_lock = asyncio.Lock()
        wait_seconds = getattr(self, "interactive_search_wait", 8)
        async with lock:
            entity = await self._dialog_entity(dialog)
            sent = await self.telegram.call(self.telegram.client.send_message, entity, query)
            sent_id = int(getattr(sent, "id", 0) or 0)
            if sent_id < 1:
                raise RuntimeError("Telegram 未返回搜索消息标识")
            deadline = asyncio.get_running_loop().time() + wait_seconds
            seen = set()
            found = []
            discovered = set()
            while asyncio.get_running_loop().time() < deadline:
                async for message in self.telegram.client.iter_messages(entity, min_id=sent_id, reverse=True, limit=100):
                    if message.id in seen:
                        continue
                    seen.add(message.id)
                    discovered.update(message_source_usernames(message))
                    metadata = self._match(message, {"keyword": query, "file_types": file_types})
                    if metadata:
                        found.append(self._candidate(dialog, message, metadata))
                        if len(found) >= limit:
                            return found, discovered
                if found or discovered:
                    return found, discovered
                await asyncio.sleep(min(1, wait_seconds))
            return found, discovered

    async def inspect_candidate(self, candidate_id: str) -> dict:
        candidate = self._decode_candidate(candidate_id)
        dialog = self.db.dialog(candidate["source_id"])
        if not dialog or dialog["status"] != "normal" or dialog["resource_compliance_status"] != "allowed":
            raise ResourceCapabilityError("source_unavailable", "资源来源当前不可用")
        if not self.telegram.authorized:
            raise ResourceCapabilityError("authentication_required", "当前能力需要先完成认证")

        await self.telegram._ensure_connected()
        try:
            entity = await self._dialog_entity(dialog)
            message = await self.telegram.call(self.telegram.client.get_messages, entity, ids=candidate["message_id"])
        except Exception as exc:
            raise ResourceCapabilityError("resource_unavailable", "资源当前不可用", retryable=True) from exc
        metadata = self._match(message, {"keyword": "", "file_types": []}) if message else None
        if not metadata:
            raise ResourceCapabilityError("resource_not_found", "资源不存在或已不符合安全限制")
        if metadata["file_id"] != candidate["file_id"]:
            raise ResourceCapabilityError("resource_changed", "资源内容已发生变化")
        return self._candidate(dialog, message, metadata, candidate_id)

    def _candidate(self, dialog: dict, message, metadata: dict, candidate_id: str | None = None) -> dict:
        return {
            "candidate_id": candidate_id or self._encode_candidate(dialog["id"], message.id, metadata["file_id"]),
            "name": metadata["filename"],
            "size": metadata["size"],
            "mime_type": metadata["mime_type"],
            "source": {"id": dialog["id"], "name": dialog["name"], "username": dialog.get("username")},
            "published_at": message.date.isoformat() if getattr(message, "date", None) else None,
            "preview": (message.raw_text or "")[:500],
            "license": "unknown",
            "risk": "来源未提供可验证的许可证信息",
        }

    async def _dialog_entity(self, dialog: dict):
        return await self.telegram.call(self.telegram.client.get_entity, dialog.get("username") or dialog["telegram_id"])

    def _encode_candidate(self, source_id: int, message_id: int, file_id: str) -> str:
        payload = json.dumps({"s": source_id, "m": message_id, "f": file_id, "t": int(datetime.now(timezone.utc).timestamp())}, separators=(",", ":")).encode()
        encoded = base64.urlsafe_b64encode(payload).rstrip(b"=")
        return encoded.decode()

    def _decode_candidate(self, candidate_id: str) -> dict:
        try:
            decoded = _decode_base64(candidate_id)
            if base64.urlsafe_b64encode(decoded).rstrip(b"=").decode() != candidate_id:
                raise ValueError
            payload = json.loads(decoded)
            issued_at = datetime.fromtimestamp(int(payload["t"]), timezone.utc)
            if datetime.now(timezone.utc) - issued_at > CANDIDATE_TTL or issued_at > datetime.now(timezone.utc) + timedelta(minutes=1):
                raise ValueError
            return {"source_id": int(payload["s"]), "message_id": int(payload["m"]), "file_id": str(payload["f"])}
        except (binascii.Error, KeyError, OverflowError, TypeError, ValueError, json.JSONDecodeError) as exc:
            raise ResourceCapabilityError("invalid_argument", "候选 ID 无效或已过期") from exc

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

    def cancel(self, task_id: int, owner_id: int) -> None:
        task = self.db.resource_task(task_id, owner_id)
        if not task or task["status"] not in {"queued", "running"}:
            raise ValueError("任务不存在或当前状态不能取消")
        self.db.update_resource_task(task_id, status="cancelled", cancel_requested=True, completed_at=datetime.now(timezone.utc))
        download = self.active_downloads.get(task_id)
        if download:
            download.cancel()

    def retry(self, task_id: int, owner_id: int) -> None:
        task = self.db.resource_task(task_id, owner_id)
        if not task or task["status"] not in {"failed", "cancelled"}:
            raise ValueError("仅失败或已取消任务可以重试")
        found, completed = self.db.prepare_resource_retry(task_id)
        self.db.update_resource_task(task_id, status="queued", files_found=found, files_completed=completed, error_message=None, completed_at=None, cancel_requested=False)
        self.enqueue(task_id)

    def delete_task(self, task_id: int, delete_files: bool, owner_id: int) -> None:
        task = self.db.resource_task(task_id, owner_id)
        if not task:
            raise ValueError("资源任务不存在")
        if task["status"] in {"queued", "running"}:
            raise ValueError("请先取消任务，再执行删除")
        files = self.db.resource_task_files(task_id, owner_id) if delete_files else []
        for file in files:
            self._delete_telegram_resource(file["id"], owner_id)
        self.db.delete_resource_task(task_id, delete_files, owner_id)

    def delete_file(self, file_id: int, owner_id: int) -> None:
        file = self.db.resource_file(file_id, owner_id)
        if not file:
            raise ValueError("资源文件不存在")
        if file["download_status"] in {"queued", "downloading"} or file["upload_status"] in {"pending", "uploading"}:
            raise ValueError("文件仍在处理中，暂不能删除")
        self._delete_telegram_resource(file_id, owner_id)

    def _delete_telegram_resource(self, file_id: int, owner_id: int) -> None:
        base_url = os.getenv("XIANYU_INTERNAL_URL", "http://server:8888").rstrip("/")
        request = urllib.request.Request(
            f"{base_url}/xy/internal/telegram-resources/files/{file_id}",
            data=json.dumps({"ownerId": owner_id}).encode(),
            headers={"Content-Type": "application/json"},
            method="DELETE",
        )
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                result = json.load(response)
        except Exception as exc:
            raise RuntimeError("Telegram 文件删除失败") from exc
        if result.get("code") != 0:
            raise RuntimeError("资源目录拒绝删除 Telegram 文件")

    async def _worker(self) -> None:
        while True:
            task_id = await self.queue.get()
            try:
                async with self.slots:
                    await self._run(task_id)
            except Exception:
                LOGGER.exception("Resource worker failed for task %s", task_id)
                await asyncio.sleep(1)
                self.enqueue(task_id)
            finally:
                self.queue.task_done()

    async def _run(self, task_id: int) -> None:
        task = self.db.resource_task(task_id)
        if not task or task["status"] != "queued":
            return
        self.db.update_resource_task(task_id, status="running", started_at=datetime.now(timezone.utc), error_message=None)
        flooded_api_ids: set[int] = set()
        try:
            if getattr(self.telegram, "credential_source", "env") == "database":
                await self.telegram.refresh_api_credentials()
            if not self.telegram.authorized:
                if getattr(self.telegram, "credential_source", "env") == "database":
                    await self.telegram.call_with_fallback(self.telegram.connect, flooded_api_ids)
                else:
                    await self.telegram.connect()
            if not self.telegram.authorized:
                raise RuntimeError("Telegram 未登录")
            if getattr(self.telegram, "credential_source", "env") == "database":
                await self.telegram.call_with_fallback(self.telegram._ensure_connected, flooded_api_ids)
            else:
                await self.telegram._ensure_connected()
            if task["source_mode"] == "smart" and task["source_strategy"] == "pending":
                source_ids, strategy = await self.plan_sources(task["keyword"], task["source_mode"], [])
                if self._cancelled(task_id):
                    return
                self.db.update_resource_task(task_id, source_dialog_ids=source_ids, source_strategy=strategy)
                task = self.db.resource_task(task_id)
                if not task:
                    return
            completed = task["files_completed"]
            found = task["files_found"]
            failed = 0
            for dialog in task["sources"]:
                if dialog["status"] != "normal" or dialog["resource_compliance_status"] != "allowed" or not dialog["username"]:
                    continue
                if self._cancelled(task_id):
                    self.db.update_resource_task(task_id, status="cancelled", completed_at=datetime.now(timezone.utc))
                    return
                if getattr(self.telegram, "credential_source", "env") == "database":
                    entity = await self.telegram.call_with_fallback(
                        lambda: self.telegram.client.get_entity(dialog.get("username") or dialog["telegram_id"]),
                        flooded_api_ids,
                    )
                else:
                    entity = await self._dialog_entity(dialog)
                try:
                    async for message in self._resource_messages(entity, task["keyword"], max(task["max_downloads"] * 20, 100), flooded_api_ids):
                        if found >= task["max_downloads"]:
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
                        if await self._download(task_id, record, message, flooded_api_ids):
                            completed += 1
                            self.db.update_resource_task(task_id, files_completed=completed)
                        elif self._cancelled(task_id):
                            return
                        else:
                            failed += 1
                except errors.FloodWaitError:
                    raise
                if found >= task["max_downloads"]:
                    break
            self.db.update_resource_task(task_id, status="failed" if failed else "completed", error_message=f"{failed} 个文件下载或上传失败" if failed else None, completed_at=datetime.now(timezone.utc))
        except asyncio.CancelledError:
            raise
        except Exception as exc:
            message = f"FloodWait: {exc.seconds} seconds" if isinstance(exc, errors.FloodWaitError) else str(exc)
            self.db.update_resource_task(task_id, status="failed", error_message=message[:2000], completed_at=datetime.now(timezone.utc))

    async def _resource_messages(self, entity, keyword: str, limit: int, flooded_api_ids: set[int]):
        offset_id = 0
        yielded = 0
        while yielded < limit:
            try:
                async for message in self.telegram.client.iter_messages(
                    entity, search=keyword, limit=limit - yielded, offset_id=offset_id,
                ):
                    offset_id = message.id
                    yielded += 1
                    yield message
                return
            except errors.FloodWaitError as exc:
                if getattr(self.telegram, "credential_source", "env") != "database":
                    raise
                await self.telegram.switch_after_flood_wait(exc, flooded_api_ids)

    def _match(self, message, task: dict) -> dict | None:
        document = getattr(message, "document", None)
        file = getattr(message, "file", None)
        filename = getattr(file, "name", None)
        if not document or not filename:
            return None
        cleaned = safe_filename(filename)
        extension = cleaned.rsplit(".", 1)[-1].lower() if "." in cleaned else ""
        haystack = f"{cleaned}\n{message.raw_text or ''}".lower()
        if extension not in DEFAULT_TYPES or task["keyword"].lower() not in haystack or task["file_types"] and extension not in set(task["file_types"]):
            return None
        if any(term in haystack for term in BLOCKED_TERMS):
            return None
        size = int(getattr(file, "size", 0) or 0)
        if size <= 0 or size > self.max_file_size:
            return None
        return {"filename": cleaned, "size": size, "mime_type": getattr(file, "mime_type", None), "file_id": str(document.id)}

    async def _download(self, task_id: int, record: dict, message, flooded_api_ids: set[int] | None = None) -> bool:
        task_dir = self.tmp_root / str(task_id)
        task_dir.mkdir(parents=True, exist_ok=True)
        path = task_dir / f"{record['id']}-{safe_filename(record['original_filename'])}"
        self.db.update_resource_file(record["id"], download_status="downloading", attempts=record["attempts"] + 1)
        try:
            try:
                if flooded_api_ids is not None and getattr(self.telegram, "credential_source", "env") == "database":
                    async def download_with_fallback():
                        while True:
                            try:
                                return await self.telegram.client.download_media(message, file=str(path))
                            except errors.FloodWaitError as exc:
                                path.unlink(missing_ok=True)
                                await self.telegram.switch_after_flood_wait(exc, flooded_api_ids)

                    download = asyncio.create_task(download_with_fallback())
                else:
                    download = asyncio.create_task(self.telegram.call(self.telegram.client.download_media, message, file=str(path)))
                self.active_downloads[task_id] = download
                try:
                    async with asyncio.timeout(DOWNLOAD_TIMEOUT_SECONDS):
                        result = await download
                finally:
                    if self.active_downloads.get(task_id) is download:
                        self.active_downloads.pop(task_id)
            except TimeoutError as exc:
                raise RuntimeError("Telegram 下载超时（30 分钟）") from exc
            if self._cancelled(task_id):
                raise asyncio.CancelledError
            if not result or not path.is_file():
                raise RuntimeError("Telegram 下载未生成文件")
            if path.stat().st_size > self.max_file_size:
                raise RuntimeError("文件超过大小限制")
            self._validate_download(path, record["file_type"])
            now = datetime.now(timezone.utc)
            object_key = record.get("object_key") or f"telegram/{now:%Y}/{now:%m}/{task_id}/{record['id']}-{safe_filename(record['original_filename'])}"
            size = path.stat().st_size
            self.db.update_resource_file(record["id"], bytes_downloaded=size, download_status="success", upload_status="uploading", object_key=object_key)
            await asyncio.to_thread(self.storage.upload, path, object_key, record["mime_type"])
            await asyncio.to_thread(
                self._register_telegram_resource,
                {
                    "ownerId": record["owner_id"], "name": record["original_filename"], "size": size,
                    "mimeType": record["mime_type"] or "", "sha256": await asyncio.to_thread(_sha256_file, path),
                    "objectKey": object_key, "sourceName": record["source_name"],
                    "sourceReference": f"telegram://{record['telegram_chat_id']}/{record['telegram_message_id']}",
                    "externalId": f"{record['telegram_chat_id']}:{record['telegram_message_id']}",
                },
            )
            self.db.update_resource_file(record["id"], upload_status="success", object_key=object_key, completed_at=now, error_message=None)
            return True
        except asyncio.CancelledError:
            if not self._cancelled(task_id):
                raise
            name = Path(record["original_filename"])
            self.db.update_resource_file(
                record["id"], original_filename=f"{name.stem}（已中止）{name.suffix}",
                bytes_downloaded=path.stat().st_size if path.is_file() else 0,
                download_status="cancelled", upload_status="cancelled", error_message="任务已取消",
                completed_at=datetime.now(timezone.utc),
            )
            return False

        except Exception as exc:
            self.db.update_resource_file(record["id"], download_status="failed", upload_status="failed", error_message=str(exc)[:2000], completed_at=datetime.now(timezone.utc))
            if isinstance(exc, errors.FloodWaitError):
                raise
            return False
        finally:
            path.unlink(missing_ok=True)
            try:
                task_dir.rmdir()
            except OSError:
                pass

    async def download_and_upload(
        self, telegram_id: int, message_id: int, upload_url: str,
        dialog_type: str = "", username: str | None = None, access_hash: int | None = None,
    ) -> dict:
        if telegram_id == 0 or message_id <= 0 or not upload_url:
            raise ResourceCapabilityError("invalid_argument", "Telegram 消息引用无效")
        if not self.telegram.authorized:
            raise ResourceCapabilityError("authentication_required", "Telegram 尚未登录")
        await self.telegram._ensure_connected()
        entity = await self.telegram.resolve_entity(telegram_id, dialog_type, username, access_hash)
        message = await self.telegram.call(self.telegram.client.get_messages, entity, ids=message_id)
        metadata = self._match(message, {"keyword": "", "file_types": []}) if message else None
        if not metadata or metadata["size"] > self.max_file_size:
            raise ResourceCapabilityError("resource_changed", "Telegram 文件不存在或超过大小限制")
        temp_dir = Path(tempfile.mkdtemp(prefix="telegram-upload-", dir=self.tmp_root))
        path = temp_dir / safe_filename(metadata["filename"])
        try:
            downloaded = await self.telegram.call(self.telegram.client.download_media, message, file=str(path))
            if not downloaded or not path.is_file():
                raise ResourceCapabilityError("download_failed", "Telegram 文件下载失败", retryable=True)
            size = path.stat().st_size
            if size != metadata["size"]:
                raise ResourceCapabilityError("resource_changed", "Telegram 文件大小已发生变化")
            self._validate_download(path, path.suffix.lower().lstrip("."))
            digest = await asyncio.to_thread(_sha256_file, path)
            await asyncio.to_thread(_put_file, upload_url, path, metadata["mime_type"])
            return {
                "name": metadata["filename"], "size": size, "mime_type": metadata["mime_type"],
                "sha256": digest,
            }
        finally:
            shutil.rmtree(temp_dir, ignore_errors=True)

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

    def _register_telegram_resource(self, payload: dict) -> None:
        base_url = os.getenv("XIANYU_INTERNAL_URL", "http://server:8888").rstrip("/")
        request = urllib.request.Request(
            base_url + "/xy/internal/telegram-resources/import",
            data=json.dumps(payload, ensure_ascii=False).encode(),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                result = json.load(response)
        except Exception as exc:
            raise RuntimeError("资源目录登记失败") from exc
        if result.get("code") != 0:
            raise RuntimeError("资源目录拒绝登记 Telegram 文件")

    def _cancelled(self, task_id: int) -> bool:
        task = self.db.resource_task(task_id)
        return not task or task["cancel_requested"]
