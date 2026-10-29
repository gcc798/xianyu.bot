import os
import tempfile
import unittest
from datetime import datetime, timezone
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import Mock, patch

from resource_service import ResourceCapabilityError, ResourceService


class FakeClient:
    def __init__(self, messages):
        self.messages = messages

    async def get_entity(self, telegram_id):
        return telegram_id

    async def get_messages(self, _entity, ids):
        return next((message for message in self.messages if message.id == ids), None)

    async def iter_messages(self, _entity, search, limit):
        _ = limit
        for message in self.messages:
            if search.casefold() in f"{message.file.name} {message.raw_text}".casefold():
                yield message

    async def download_media(self, _message, file):
        Path(file).write_bytes(b"x" * 2048)
        return file


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
        "status": "normal", "resource_compliance_status": "allowed",
    }

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
    async def test_search_and_inspect_do_not_download(self):
        service = self.service([message()])
        candidates = await service.search_candidates("Go", ["pdf"], 5)
        self.assertEqual(len(candidates), 1)
        self.assertEqual(candidates[0]["license"], "unknown")
        self.assertNotIn("candidate-test-secret", candidates[0]["candidate_id"])

        inspected = await service.inspect_candidate(candidates[0]["candidate_id"])
        self.assertEqual(inspected["name"], "Go并发编程.pdf")
        self.assertEqual(inspected["source"]["name"], "Go 资料")

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
            service._gva_json = Mock(side_effect=[{"uploadUrl": "http://rustfs/upload"}, {"resource": {"ID": 9}}])
            candidate = (await service.search_candidates("Go", None, 1))[0]["candidate_id"]
            with patch("resource_service._put_file") as upload:
                task = await service.submit_acquisition(candidate, 17, "request-17", "delivery-token")
                worker = service.acquisitions[task["id"]]
                await worker

            self.assertEqual(service.db.acquisition["state"], "completed")
            self.assertEqual(service.db.acquisition["stage"], "completed")
            upload.assert_called_once()
            complete_payload = service._gva_json.call_args_list[1].args[2]
            self.assertEqual(complete_payload["size"], 2048)
            self.assertEqual(len(complete_payload["sha256"]), 64)


if __name__ == "__main__":
    unittest.main()
