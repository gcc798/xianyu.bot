import tempfile
from datetime import datetime, timedelta, timezone
from pathlib import Path

from analyzer import cleanup_tags, telegram_link
from database import Database


def test_metadata_scan_storage():
    now = datetime.now(timezone.utc)
    assert "长期无消息" in cleanup_tags({"last_message_at": (now - timedelta(days=200)).isoformat(), "status": "normal"}, now)
    assert "我创建" in cleanup_tags({"last_message_at": now.isoformat(), "is_creator": 1, "status": "normal"}, now)
    assert telegram_link({"username": "test", "is_channel": 1}) == "https://t.me/test"
    assert telegram_link({"username": None, "is_channel": 1, "telegram_id": 123, "last_message_id": 9}) == "https://t.me/c/123/9"
    assert telegram_link({"username": None, "is_channel": 0}) is None
    with tempfile.TemporaryDirectory() as tmp:
        db = Database(Path(tmp) / "test.db")
        item = {"telegram_id": 1, "access_hash": None, "name": "Test", "username": None, "dialog_type": "普通群", "member_count": 2, "is_channel": 0, "is_group": 1, "is_megagroup": 0, "is_broadcast": 0, "is_archived": 0, "is_muted": 0, "unread_count": 0, "last_message_id": 3, "last_message_at": now.isoformat(), "last_read_message_id": 2, "is_creator": 1, "status": "normal"}
        dialog_id = db.upsert_dialog(item)
        db.upsert_dialog({**item, "last_message_id": 4, "member_count": 3})
        row = db.dialog(dialog_id)
        assert row["last_message_id"] == 4 and row["member_count"] == 3 and row["is_creator"] == 1
        rows, total = db.list_dialogs({})
        assert total == 1 and rows[0]["name"] == "Test"
        rows, total = db.list_dialogs({"q": "es"})
        assert total == 1 and rows[0]["name"] == "Test"


if __name__ == "__main__":
    test_metadata_scan_storage()
    print("ok")
