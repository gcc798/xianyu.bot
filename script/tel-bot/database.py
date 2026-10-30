from __future__ import annotations

import csv
import io
import json
from contextlib import contextmanager
from datetime import datetime, timezone

import psycopg
from psycopg.rows import dict_row

TABLES = {"dialogs": "tel_dialogs", "scan_history": "tel_scan_history", "leave_history": "tel_leave_history", "settings": "tel_settings"}
SORTS = {"name": "d.name", "type": "d.dialog_type", "members": "d.member_count", "last_message": "d.last_message_at", "unread": "d.unread_count", "status": "d.status"}
ADULT_NAME_PATTERN = (
    r"成人|色情|情色|成人视频|裸聊|裸舞|裸播|裸照|约炮|炮友|一夜情|援交|福利姬|福利视频|福利资源|"
    r"熟女|少妇|人妻|萝莉|乱伦|母狗|调教|换妻|群交|迷奸|强奸|自拍偷拍|国产自拍|鲍鱼资源|"
    r"草榴|杏吧|色花堂|海角社区|stripchat|onlyfans|pornhub|xvideos|xhamster|redtube|"
    r"18禁|无码|有码|av女优|\m(?:porn|xxx|nsfw|jav)\M"
)


class Database:
    def __init__(self, dsn: str):
        self.dsn = dsn
        with self.connect() as db:
            db.execute(f"""
                CREATE TABLE IF NOT EXISTS {TABLES['dialogs']} (
                    id BIGSERIAL PRIMARY KEY, telegram_id BIGINT NOT NULL UNIQUE, access_hash BIGINT,
                    name TEXT NOT NULL, username TEXT, dialog_type TEXT NOT NULL, member_count INTEGER,
                    is_channel BOOLEAN NOT NULL DEFAULT FALSE, is_group BOOLEAN NOT NULL DEFAULT FALSE,
                    is_megagroup BOOLEAN NOT NULL DEFAULT FALSE, is_broadcast BOOLEAN NOT NULL DEFAULT FALSE,
                    is_archived BOOLEAN NOT NULL DEFAULT FALSE, is_muted BOOLEAN NOT NULL DEFAULT FALSE,
                    unread_count INTEGER NOT NULL DEFAULT 0, last_message_id BIGINT, last_message_at TIMESTAMPTZ,
                    last_read_message_id BIGINT, status TEXT NOT NULL DEFAULT 'normal', is_creator BOOLEAN NOT NULL DEFAULT FALSE,
                    created_at TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL
                )
            """)
            db.execute(f"""
                CREATE TABLE IF NOT EXISTS {TABLES['scan_history']} (
                    id BIGSERIAL PRIMARY KEY, started_at TIMESTAMPTZ NOT NULL, finished_at TIMESTAMPTZ,
                    status TEXT NOT NULL, dialogs_total INTEGER NOT NULL DEFAULT 0, dialogs_success INTEGER NOT NULL DEFAULT 0,
                    dialogs_failed INTEGER NOT NULL DEFAULT 0, error_message TEXT
                )
            """)
            db.execute(f"""
                CREATE TABLE IF NOT EXISTS {TABLES['leave_history']} (
                    id BIGSERIAL PRIMARY KEY, dialog_id BIGINT, telegram_id BIGINT NOT NULL, name TEXT NOT NULL,
                    left_at TIMESTAMPTZ NOT NULL, result TEXT NOT NULL, error_message TEXT
                )
            """)
            db.execute(f"CREATE TABLE IF NOT EXISTS {TABLES['settings']} (key TEXT PRIMARY KEY, value TEXT)")
            db.execute(f"INSERT INTO {TABLES['settings']}(key,value) VALUES('adult_mode','false'),('adult_dialog_ids','[]') ON CONFLICT(key) DO NOTHING")
            db.execute("ALTER TABLE tel_dialogs ADD COLUMN IF NOT EXISTS resource_compliance_status TEXT NOT NULL DEFAULT 'allowed'")
            db.execute("""
                CREATE TABLE IF NOT EXISTS tel_resource_tasks (
                    id BIGSERIAL PRIMARY KEY,
                    keyword TEXT NOT NULL,
                    file_types TEXT[] NOT NULL DEFAULT '{}',
                    source_dialog_ids BIGINT[] NOT NULL DEFAULT '{}',
                    max_downloads INTEGER NOT NULL,
                    status TEXT NOT NULL DEFAULT 'queued',
                    files_found INTEGER NOT NULL DEFAULT 0,
                    files_completed INTEGER NOT NULL DEFAULT 0,
                    error_message TEXT,
                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    started_at TIMESTAMPTZ,
                    completed_at TIMESTAMPTZ,
                    cancel_requested BOOLEAN NOT NULL DEFAULT FALSE,
                    CHECK (max_downloads BETWEEN 1 AND 100),
                    CHECK (status IN ('queued','running','completed','failed','cancelled'))
                )
            """)
            db.execute("""
                CREATE TABLE IF NOT EXISTS tel_resource_files (
                    id BIGSERIAL PRIMARY KEY,
                    task_id BIGINT NOT NULL REFERENCES tel_resource_tasks(id) ON DELETE CASCADE,
                    telegram_chat_id BIGINT NOT NULL,
                    telegram_message_id BIGINT NOT NULL,
                    telegram_file_id TEXT,
                    original_filename TEXT NOT NULL,
                    file_type TEXT NOT NULL,
                    file_size BIGINT NOT NULL DEFAULT 0,
                    mime_type TEXT,
                    object_key TEXT,
                    download_status TEXT NOT NULL DEFAULT 'queued',
                    upload_status TEXT NOT NULL DEFAULT 'pending',
                    bytes_downloaded BIGINT NOT NULL DEFAULT 0,
                    attempts INTEGER NOT NULL DEFAULT 0,
                    error_message TEXT,
                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    completed_at TIMESTAMPTZ,
                    UNIQUE (telegram_chat_id, telegram_message_id),
                    CHECK (download_status IN ('queued','downloading','success','failed','cancelled','duplicate')),
                    CHECK (upload_status IN ('pending','uploading','success','failed','cancelled'))
                )
            """)
            db.execute("""
                CREATE TABLE IF NOT EXISTS tel_resource_sources (
                    id BIGSERIAL PRIMARY KEY,
                    file_id BIGINT NOT NULL REFERENCES tel_resource_files(id) ON DELETE CASCADE,
                    dialog_id BIGINT REFERENCES tel_dialogs(id) ON DELETE SET NULL,
                    telegram_chat_id BIGINT NOT NULL,
                    telegram_message_id BIGINT NOT NULL,
                    source_name TEXT NOT NULL,
                    source_username TEXT,
                    message_text TEXT,
                    compliance_status TEXT NOT NULL DEFAULT 'allowed',
                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    UNIQUE (file_id, telegram_chat_id, telegram_message_id)
                )
            """)
            db.execute("CREATE INDEX IF NOT EXISTS idx_tel_resource_files_task ON tel_resource_files(task_id)")
            db.execute("CREATE INDEX IF NOT EXISTS idx_tel_resource_files_filter ON tel_resource_files(download_status,file_type)")
            db.execute("CREATE UNIQUE INDEX IF NOT EXISTS idx_tel_resource_files_telegram_file ON tel_resource_files(telegram_file_id) WHERE telegram_file_id IS NOT NULL")
            db.execute("ALTER TABLE tel_resource_tasks ADD COLUMN IF NOT EXISTS source_mode TEXT NOT NULL DEFAULT 'manual'")
            db.execute("ALTER TABLE tel_resource_tasks ADD COLUMN IF NOT EXISTS source_strategy TEXT NOT NULL DEFAULT 'manual'")
            db.execute("""
                CREATE TABLE IF NOT EXISTS tel_acquisition_tasks (
                    id BIGSERIAL PRIMARY KEY,
                    gva_task_id BIGINT NOT NULL UNIQUE,
                    candidate_id TEXT NOT NULL,
                    idempotency_key TEXT NOT NULL UNIQUE,
                    state TEXT NOT NULL DEFAULT 'queued',
                    stage TEXT,
                    bytes_completed BIGINT NOT NULL DEFAULT 0,
                    bytes_total BIGINT NOT NULL DEFAULT 0,
                    result_json TEXT,
                    error_code TEXT,
                    error_message TEXT,
                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    started_at TIMESTAMPTZ,
                    completed_at TIMESTAMPTZ,
                    CHECK (state IN ('queued','running','completed','failed','cancelled'))
                )
            """)
            db.execute(f"UPDATE {TABLES['scan_history']} SET status='interrupted', finished_at=%s WHERE status='running'", (now_iso(),))
            db.execute("UPDATE tel_resource_tasks SET status='failed',completed_at=NOW(),error_message='service restarted' WHERE status='running'")
            db.execute("UPDATE tel_acquisition_tasks SET state='failed',completed_at=NOW(),error_code='internal_error',error_message='service restarted' WHERE state IN ('queued','running')")

    @contextmanager
    def connect(self):
        with psycopg.connect(self.dsn, row_factory=dict_row) as db:
            yield db

    def setting(self, key: str, value: str | None = None) -> str | None:
        with self.connect() as db:
            if value is not None:
                db.execute(f"INSERT INTO {TABLES['settings']}(key,value) VALUES(%s,%s) ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value", (key, value))
            row = db.execute(f"SELECT value FROM {TABLES['settings']} WHERE key=%s", (key,)).fetchone()
            return row["value"] if row else None

    def health(self) -> bool:
        with self.connect() as db:
            return db.execute("SELECT 1 AS ok").fetchone()["ok"] == 1

    def adult_mode(self) -> bool:
        return self.setting("adult_mode") == "true"

    def set_adult_mode(self, enabled: bool) -> None:
        self.setting("adult_mode", "true" if enabled else "false")

    def adult_dialog_ids(self) -> list[int]:
        try:
            values = json.loads(self.setting("adult_dialog_ids") or "[]")
            if not isinstance(values, list):
                return []
            return list(dict.fromkeys(int(value) for value in values))
        except (TypeError, ValueError, json.JSONDecodeError):
            return []

    def save_adult_dialog_ids(self, dialog_ids: list[int]) -> None:
        self.setting("adult_dialog_ids", json.dumps(sorted(set(dialog_ids))))

    def _adult_visibility(self, alias: str = "d") -> tuple[str | None, list]:
        if self.adult_mode():
            return None, []
        return f"NOT ({alias}.id=ANY(%s) OR LOWER({alias}.name) ~ %s OR LOWER(COALESCE({alias}.username,'')) ~ %s)", [self.adult_dialog_ids(), ADULT_NAME_PATTERN, ADULT_NAME_PATTERN]

    def start_scan(self) -> int:
        with self.connect() as db:
            return db.execute(f"INSERT INTO {TABLES['scan_history']}(started_at,status) VALUES(%s,'running') RETURNING id", (now_iso(),)).fetchone()["id"]

    def update_scan(self, scan_id: int, **values) -> None:
        if values:
            with self.connect() as db:
                assignments = ", ".join(f"{key}=%s" for key in values)
                db.execute(f"UPDATE {TABLES['scan_history']} SET {assignments} WHERE id=%s", (*values.values(), scan_id))

    def upsert_dialog(self, item: dict) -> int:
        now = now_iso(); columns = list(item)
        with self.connect() as db:
            names = ",".join(columns); placeholders = ",".join(["%s"] * len(columns))
            updates = ",".join(f"{c}=EXCLUDED.{c}" for c in columns if c != "telegram_id")
            return db.execute(f"INSERT INTO {TABLES['dialogs']} ({names},created_at,updated_at) VALUES ({placeholders},%s,%s) ON CONFLICT(telegram_id) DO UPDATE SET {updates},updated_at=EXCLUDED.updated_at RETURNING id", (*item.values(), now, now)).fetchone()["id"]

    def dialog(self, dialog_id: int) -> dict | None:
        with self.connect() as db:
            return db.execute(f"SELECT * FROM {TABLES['dialogs']} WHERE id=%s", (dialog_id,)).fetchone()

    def latest_scan(self) -> dict | None:
        with self.connect() as db:
            return db.execute(f"SELECT * FROM {TABLES['scan_history']} ORDER BY id DESC LIMIT 1").fetchone()

    def scan_history(self, scan_id: int) -> dict | None:
        with self.connect() as db:
            return db.execute(f"SELECT * FROM {TABLES['scan_history']} WHERE id=%s", (scan_id,)).fetchone()

    def summary(self) -> dict:
        visibility, params = self._adult_visibility()
        where = "d.status!='left'" + (f" AND {visibility}" if visibility else "")
        with self.connect() as db:
            row = db.execute(f"SELECT COUNT(*) AS total, COALESCE(SUM(is_group::int),0) AS groups_count, COALESCE(SUM(is_broadcast::int),0) AS channels_count, COALESCE(SUM((status!='normal')::int),0) AS abnormal, COALESCE(SUM(is_muted::int),0) AS muted, COALESCE(SUM(is_archived::int),0) AS archived, COALESCE(SUM(is_creator::int),0) AS created_by_me, COALESCE(SUM(unread_count),0) AS unread FROM {TABLES['dialogs']} d WHERE {where}", params).fetchone()
            return dict(row)

    def list_dialogs(self, filters: dict, page: int = 1, page_size: int = 30, sort: str = "last_message", order: str = "desc") -> tuple[list[dict], int]:
        where, params = ["d.status!='left'"], []
        visibility, visibility_params = self._adult_visibility()
        if visibility:
            where.append(visibility); params.extend(visibility_params)
        if filters.get("q"): where.append("d.name ILIKE %s"); params.append(f"%{filters['q']}%")
        kind = filters.get("filter")
        if kind == "groups": where.append("d.is_group")
        elif kind == "channels": where.append("d.is_broadcast")
        elif kind == "normal": where.append("d.status='normal'")
        elif kind == "invalid": where.append("d.status!='normal'")
        elif kind == "muted": where.append("d.is_muted")
        elif kind == "archived": where.append("d.is_archived")
        elif kind == "creator": where.append("d.is_creator")
        for key, column, op in (("members_min", "d.member_count", ">="), ("members_max", "d.member_count", "<=")):
            if filters.get(key) not in (None, ""): where.append(f"COALESCE({column},0){op}%s"); params.append(int(filters[key]))
        base = " FROM " + TABLES["dialogs"] + " d WHERE " + " AND ".join(where); direction = "ASC" if order == "asc" else "DESC"
        with self.connect() as db:
            total = db.execute("SELECT COUNT(*)" + base, params).fetchone()["count"]
            rows = db.execute("SELECT d.*" + base + f" ORDER BY {SORTS.get(sort, SORTS['last_message'])} {direction} NULLS LAST,d.name LIMIT %s OFFSET %s", (*params, page_size, (page - 1) * page_size)).fetchall()
            return [dict(r) for r in rows], total

    def record_leave(self, dialog: dict, result: str, error: str | None = None) -> None:
        with self.connect() as db:
            db.execute(f"INSERT INTO {TABLES['leave_history']}(dialog_id,telegram_id,name,left_at,result,error_message) VALUES(%s,%s,%s,%s,%s,%s)", (dialog["id"], dialog["telegram_id"], dialog["name"], now_iso(), result, error))
            if result == "success": db.execute(f"UPDATE {TABLES['dialogs']} SET status='left',updated_at=%s WHERE id=%s", (now_iso(), dialog["id"]))

    def export_csv(self, filters: dict) -> str:
        rows, _ = self.list_dialogs(filters, page_size=1_000_000); output = io.StringIO()
        fields = ["name", "username", "dialog_type", "member_count", "last_message_at", "unread_count", "is_muted", "is_archived", "is_creator", "status"]
        writer = csv.DictWriter(output, fieldnames=fields, extrasaction="ignore"); writer.writeheader(); writer.writerows(rows)
        return output.getvalue()

    def resource_dialogs(self) -> list[dict]:
        visibility, params = self._adult_visibility()
        adult_filter = f" AND {visibility}" if visibility else ""
        with self.connect() as db:
            return db.execute(f"SELECT d.id,d.telegram_id,d.name,d.username,d.dialog_type,d.is_group,d.is_broadcast,d.member_count,d.last_message_at FROM tel_dialogs d WHERE d.status='normal' AND d.resource_compliance_status='allowed' AND d.username IS NOT NULL AND (d.is_group OR d.is_broadcast){adult_filter} ORDER BY d.last_message_at DESC NULLS LAST,d.name", params).fetchall()

    def adult_classification_dialogs(self) -> list[dict]:
        with self.connect() as db:
            return db.execute("SELECT id,name,username,dialog_type FROM tel_dialogs WHERE status!='left' ORDER BY id").fetchall()

    def create_resource_task(self, keyword: str, file_types: list[str], source_ids: list[int], max_downloads: int, source_mode: str = "manual", source_strategy: str = "manual") -> dict:
        visibility, visibility_params = self._adult_visibility()
        adult_filter = f" AND {visibility}" if visibility else ""
        with self.connect() as db:
            allowed = db.execute(f"SELECT d.id FROM tel_dialogs d WHERE d.id=ANY(%s) AND d.status='normal' AND d.resource_compliance_status='allowed' AND d.username IS NOT NULL AND (d.is_group OR d.is_broadcast){adult_filter}", (source_ids, *visibility_params)).fetchall()
            if len(allowed) != len(set(source_ids)):
                raise ValueError("搜索范围包含未扫描、非公开或状态异常的群组/频道")
            return db.execute("INSERT INTO tel_resource_tasks(keyword,file_types,source_dialog_ids,max_downloads,source_mode,source_strategy) VALUES(%s,%s,%s,%s,%s,%s) RETURNING *", (keyword, file_types, source_ids, max_downloads, source_mode, source_strategy)).fetchone()

    def resource_task(self, task_id: int) -> dict | None:
        visibility, visibility_params = self._adult_visibility()
        adult_filter = f" AND {visibility}" if visibility else ""
        with self.connect() as db:
            row = db.execute("SELECT * FROM tel_resource_tasks WHERE id=%s", (task_id,)).fetchone()
            if row:
                row["sources"] = db.execute(f"SELECT d.id,d.telegram_id,d.name,d.username,d.dialog_type,d.status,d.resource_compliance_status FROM tel_dialogs d WHERE d.id=ANY(%s){adult_filter} ORDER BY array_position(%s::bigint[],d.id)", (row["source_dialog_ids"], *visibility_params, row["source_dialog_ids"])).fetchall()
            return row

    def list_resource_tasks(self, page: int = 1, page_size: int = 20) -> tuple[list[dict], int]:
        with self.connect() as db:
            total = db.execute("SELECT COUNT(*) FROM tel_resource_tasks").fetchone()["count"]
            rows = db.execute("SELECT * FROM tel_resource_tasks ORDER BY id DESC LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size)).fetchall()
            return rows, total

    def queued_resource_tasks(self) -> list[int]:
        with self.connect() as db:
            return [row["id"] for row in db.execute("SELECT id FROM tel_resource_tasks WHERE status='queued' AND NOT cancel_requested ORDER BY id").fetchall()]

    def create_acquisition_task(self, gva_task_id: int, candidate_id: str, idempotency_key: str) -> tuple[dict, bool]:
        with self.connect() as db:
            row = db.execute(
                "INSERT INTO tel_acquisition_tasks(gva_task_id,candidate_id,idempotency_key) VALUES(%s,%s,%s) ON CONFLICT(idempotency_key) DO NOTHING RETURNING *",
                (gva_task_id, candidate_id, idempotency_key),
            ).fetchone()
            if row:
                return row, True
            existing = db.execute("SELECT * FROM tel_acquisition_tasks WHERE idempotency_key=%s", (idempotency_key,)).fetchone()
            if not existing or existing["gva_task_id"] != gva_task_id or existing["candidate_id"] != candidate_id:
                raise ValueError("幂等键已用于其他资源任务")
            if existing["state"] in {"failed", "cancelled"}:
                existing = db.execute(
                    "UPDATE tel_acquisition_tasks SET state='queued',stage=NULL,bytes_completed=0,bytes_total=0,result_json=NULL,error_code=NULL,error_message=NULL,started_at=NULL,completed_at=NULL WHERE id=%s RETURNING *",
                    (existing["id"],),
                ).fetchone()
                return existing, True
            return existing, False

    def acquisition_task(self, task_id: int) -> dict | None:
        with self.connect() as db:
            return db.execute("SELECT * FROM tel_acquisition_tasks WHERE id=%s", (task_id,)).fetchone()

    def update_acquisition_task(self, task_id: int, **values) -> None:
        if values:
            with self.connect() as db:
                assignments = ", ".join(f"{key}=%s" for key in values)
                db.execute(f"UPDATE tel_acquisition_tasks SET {assignments} WHERE id=%s", (*values.values(), task_id))

    def update_resource_task(self, task_id: int, **values) -> None:
        allowed = {"status", "files_found", "files_completed", "error_message", "started_at", "completed_at", "cancel_requested"}
        if not values or not set(values) <= allowed:
            raise ValueError("invalid task update")
        with self.connect() as db:
            assignments = ",".join(f"{key}=%s" for key in values)
            db.execute(f"UPDATE tel_resource_tasks SET {assignments} WHERE id=%s", (*values.values(), task_id))

    def prepare_resource_retry(self, task_id: int) -> tuple[int, int]:
        with self.connect() as db:
            db.execute("DELETE FROM tel_resource_files WHERE task_id=%s AND (download_status='failed' OR upload_status='failed')", (task_id,))
            row = db.execute("SELECT COUNT(*) AS found,COUNT(*) FILTER (WHERE upload_status='success') AS completed FROM tel_resource_files WHERE task_id=%s", (task_id,)).fetchone()
            return row["found"], row["completed"]

    def claim_resource_file(self, task_id: int, dialog: dict, message_id: int, file_id: str | None, filename: str, file_size: int, mime_type: str | None, message_text: str) -> dict | None:
        extension = filename.rsplit(".", 1)[-1].lower() if "." in filename else ""
        with self.connect() as db:
            row = db.execute("""
                INSERT INTO tel_resource_files(task_id,telegram_chat_id,telegram_message_id,telegram_file_id,original_filename,file_type,file_size,mime_type)
                VALUES(%s,%s,%s,%s,%s,%s,%s,%s) ON CONFLICT(telegram_chat_id,telegram_message_id) DO NOTHING RETURNING *
            """, (task_id, dialog["telegram_id"], message_id, file_id, filename, extension, file_size, mime_type)).fetchone()
            if not row:
                return None
            db.execute("""
                INSERT INTO tel_resource_sources(file_id,dialog_id,telegram_chat_id,telegram_message_id,source_name,source_username,message_text)
                VALUES(%s,%s,%s,%s,%s,%s,%s)
            """, (row["id"], dialog["id"], dialog["telegram_id"], message_id, dialog["name"], dialog["username"], message_text[:4000]))
            return row

    def update_resource_file(self, file_id: int, **values) -> None:
        allowed = {"download_status", "upload_status", "bytes_downloaded", "attempts", "error_message", "object_key", "completed_at"}
        if not values or not set(values) <= allowed:
            raise ValueError("invalid file update")
        with self.connect() as db:
            assignments = ",".join(f"{key}=%s" for key in values)
            db.execute(f"UPDATE tel_resource_files SET {assignments} WHERE id=%s", (*values.values(), file_id))

    def resource_file(self, file_id: int) -> dict | None:
        with self.connect() as db:
            return db.execute("SELECT f.*,s.source_name,s.source_username FROM tel_resource_files f JOIN tel_resource_sources s ON s.file_id=f.id WHERE f.id=%s", (file_id,)).fetchone()

    def list_resource_files(self, filters: dict, page: int = 1, page_size: int = 20) -> tuple[list[dict], int]:
        where, params = ["TRUE"], []
        for key, column in (("task_id", "f.task_id"), ("file_type", "f.file_type"), ("status", "f.download_status"), ("source_id", "s.dialog_id")):
            if filters.get(key) not in (None, ""):
                where.append(f"{column}=%s"); params.append(filters[key])
        if filters.get("filename"):
            where.append("f.original_filename ILIKE %s"); params.append(f"%{filters['filename']}%")
        base = " FROM tel_resource_files f JOIN tel_resource_sources s ON s.file_id=f.id WHERE " + " AND ".join(where)
        with self.connect() as db:
            total = db.execute("SELECT COUNT(*)" + base, params).fetchone()["count"]
            rows = db.execute("SELECT f.*,s.source_name,s.source_username" + base + " ORDER BY f.id DESC LIMIT %s OFFSET %s", (*params, page_size, (page - 1) * page_size)).fetchall()
            return rows, total


def now_iso() -> datetime:
    return datetime.now(timezone.utc)
