import asyncio
from datetime import datetime, timedelta, timezone
from io import BytesIO
import json
import os
from pathlib import Path
from tempfile import TemporaryDirectory
from types import SimpleNamespace
import uuid
import zipfile
from unittest.mock import AsyncMock, patch

from analyzer import cleanup_tags, telegram_link
from database import ADULT_NAME_PATTERN, Database
import resource_service
from resource_service import ResourceService, rank_sources
from storage import safe_filename
from telegram_service import TelegramService
from telethon import errors, functions, types


def test_resource_validation():
    assert resource_service.FILE_TYPE_DICTIONARY["mp4"] == {"label": "MP4 视频", "category": "video", "preview": "video"}
    assert resource_service.FILE_TYPE_DICTIONARY["docx"]["preview"] is None
    assert safe_filename("../../etc/passwd") == "passwd"
    assert safe_filename("..\\..\\bad:name?.pdf") == "bad_name_.pdf"
    service = ResourceService.__new__(ResourceService)
    service.max_file_size = 10
    message = SimpleNamespace(document=SimpleNamespace(id=7), file=SimpleNamespace(name="training.pdf", size=9, mime_type="application/pdf"), raw_text="authorized training")
    task = {"keyword": "training", "file_types": ["pdf"]}
    assert service._match(message, task)["filename"] == "training.pdf"
    task["file_types"] = []
    assert service._match(message, task)["filename"] == "training.pdf"
    message.file.name = "training.mp4"
    message.file.mime_type = "video/mp4"
    assert service._match(message, task)["filename"] == "training.mp4"
    message.file.name = "training.exe"
    assert service._match(message, task) is None
    message.file.name = "training.pdf"
    message.file.mime_type = "application/pdf"
    ranked = rank_sources("Kubernetes 指南", [{"id": 1, "name": "日常闲聊", "username": "chat"}, {"id": 2, "name": "Kubernetes 资料", "username": "k8s"}])
    assert [item["id"] for item in ranked] == [2, 1]
    message.raw_text = "crack training"
    assert service._match(message, task) is None
    message.raw_text = "authorized training"; message.file.size = 11
    assert service._match(message, task) is None
    archive_path = Path(os.getenv("TELEGRAM_TMP_ROOT", Path(__file__).resolve().parent / "tmp")) / f"{uuid.uuid4()}.zip"
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


def test_resource_download_timeout_records_failure():
    class FakeDatabase:
        updates = []

        def update_resource_file(self, _file_id, **values):
            self.updates.append(values)

    class FakeTelegram:
        async def download_media(self, *_args, **_kwargs):
            await asyncio.sleep(1)

        async def call(self, func, *args, **kwargs):
            return await func(*args, **kwargs)

    with TemporaryDirectory() as temp_dir, patch("resource_service.DOWNLOAD_TIMEOUT_SECONDS", 0.01):
        service = ResourceService.__new__(ResourceService)
        service.db = FakeDatabase()
        service.telegram = FakeTelegram()
        service.telegram.client = service.telegram
        service.storage = SimpleNamespace()
        service.tmp_root = Path(temp_dir)
        service.max_file_size = 1024
        service.active_downloads = {}
        record = {"id": 1, "attempts": 0, "original_filename": "file.pdf", "file_type": "pdf", "mime_type": "application/pdf"}
        assert asyncio.run(service._download(1, record, SimpleNamespace())) is False
        assert service.db.updates[-1]["download_status"] == "failed"
        assert "下载超时" in service.db.updates[-1]["error_message"]


def test_resource_cancel_interrupts_active_download():
    async def scenario():
        started = asyncio.Event()

        class FakeDatabase:
            task = {"id": 9, "status": "running", "cancel_requested": False}
            file_updates = []

            def resource_task(self, _task_id):
                return self.task

            def update_resource_task(self, _task_id, **values):
                self.task.update(values)

            def update_resource_file(self, _file_id, **values):
                self.file_updates.append(values)

        class FakeTelegram:
            client = None

            def __init__(self):
                self.client = self

            async def download_media(self, _message, file):
                Path(file).write_bytes(b"partial")
                started.set()
                await asyncio.Event().wait()

            async def call(self, func, *args, **kwargs):
                return await func(*args, **kwargs)

        with TemporaryDirectory() as temp_dir:
            service = ResourceService.__new__(ResourceService)
            service.db = FakeDatabase()
            service.telegram = FakeTelegram()
            service.storage = SimpleNamespace()
            service.tmp_root = Path(temp_dir)
            service.max_file_size = 1024
            service.active_downloads = {}
            record = {"id": 3, "attempts": 0, "original_filename": "video.mp4", "file_type": "mp4", "mime_type": "video/mp4"}
            running = asyncio.create_task(service._download(9, record, SimpleNamespace()))
            await started.wait()
            service.cancel(9, 11)
            assert await asyncio.wait_for(running, 0.2) is False
            assert service.db.task["status"] == "cancelled"
            assert service.db.file_updates[-1]["original_filename"] == "video（已中止）.mp4"
            assert service.db.file_updates[-1]["bytes_downloaded"] == 7
            assert service.db.file_updates[-1]["download_status"] == "cancelled"
            assert service.db.file_updates[-1]["upload_status"] == "cancelled"

    asyncio.run(scenario())


def test_resource_deletion_keeps_or_removes_related_files():
    class FakeDatabase:
        task = {"id": 9, "status": "completed"}
        files = [{"id": 3, "object_key": "telegram/3.pdf"}, {"id": 4, "object_key": None}]
        deleted_tasks = []

        def resource_task(self, _task_id, _owner_id=None):
            return self.task

        def resource_task_files(self, _task_id, _owner_id=None):
            return self.files

        def delete_resource_task(self, task_id, delete_files, _owner_id):
            self.deleted_tasks.append((task_id, delete_files))

        def resource_file(self, _file_id, _owner_id=None):
            return {"id": 3, "object_key": "telegram/3.pdf", "download_status": "success", "upload_status": "success"}

    service = ResourceService.__new__(ResourceService)
    service.db = FakeDatabase()
    deleted_files = []
    service._delete_telegram_resource = lambda file_id, owner_id: deleted_files.append((file_id, owner_id))

    service.delete_task(9, False, 11)
    assert service.db.deleted_tasks == [(9, False)]
    assert deleted_files == []

    service.delete_task(9, True, 11)
    assert service.db.deleted_tasks[-1] == (9, True)
    assert deleted_files == [(3, 11), (4, 11)]

    service.delete_file(3, 11)
    assert deleted_files[-1] == (3, 11)

    service.db.task = {"id": 10, "status": "running"}
    try:
        service.delete_task(10, True, 11)
        raise AssertionError("running task was deleted")
    except ValueError as exc:
        assert "先取消任务" in str(exc)


def test_telegram_download_registers_resource_with_go_owner():
    service = ResourceService.__new__(ResourceService)
    with patch("resource_service.urllib.request.urlopen") as urlopen:
        urlopen.return_value.__enter__.return_value = BytesIO(b'{"code":0,"data":{"resource":{}}}')
        service._register_telegram_resource({"ownerId": 11, "externalId": "1:2"})

    request = urlopen.call_args.args[0]
    assert request.full_url.endswith("/xy/internal/telegram-resources/import")
    assert request.get_header("X-telegram-internal-token") is None
    assert json.loads(request.data) == {"ownerId": 11, "externalId": "1:2"}


def test_adult_dialog_filter_uses_cached_ids_without_ai():
    settings = {"adult_mode": "false", "adult_dialog_ids": '[7, "8"]'}
    database = Database.__new__(Database)
    database.setting = lambda key, value=None: settings.get(key) if value is None else settings.setdefault(key, value)
    clause, params = database._adult_visibility()
    assert clause and params[0] == [7, 8] and "LOWER(d.name)" in clause
    assert all(term in ADULT_NAME_PATTERN for term in ("熟女", "stripchat", "裸舞", "乱伦"))

    class FakeDatabase:
        saved = None

        def setting(self, _key):
            return None

        def adult_classification_dialogs(self):
            return [{"id": 7, "name": "test", "username": "test", "dialog_type": "频道"}, {"id": 8, "name": "safe", "username": "safe", "dialog_type": "频道"}]

        def save_adult_dialog_ids(self, dialog_ids):
            self.saved = dialog_ids

    names = ("TG_RESOURCE_AI_BASE_URL", "TG_RESOURCE_AI_API_KEY", "TG_RESOURCE_AI_MODEL")
    previous = {name: os.environ.pop(name, None) for name in names}
    try:
        service = ResourceService.__new__(ResourceService)
        service.db = FakeDatabase()
        assert asyncio.run(service.refresh_adult_dialogs()) == "cached"
        assert service.db.saved is None

        original = resource_service.request_ai_json
        resource_service.request_ai_json = lambda *_args, **_kwargs: {"adult_dialog_ids": [7, "7", 999, "invalid"]}
        try:
            assert asyncio.run(service.refresh_adult_dialogs()) == "ai"
            assert service.db.saved == [7]
        finally:
            resource_service.request_ai_json = original
    finally:
        for name, value in previous.items():
            if value is not None:
                os.environ[name] = value


def test_ai_settings_prefer_database_and_keep_environment_fallback():
    class FakeDatabase:
        values = {"ai_base_url": "https://database.example/v1", "ai_model": "database-model"}

        def setting(self, key):
            return self.values.get(key)

    names = ("TG_RESOURCE_AI_BASE_URL", "TG_RESOURCE_AI_API_KEY", "TG_RESOURCE_AI_MODEL")
    previous = {name: os.environ.get(name) for name in names}
    os.environ.update({"TG_RESOURCE_AI_BASE_URL": "https://environment.example/v1", "TG_RESOURCE_AI_API_KEY": "environment-secret", "TG_RESOURCE_AI_MODEL": "environment-model"})
    try:
        service = ResourceService.__new__(ResourceService)
        service.db = FakeDatabase()
        assert service.ai_settings() == {"base_url": "https://database.example/v1", "api_key": "environment-secret", "model": "database-model"}
    finally:
        for name, value in previous.items():
            if value is None:
                os.environ.pop(name, None)
            else:
                os.environ[name] = value


def test_ai_probe_does_not_refresh_adult_cache():
    original = resource_service.request_ai_json
    resource_service.request_ai_json = lambda config, *_args: {"ok": config["api_key"] == "secret"}
    try:
        service = ResourceService.__new__(ResourceService)
        assert asyncio.run(service.probe_ai({"base_url": "https://ai.example/v1", "api_key": "secret", "model": "test"})) is True
    finally:
        resource_service.request_ai_json = original


def test_smart_source_planning_runs_in_the_queued_worker():
    async def scenario():
        task = {
            "id": 1,
            "status": "queued",
            "keyword": "authorized training",
            "source_mode": "smart",
            "source_strategy": "pending",
            "source_dialog_ids": [],
            "sources": [],
            "files_found": 0,
            "files_completed": 0,
            "max_downloads": 1,
            "cancel_requested": False,
        }

        class FakeDatabase:
            def resource_task(self, _task_id):
                return dict(task)

            def update_resource_task(self, _task_id, **values):
                task.update(values)

        class FakeTelegram:
            authorized = True

            async def _ensure_connected(self):
                return None

        service = ResourceService.__new__(ResourceService)
        service.db = FakeDatabase()
        service.telegram = FakeTelegram()
        service.plan_sources = AsyncMock(return_value=([17], "ai"))

        await service._run(1)

        service.plan_sources.assert_awaited_once_with("authorized training", "smart", [])
        assert task["source_dialog_ids"] == [17]
        assert task["source_strategy"] == "ai"
        assert task["status"] == "completed"

    asyncio.run(scenario())


def test_resource_task_attempt_limit_includes_failed_downloads():
    async def scenario():
        task = {
            "id": 1,
            "status": "queued",
            "keyword": "training",
            "file_types": [],
            "source_mode": "manual",
            "source_strategy": "manual",
            "sources": [{"id": 2, "telegram_id": 22, "name": "Training", "username": "training_group", "status": "normal", "resource_compliance_status": "allowed"}],
            "files_found": 0,
            "files_completed": 0,
            "max_downloads": 1,
            "cancel_requested": False,
        }
        candidates = [
            SimpleNamespace(id=10, document=SimpleNamespace(id=110), file=SimpleNamespace(name="training.txt", size=100, mime_type="text/plain"), raw_text="training"),
            SimpleNamespace(id=11, document=SimpleNamespace(id=111), file=SimpleNamespace(name="training.txt", size=200, mime_type="text/plain"), raw_text="training"),
        ]
        claimed = []

        class FakeDatabase:
            def resource_task(self, _task_id):
                return dict(task)

            def update_resource_task(self, _task_id, **values):
                task.update(values)

            def claim_resource_file(self, _task_id, _dialog, message_id, *_args):
                claimed.append(message_id)
                return {"id": message_id, "attempts": 0}

        class FakeClient:
            async def iter_messages(self, *_args, **_kwargs):
                for candidate in candidates:
                    yield candidate

        class FakeTelegram:
            authorized = True
            client = FakeClient()

            async def _ensure_connected(self):
                return None

        service = ResourceService.__new__(ResourceService)
        service.db = FakeDatabase()
        service.telegram = FakeTelegram()
        service.max_file_size = 1024
        service._dialog_entity = AsyncMock(return_value=object())
        service._download = AsyncMock(return_value=False)

        await service._run(1)

        assert claimed == [10]
        service._download.assert_awaited_once()
        assert task["files_found"] == 1
        assert task["files_completed"] == 0
        assert task["status"] == "failed"

    asyncio.run(scenario())


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
            task_keyword = "training-" + uuid.uuid4().hex
            owner_id = 11
            task = db.create_resource_task(owner_id, task_keyword, ["pdf"], [dialog_id], 2)
            tasks, total = db.list_resource_tasks({"q": task_keyword[9:20], "status": "queued", "owner_id": owner_id})
            assert total == 1 and tasks[0]["id"] == task["id"]
            assert db.list_resource_tasks({"q": task_keyword, "status": "completed", "owner_id": owner_id})[1] == 0
            assert db.list_resource_tasks({"q": task_keyword, "owner_id": 999})[1] == 0
            assert db.resource_task(task["id"], 999) is None
            manager = ResourceService(SimpleNamespace(), db, SimpleNamespace())
            manager.cancel(task["id"], owner_id)
            assert db.resource_task(task["id"])["status"] == "cancelled"
            manager.retry(task["id"], owner_id)
            assert db.resource_task(task["id"])["status"] == "queued"
            telegram_file_id = "file-" + uuid.uuid4().hex
            resource = db.claim_resource_file(task["id"], row, 99, telegram_file_id, "training.pdf", 123, "application/pdf", "authorized training")
            assert resource and db.claim_resource_file(task["id"], row, 99, telegram_file_id, "training.pdf", 123, "application/pdf", "authorized training") is None
            db.update_resource_task(task["id"], status="running")
            db.update_resource_file(resource["id"], download_status="downloading")
            running_tasks, _ = db.list_resource_tasks({"q": task_keyword, "owner_id": owner_id})
            assert running_tasks[0]["current_stage"] == "downloading" and running_tasks[0]["current_file_name"] == "training.pdf"
            db.update_resource_file(resource["id"], download_status="success", upload_status="success", object_key="telegram/test/training.pdf", completed_at=now)
            assert db.resource_file(resource["id"], owner_id)["source_name"] == row["name"]
            assert db.resource_file(resource["id"], 999) is None
            db.update_resource_task(task["id"], status="cancelled", cancel_requested=True, completed_at=now)
            assert db.resource_task(task["id"])["status"] == "cancelled"
            acquisition, created = db.create_acquisition_task(task["id"], "candidate-test", "acquisition-" + uuid.uuid4().hex)
            assert created is True
            db.update_acquisition_task(acquisition["id"], state="failed", error_code="timeout")
            retried, created = db.create_acquisition_task(task["id"], "candidate-test", acquisition["idempotency_key"])
            assert created is True and retried["id"] == acquisition["id"] and retried["state"] == "queued"
        finally:
            conn.execute("DELETE FROM tel_acquisition_tasks WHERE id=%s", (locals().get("acquisition", {}).get("id", -1),))
            conn.execute("DELETE FROM tel_resource_files WHERE telegram_chat_id=%s", (telegram_id,))
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


def test_join_records_the_joined_dialog():
    entity = SimpleNamespace(id=700, username="books")

    class Client:
        async def get_entity(self, username):
            assert username == "books"
            return entity

        async def __call__(self, request):
            assert isinstance(request, functions.channels.JoinChannelRequest)

        async def get_dialogs(self, limit=None):
            assert limit is None
            return [SimpleNamespace(entity=entity)]

    class FakeDatabase:
        def upsert_dialog(self, dialog):
            assert dialog == {"telegram_id": 700}
            return 7

        def dialog(self, dialog_id):
            assert dialog_id == 7
            return {"id": 7, "telegram_id": 700, "status": "normal"}

    service = TelegramService.__new__(TelegramService)
    service.client = Client()
    service.db = FakeDatabase()
    service._dialog_data = lambda dialog: {"telegram_id": dialog.entity.id}

    joined = asyncio.run(service.join("https://t.me/books"))
    assert joined["id"] == 7


def test_connect_retries_transient_failure():
    service = TelegramService.__new__(TelegramService)
    service.authorized = False
    attempts = 0

    async def ensure_connected():
        nonlocal attempts
        attempts += 1
        if attempts == 1:
            raise ConnectionError("temporary failure")

    service._ensure_connected = ensure_connected
    service.call = TelegramService.call.__get__(service)
    service.client = SimpleNamespace(
        is_user_authorized=AsyncMock(return_value=True),
        get_me=AsyncMock(return_value=SimpleNamespace(first_name="XY", last_name=None, id=7)),
    )
    service.db = SimpleNamespace(setting=lambda *_args: None)

    with patch("telegram_service.asyncio.sleep", new=AsyncMock()):
        asyncio.run(service.connect())
    assert attempts == 2
    assert service.authorized is True


def test_call_does_not_retry_flood_wait():
    service = TelegramService.__new__(TelegramService)
    action = AsyncMock(side_effect=errors.FloodWaitError(request=None, capture=37))

    with patch("telegram_service.asyncio.sleep", new=AsyncMock()) as sleep:
        try:
            asyncio.run(service.call(action))
            raise AssertionError("FloodWait was swallowed")
        except errors.FloodWaitError:
            pass

    action.assert_awaited_once()
    sleep.assert_not_awaited()


def test_resource_task_records_flood_wait_without_waiting():
    task = {
        "id": 9, "status": "queued", "keyword": "k8s", "source_mode": "manual",
        "source_strategy": "manual", "sources": [{
            "status": "normal", "resource_compliance_status": "allowed", "username": "group",
        }], "files_completed": 0, "files_found": 0, "max_downloads": 1,
    }
    updates = []

    class FakeDatabase:
        def resource_task(self, _task_id):
            return task

        def update_resource_task(self, _task_id, **values):
            task.update(values)
            updates.append(values)

    class FakeClient:
        async def iter_messages(self, *_args, **_kwargs):
            raise errors.FloodWaitError(request=None, capture=86400)
            yield

    class FakeTelegram:
        authorized = True
        client = FakeClient()

        async def _ensure_connected(self):
            pass

    service = ResourceService.__new__(ResourceService)
    service.db = FakeDatabase()
    service.telegram = FakeTelegram()
    service._dialog_entity = AsyncMock(return_value=object())
    service._cancelled = lambda _task_id: False

    asyncio.run(service._run(9))

    assert task["status"] == "failed"
    assert task["error_message"] == "FloodWait: 86400 seconds"
    assert len(updates) == 2


def test_switch_after_flood_wait_marks_config_and_selects_lower_api_id():
    service = TelegramService.__new__(TelegramService)
    service.api_id = 21724
    service.api_hash = "hash-21724"
    service.credential_source = "database"
    service.authorized = True
    service.record_flood_wait = AsyncMock()
    service.refresh_api_credentials = AsyncMock(side_effect=lambda: setattr(service, "api_id", 16623))
    service._ensure_connected = AsyncMock()
    service.client = SimpleNamespace(is_user_authorized=AsyncMock(return_value=True))
    flooded = set()

    asyncio.run(service.switch_after_flood_wait(errors.FloodWaitError(request=None, capture=30), flooded))

    assert flooded == {21724}
    assert service.api_id == 16623
    assert service.authorized is True
    service.record_flood_wait.assert_awaited_once()


def test_fifth_flood_wait_stops_resource_task():
    service = TelegramService.__new__(TelegramService)
    service.api_id = 4
    service.credential_source = "database"
    service.record_flood_wait = AsyncMock()
    service.refresh_api_credentials = AsyncMock()
    flooded = {8, 6, 5, 21724}

    try:
        asyncio.run(service.switch_after_flood_wait(errors.FloodWaitError(request=None, capture=60), flooded))
        raise AssertionError("the fifth FloodWait did not stop the task")
    except RuntimeError as exc:
        assert "5 个 Telegram API ID" in str(exc)

    assert flooded == {8, 6, 5, 21724, 4}
    service.refresh_api_credentials.assert_not_awaited()


def test_resource_task_resumes_after_switching_api_id():
    task = {
        "id": 19, "status": "queued", "keyword": "k8s", "file_types": [],
        "source_mode": "manual", "source_strategy": "manual", "sources": [{
            "id": 2, "telegram_id": 22, "name": "K8s", "username": "k8s_group",
            "status": "normal", "resource_compliance_status": "allowed",
        }], "files_completed": 0, "files_found": 0, "max_downloads": 1,
    }
    candidate = SimpleNamespace(
        id=101, document=SimpleNamespace(id=201),
        file=SimpleNamespace(name="k8s.txt", size=10, mime_type="text/plain"), raw_text="k8s",
    )
    updates = []

    class FakeDatabase:
        def resource_task(self, _task_id):
            return task

        def update_resource_task(self, _task_id, **values):
            task.update(values)
            updates.append(values)

        def claim_resource_file(self, *_args):
            return {"id": 201, "attempts": 0}

    class FakeClient:
        def __init__(self):
            self.scans = 0

        async def get_entity(self, _entity):
            return object()

        async def iter_messages(self, *_args, **_kwargs):
            self.scans += 1
            if self.scans == 1:
                raise errors.FloodWaitError(request=None, capture=30)
            yield candidate

    class FakeTelegram:
        credential_source = "database"
        authorized = True
        api_id = 21724

        def __init__(self):
            self.client = FakeClient()

        async def _ensure_connected(self):
            pass

        async def refresh_api_credentials(self):
            pass

        async def switch_after_flood_wait(self, exc, flooded):
            flooded.add(self.api_id)
            self.api_id = 16623

        async def call_with_fallback(self, operation, _flooded):
            return await operation()

    service = ResourceService.__new__(ResourceService)
    service.db = FakeDatabase()
    service.telegram = FakeTelegram()
    service.max_file_size = 1024
    service._cancelled = lambda _task_id: False
    service._download = AsyncMock(return_value=True)

    asyncio.run(service._run(19))

    assert task["status"] == "completed", task
    assert task["files_found"] == 1
    assert task["files_completed"] == 1
    assert service.telegram.client.scans == 2
    service._download.assert_awaited_once()


if __name__ == "__main__":
    test_resource_validation()
    test_adult_dialog_filter_uses_cached_ids_without_ai()
    test_ai_settings_prefer_database_and_keep_environment_fallback()
    test_ai_probe_does_not_refresh_adult_cache()
    test_metadata_scan_storage()
    test_qr_login_migrates_before_rendering()
    test_join_records_the_joined_dialog()
    test_connect_retries_transient_failure()
    test_call_does_not_retry_flood_wait()
    test_switch_after_flood_wait_marks_config_and_selects_lower_api_id()
    test_fifth_flood_wait_stops_resource_task()
    test_resource_task_resumes_after_switching_api_id()
    print("ok")
