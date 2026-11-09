from __future__ import annotations

from datetime import datetime, timedelta, timezone


def cleanup_tags(row: dict, now: datetime | None = None) -> list[str]:
    now = now or datetime.now(timezone.utc)
    tags = []
    last = row.get("last_message_at")
    if last:
        last_at = datetime.fromisoformat(last) if isinstance(last, str) else last
        if last_at.tzinfo is None:
            last_at = last_at.replace(tzinfo=timezone.utc)
        age = now - last_at
        if age > timedelta(days=180):
            tags.append("长期无消息")
        elif age > timedelta(days=90):
            tags.append("极低活跃")
        elif age > timedelta(days=30):
            tags.append("低活跃")
    else:
        tags.append("无消息记录")
    if row.get("is_creator"):
        tags.append("我创建")
    if row.get("is_muted"):
        tags.append("已静音")
    if row.get("is_archived"):
        tags.append("已归档")
    if row.get("status") != "normal":
        tags.append("失效" if row.get("status") in {"forbidden", "deleted", "left"} else "不可访问")
    return tags


def telegram_link(row: dict) -> str | None:
    if row.get("username"):
        return f"https://t.me/{row['username'].lstrip('@')}"
    if row.get("is_channel") and row.get("last_message_id"):
        return f"https://t.me/c/{row['telegram_id']}/{row['last_message_id']}"
    return None
