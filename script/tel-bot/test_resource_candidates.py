import os
import tempfile
import unittest
from datetime import datetime, timezone
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import Mock, patch

from resource_service import ResourceCapabilityError, ResourceService, is_interactive_search_group, message_source_usernames


class FakeClient:
    def __init__(self, messages):
        self.messages = messages
        self.entity_requests = []
        self.sent = []

    async def get_entity(self, identifier):
        self.entity_requests.append(identifier)
        return identifier

    async def get_messages(self, _entity, ids):
        return next((message for message in self.messages if message.id == ids), None)

    async def iter_messages(self, _entity, search=None, limit=100, min_id=0, reverse=False):
        _ = limit
        for message in self.messages:
            if message.id <= min_id:
                continue
            filename = getattr(getattr(message, "file", None), "name", "") or ""
            if search is None or search.casefold() in f"{filename} {message.raw_text}".casefold():
                yield message

    async def send_message(self, entity, text):
        self.sent.append((entity, text))
        return SimpleNamespace(id=10)

    async def download_media(self, _message, file):
        Path(file).write_bytes(b"x" * 2048)
        return file


class RoutedFakeClient(FakeClient):
    def __init__(self, messages_by_entity):
        super().__init__([message for messages in messages_by_entity.values() for message in messages])
        self.messages_by_entity = messages_by_entity

    async def iter_messages(self, entity, search=None, limit=100, min_id=0, reverse=False):
        _ = limit, reverse
        for current in self.messages_by_entity.get(entity, []):
            if current.id <= min_id:
                continue
            filename = getattr(getattr(current, "file", None), "name", "") or ""
            if search is None or search.casefold() in f"{filename} {current.raw_text}".casefold():
                yield current


class FakeTelegram:
    def __init__(self, messages):
        self.authorized = True
        self.client = FakeClient(messages)

    async def _ensure_connected(self):
        return None

    async def call(self, function, *args, **kwargs):
        return await function(*args, **kwargs)


class FakeDatabase:
    dialog_record = {
        "id": 7, "telegram_id": 700, "name": "Go 资料", "username": "go_docs",
        "dialog_type": "频道", "last_message_at": datetime.now(timezone.utc),
        "is_group": False, "is_broadcast": True,
        "status": "normal", "resource_compliance_status": "allowed",
    }

    def __init__(self):
        self.dialog_record = self.dialog_record.copy()

    def resource_dialogs(self):
        return [self.dialog_record.copy()]

    def dialog(self, dialog_id):
        return self.dialog_record.copy() if dialog_id == 7 else None

    def create_acquisition_task(self, gva_task_id, candidate_id, idempotency_key):
        self.acquisition = {
            "id": 1, "gva_task_id": gva_task_id, "candidate_id": candidate_id,
            "idempotency_key": idempotency_key, "state": "queued",
        }
        return self.acquisition.copy(), True

    def acquisition_task(self, task_id):
        return self.acquisition.copy() if getattr(self, "acquisition", {}).get("id") == task_id else None

    def update_acquisition_task(self, task_id, **values):
        if self.acquisition["id"] == task_id:
            self.acquisition.update(values)


def message(file_id="900"):
    return SimpleNamespace(
        id=11,
        document=SimpleNamespace(id=int(file_id)),
        file=SimpleNamespace(name="Go并发编程.pdf", size=2048, mime_type="application/pdf"),
        raw_text="Go 并发编程学习资料",
        date=datetime(2026, 10, 29, tzinfo=timezone.utc),
    )


class ResourceCandidateTest(unittest.IsolatedAsyncioTestCase):
    def service(self, messages):
        service = ResourceService.__new__(ResourceService)
        service.telegram = FakeTelegram(messages)
        service.db = FakeDatabase()
        service.max_file_size = 10 * 1024 * 1024
        service.acquisitions = {}
        return service

    @patch.dict(os.environ, {"TELEGRAM_INTERNAL_TOKEN": "candidate-test-secret"})
    async def test_group_name_containing_search_character_uses_interactive_search(self):
        service = self.service([message()])
        service.db.dialog_record.update(name="搜书群", is_group=True, is_broadcast=False)
        service.interactive_search_wait = 0.01

        candidates = await service.search_candidates("Go", ["pdf"], 1)

        self.assertEqual(len(candidates), 1)
        self.assertEqual(service.telegram.client.sent, [("go_docs", "Go")])
        self.assertTrue(is_interactive_search_group(service.db.dialog_record))
        self.assertFalse(is_interactive_search_group({**service.db.dialog_record, "is_group": False}))

    def test_extracts_linked_sources_from_text_entities_and_buttons(self):
        linked = SimpleNamespace(
            raw_text="结果在 https://t.me/Book_Search",
            entities=[SimpleNamespace(url="https://t.me/NovelFiles")],
            buttons=[[SimpleNamespace(url="https://t.me/MoreBooks")]],
        )

        self.assertEqual(message_source_usernames(linked), {"book_search", "novelfiles", "morebooks"})

    @patch.dict(os.environ, {"TELEGRAM_INTERNAL_TOKEN": "candidate-test-secret"})
    async def test_searches_known_source_returned_by_interactive_group(self):
        service = self.service([])
        search_group = {**service.db.dialog_record, "id": 7, "name": "搜书群", "username": "book_search", "is_group": True, "is_broadcast": False}
        linked_channel = {**service.db.dialog_record, "id": 8, "name": "图书频道", "username": "book_files"}
        service.db.resource_dialogs = lambda: [search_group, linked_channel]
        linked_message = SimpleNamespace(id=11, document=None, file=None, raw_text="结果频道 https://t.me/book_files", entities=None, buttons=None)
        service.telegram.client = RoutedFakeClient({"book_search": [linked_message], "book_files": [message()]})
        service.interactive_search_wait = 0.01

        candidates = await service.search_candidates("Go", ["pdf"], 1)

        self.assertEqual(candidates[0]["source"]["name"], "图书频道")
        self.assertEqual(service.telegram.client.sent, [("book_search", "Go")])

    @patch.dict(os.environ, {"TELEGRAM_INTERNAL_TOKEN": "candidate-test-secret"})
    async def test_search_and_inspect_do_not_download(self):
        service = self.service([message()])
        candidates = await service.search_candidates("Go", ["pdf"], 5)
        self.assertEqual(len(candidates), 1)
        self.assertEqual(candidates[0]["license"], "unknown")
        self.assertNotIn("candidate-test-secret", candidates[0]["candidate_id"])

        inspected = await service.inspect_candidate(candidates[0]["candidate_id"])
        self.assertEqual(inspected["name"], "Go并发编程.pdf")
        self.assertEqual(inspected["source"]["name"], "Go 资料")
        self.assertEqual(service.telegram.client.entity_requests, ["go_docs", "go_docs"])

    @patch.dict(os.environ, {"TELEGRAM_INTERNAL_TOKEN": "candidate-test-secret"})
    async def test_rejects_tampered_and_changed_candidates(self):
        service = self.service([message()])
        candidate = (await service.search_candidates("Go", None, 1))[0]["candidate_id"]
        with self.assertRaisesRegex(ResourceCapabilityError, "候选 ID 无效"):
            await service.inspect_candidate(candidate + "x")

        service.telegram.client.messages = [message("901")]
        with self.assertRaisesRegex(ResourceCapabilityError, "已发生变化"):
            await service.inspect_candidate(candidate)

    @patch.dict(os.environ, {"TELEGRAM_INTERNAL_TOKEN": "candidate-test-secret"})
    async def test_acquisition_downloads_exact_candidate_and_completes_delivery(self):
        with tempfile.TemporaryDirectory() as directory:
            service = self.service([message()])
            service.tmp_root = Path(directory)
            service._gva_json = Mock(side_effect=lambda path, _token, _payload: (
                {"uploadUrl": "http://rustfs/upload"} if path.endswith("/upload") else
                {"resource": {"ID": 9}} if path.endswith("/complete") else {}
            ))
            candidate = (await service.search_candidates("Go", None, 1))[0]["candidate_id"]
            with patch("resource_service._put_file") as upload:
                task = await service.submit_acquisition(candidate, 17, "request-17", "delivery-token")
                worker = service.acquisitions[task["id"]]
                await worker

            self.assertEqual(service.db.acquisition["state"], "completed")
            self.assertEqual(service.db.acquisition["stage"], "completed")
            upload.assert_called_once()
            complete_call = next(call for call in service._gva_json.call_args_list if call.args[0].endswith("/complete"))
            complete_payload = complete_call.args[2]
            self.assertEqual(complete_payload["size"], 2048)
            self.assertEqual(len(complete_payload["sha256"]), 64)
            progress_paths = [call.args[0] for call in service._gva_json.call_args_list if call.args[0].endswith("/progress")]
            self.assertEqual(len(progress_paths), 3)

    @patch.dict(os.environ, {"TELEGRAM_INTERNAL_TOKEN": "candidate-test-secret"})
    async def test_acquisition_reports_failure_to_gva(self):
        with tempfile.TemporaryDirectory() as directory:
            service = self.service([message()])
            service.tmp_root = Path(directory)
            service._gva_json = Mock(side_effect=lambda path, _token, _payload: (
                {"uploadUrl": "http://rustfs/upload"} if path.endswith("/upload") else {}
            ))
            candidate = (await service.search_candidates("Go", None, 1))[0]["candidate_id"]
            with patch("resource_service._put_file", side_effect=ResourceCapabilityError("delivery_failed", "上传失败")):
                task = await service.submit_acquisition(candidate, 18, "request-18", "delivery-token")
                await service.acquisitions[task["id"]]

            self.assertEqual(service.db.acquisition["state"], "failed")
            failure = next(call for call in service._gva_json.call_args_list if call.args[0].endswith("/fail"))
            self.assertEqual(failure.args[2]["errorCode"], "delivery_failed")
            self.assertEqual(failure.args[2]["errorMessage"], "资源获取失败")


if __name__ == "__main__":
    unittest.main()
