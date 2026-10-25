from __future__ import annotations

import asyncio
import os
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
        if task["keyword"].lower() not in haystack or extension not in set(task["file_types"]):
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
