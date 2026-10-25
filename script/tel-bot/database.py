from __future__ import annotations

import csv
import io
import sqlite3
from contextlib import contextmanager
from datetime import datetime, timezone
from pathlib import Path


SCHEMA = """
PRAGMA journal_mode=WAL;
CREATE TABLE IF NOT EXISTS dialogs (
 id INTEGER PRIMARY KEY, telegram_id INTEGER NOT NULL UNIQUE, access_hash INTEGER,
 name TEXT NOT NULL, username TEXT, dialog_type TEXT NOT NULL, member_count INTEGER,
 is_channel INTEGER NOT NULL DEFAULT 0, is_group INTEGER NOT NULL DEFAULT 0,
 is_megagroup INTEGER NOT NULL DEFAULT 0, is_broadcast INTEGER NOT NULL DEFAULT 0,
 is_archived INTEGER NOT NULL DEFAULT 0, is_muted INTEGER NOT NULL DEFAULT 0,
 unread_count INTEGER NOT NULL DEFAULT 0, last_message_id INTEGER,
 last_message_at TEXT, last_read_message_id INTEGER, status TEXT NOT NULL DEFAULT 'normal',
 is_creator INTEGER NOT NULL DEFAULT 0,
 created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS scan_history (
 id INTEGER PRIMARY KEY, started_at TEXT NOT NULL, finished_at TEXT, status TEXT NOT NULL,
 dialogs_total INTEGER NOT NULL DEFAULT 0, dialogs_success INTEGER NOT NULL DEFAULT 0,
 dialogs_failed INTEGER NOT NULL DEFAULT 0,
 error_message TEXT
);
CREATE TABLE IF NOT EXISTS leave_history (
 id INTEGER PRIMARY KEY, dialog_id INTEGER, telegram_id INTEGER NOT NULL, name TEXT NOT NULL,
 left_at TEXT NOT NULL, result TEXT NOT NULL, error_message TEXT
);
CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT);
"""

SORTS = {
    "name": "d.name", "type": "d.dialog_type", "members": "d.member_count",
    "last_message": "d.last_message_at", "unread": "d.unread_count", "status": "d.status",
}


class Database:
    def __init__(self, path: Path):
        self.path = path
        path.parent.mkdir(parents=True, exist_ok=True)
        with self.connect() as db:
            db.executescript(SCHEMA)
            columns = {row[1] for row in db.execute("PRAGMA table_info(dialogs)")}
            if "is_creator" not in columns:
                db.execute("ALTER TABLE dialogs ADD COLUMN is_creator INTEGER NOT NULL DEFAULT 0")
            db.execute("UPDATE scan_history SET status='interrupted',finished_at=? WHERE status='running'", (now_iso(),))

    @contextmanager
    def connect(self):
        db = sqlite3.connect(self.path, timeout=30)
        db.row_factory = sqlite3.Row
        db.execute("PRAGMA foreign_keys=ON")
        try:
            yield db
            db.commit()
        finally:
            db.close()

    def setting(self, key: str, value: str | None = None) -> str | None:
        with self.connect() as db:
            if value is not None:
                db.execute("INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", (key, value))
            row = db.execute("SELECT value FROM settings WHERE key=?", (key,)).fetchone()
            return row[0] if row else None

    def start_scan(self) -> int:
        with self.connect() as db:
            return db.execute("INSERT INTO scan_history(started_at,status) VALUES(?, 'running')", (now_iso(),)).lastrowid

    def update_scan(self, scan_id: int, **values) -> None:
        if not values:
            return
        with self.connect() as db:
            db.execute(f"UPDATE scan_history SET {', '.join(f'{key}=?' for key in values)} WHERE id=?", (*values.values(), scan_id))

    def upsert_dialog(self, item: dict) -> int:
        now = now_iso()
        columns = list(item)
        with self.connect() as db:
            db.execute(
                f"INSERT INTO dialogs ({','.join(columns)},created_at,updated_at) VALUES ({','.join('?' for _ in columns)},?,?) "
                f"ON CONFLICT(telegram_id) DO UPDATE SET {','.join(f'{c}=excluded.{c}' for c in columns if c != 'telegram_id')}, updated_at=excluded.updated_at",
                (*item.values(), now, now),
            )
            return db.execute("SELECT id FROM dialogs WHERE telegram_id=?", (item["telegram_id"],)).fetchone()[0]

    def dialog(self, dialog_id: int) -> dict | None:
        with self.connect() as db:
            row = db.execute("SELECT * FROM dialogs WHERE id=?", (dialog_id,)).fetchone()
            return dict(row) if row else None

    def dialog_by_telegram(self, telegram_id: int) -> dict | None:
        with self.connect() as db:
            row = db.execute("SELECT * FROM dialogs WHERE telegram_id=?", (telegram_id,)).fetchone()
            return dict(row) if row else None

    def latest_scan(self) -> dict | None:
        with self.connect() as db:
            row = db.execute("SELECT * FROM scan_history ORDER BY id DESC LIMIT 1").fetchone()
            return dict(row) if row else None

    def summary(self) -> dict:
        with self.connect() as db:
            row = db.execute("""
                SELECT COUNT(*) total, SUM(d.is_group) groups_count, SUM(d.is_broadcast) channels_count,
                 SUM(d.status!='normal') abnormal, SUM(d.is_muted) muted, SUM(d.is_archived) archived,
                 SUM(d.is_creator) created_by_me, SUM(d.unread_count) unread
                FROM dialogs d
                WHERE d.status!='left'
            """).fetchone()
            return {k: (v or 0) for k, v in dict(row).items()}

    def list_dialogs(self, filters: dict, page: int = 1, page_size: int = 30, sort: str = "last_message", order: str = "desc") -> tuple[list[dict], int]:
        where, params = ["d.status!='left'"], []
        if filters.get("q"):
            where.append("d.name LIKE ?"); params.append(f"%{filters['q']}%")
        kind = filters.get("filter")
        if kind == "groups": where.append("d.is_group=1")
        elif kind == "channels": where.append("d.is_broadcast=1")
        elif kind == "normal": where.append("d.status='normal'")
        elif kind == "invalid": where.append("d.status!='normal'")
        elif kind == "muted": where.append("d.is_muted=1")
        elif kind == "archived": where.append("d.is_archived=1")
        elif kind == "creator": where.append("d.is_creator=1")
        for key, column, op in (("members_min", "d.member_count", ">="), ("members_max", "d.member_count", "<=")):
            if filters.get(key) not in (None, ""):
                where.append(f"COALESCE({column},0){op}?"); params.append(int(filters[key]))
        base = " FROM dialogs d WHERE " + " AND ".join(where)
        direction = "ASC" if order == "asc" else "DESC"
        select = "SELECT d.*"
        with self.connect() as db:
            total = db.execute("SELECT COUNT(*)" + base, params).fetchone()[0]
            rows = db.execute(select + base + f" ORDER BY {SORTS.get(sort, SORTS['last_message'])} {direction} NULLS LAST,d.name LIMIT ? OFFSET ?", (*params, page_size, (page - 1) * page_size)).fetchall()
            return [dict(r) for r in rows], total

    def record_leave(self, dialog: dict, result: str, error: str | None = None) -> None:
        with self.connect() as db:
            db.execute("INSERT INTO leave_history(dialog_id,telegram_id,name,left_at,result,error_message) VALUES(?,?,?,?,?,?)", (dialog["id"], dialog["telegram_id"], dialog["name"], now_iso(), result, error))
            if result == "success":
                db.execute("UPDATE dialogs SET status='left',updated_at=? WHERE id=?", (now_iso(), dialog["id"]))

    def export_csv(self, filters: dict) -> str:
        rows, _ = self.list_dialogs(filters, page_size=1_000_000)
        output = io.StringIO()
        fields = ["name", "username", "dialog_type", "member_count", "last_message_at", "unread_count", "is_muted", "is_archived", "is_creator", "status"]
        writer = csv.DictWriter(output, fieldnames=fields, extrasaction="ignore")
        writer.writeheader(); writer.writerows(rows)
        return output.getvalue()


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()
