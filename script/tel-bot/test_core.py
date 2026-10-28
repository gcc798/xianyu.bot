import asyncio
from datetime import datetime, timedelta, timezone
import os
from pathlib import Path
from types import SimpleNamespace
import uuid
import zipfile

from analyzer import cleanup_tags, telegram_link
from database import Database
from resource_service import ResourceService
from storage import safe_filename
from telegram_service import TelegramService
from telethon import functions, types


def test_resource_validation():
    assert safe_filename("../../etc/passwd") == "passwd"
    assert safe_filename("..\\..\\bad:name?.pdf") == "bad_name_.pdf"
    service = ResourceService.__new__(ResourceService)
    service.max_file_size = 10
    message = SimpleNamespace(document=SimpleNamespace(id=7), file=SimpleNamespace(name="training.pdf", size=9, mime_type="application/pdf"), raw_text="authorized training")
    task = {"keyword": "training", "file_types": ["pdf"]}
    assert service._match(message, task)["filename"] == "training.pdf"
    message.raw_text = "crack training"
    assert service._match(message, task) is None
    message.raw_text = "authorized training"; message.file.size = 11
    assert service._match(message, task) is None
    archive_path = Path(__file__).resolve().parents[2] / "tmp" / "telegram" / f"{uuid.uuid4()}.zip"
    archive_path.parent.mkdir(parents=True, exist_ok=True)
    try:
        with zipfile.ZipFile(archive_path, "w") as archive:
            archive.writestr("virus.exe", "test")
        try:
            service._validate_download(archive_path, "zip")
            raise AssertionError("dangerous archive was accepted")
        except RuntimeError:
            pass
    finally:
        archive_path.unlink(missing_ok=True)


def test_metadata_scan_storage():
    now = datetime.now(timezone.utc)
    assert "长期无消息" in cleanup_tags({"last_message_at": (now - timedelta(days=200)).isoformat(), "status": "normal"}, now)
    assert "长期无消息" in cleanup_tags({"last_message_at": now - timedelta(days=200), "status": "normal"}, now)
    assert "我创建" in cleanup_tags({"last_message_at": now.isoformat(), "is_creator": True, "status": "normal"}, now)
    assert telegram_link({"username": "test", "is_channel": 1}) == "https://t.me/test"
    assert telegram_link({"username": None, "is_channel": 1, "telegram_id": 123, "last_message_id": 9}) == "https://t.me/c/123/9"
    assert telegram_link({"username": None, "is_channel": 0}) is None
    dsn = os.getenv("DATABASE_URL")
    if not dsn:
        print("PostgreSQL integration checks skipped (DATABASE_URL not set)")
        return
    db = Database(dsn)
    telegram_id = -int(uuid.uuid4().int % 9_000_000_000_000_000_000) - 1
    settings_key = "codex_test_" + str(uuid.uuid4())
    item = {"telegram_id": telegram_id, "access_hash": None, "name": "Codex test " + str(uuid.uuid4()), "username": "codex_test_" + uuid.uuid4().hex, "dialog_type": "普通群", "member_count": 2, "is_channel": False, "is_group": True, "is_megagroup": False, "is_broadcast": False, "is_archived": False, "is_muted": False, "unread_count": 0, "last_message_id": 3, "last_message_at": now, "last_read_message_id": 2, "is_creator": True, "status": "normal"}
    with db.connect() as conn:
        try:
            dialog_id = db.upsert_dialog(item)
            db.upsert_dialog({**item, "last_message_id": 4, "member_count": 3})
            row = db.dialog(dialog_id)
            assert row["last_message_id"] == 4 and row["member_count"] == 3 and row["is_creator"] is True
            rows, total = db.list_dialogs({"q": item["name"]})
            assert total == 1 and rows[0]["telegram_id"] == telegram_id
            db.setting(settings_key, "ok")
            assert db.setting(settings_key) == "ok"
            scan_id = db.start_scan()
            db.update_scan(scan_id, status="completed", finished_at=now)
            db.record_leave(row, "failed", "test")
            task = db.create_resource_task("training", ["pdf"], [dialog_id], 2)
            manager = ResourceService(SimpleNamespace(), db, SimpleNamespace())
            manager.cancel(task["id"])
            assert db.resource_task(task["id"])["status"] == "cancelled"
            manager.retry(task["id"])
            assert db.resource_task(task["id"])["status"] == "queued"
            resource = db.claim_resource_file(task["id"], row, 99, "file-99", "training.pdf", 123, "application/pdf", "authorized training")
            assert resource and db.claim_resource_file(task["id"], row, 99, "file-99", "training.pdf", 123, "application/pdf", "authorized training") is None
            db.update_resource_file(resource["id"], download_status="success", upload_status="success", object_key="telegram/test/training.pdf", completed_at=now)
            assert db.resource_file(resource["id"])["source_name"] == row["name"]
            db.update_resource_task(task["id"], status="cancelled", cancel_requested=True, completed_at=now)
            assert db.resource_task(task["id"])["status"] == "cancelled"
        finally:
            conn.execute("DELETE FROM tel_resource_tasks WHERE id=%s", (locals().get("task", {}).get("id", -1),))
            conn.execute("DELETE FROM tel_leave_history WHERE telegram_id=%s", (telegram_id,))
            conn.execute("DELETE FROM tel_dialogs WHERE telegram_id=%s", (telegram_id,))
            conn.execute("DELETE FROM tel_scan_history WHERE id=%s", (locals().get("scan_id", -1),))
            conn.execute("DELETE FROM tel_settings WHERE key=%s", (settings_key,))


def test_qr_login_migrates_before_rendering():
    imported = types.auth.LoginToken(datetime.now(timezone.utc) + timedelta(seconds=30), b"new-token")

    class Client:
        def __init__(self):
            self.dc_id = None

        async def _switch_dc(self, dc_id):
            self.dc_id = dc_id

        async def __call__(self, request):
            assert isinstance(request, functions.auth.ImportLoginTokenRequest)
            assert request.token == b"migrate-token"
            return imported

    service = TelegramService.__new__(TelegramService)
    service.client = Client()
    login = SimpleNamespace(_resp=types.auth.LoginTokenMigrateTo(2, b"migrate-token"))

    assert asyncio.run(service._prepare_qr(login)) is True
    assert service.client.dc_id == 2
    assert login._resp is imported


if __name__ == "__main__":
    test_resource_validation()
    test_metadata_scan_storage()
    test_qr_login_migrates_before_rendering()
    print("ok")
