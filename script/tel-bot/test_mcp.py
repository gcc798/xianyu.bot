import json
import unittest
from datetime import datetime, timezone
from types import SimpleNamespace

from mcp import Client
from mcp.server.transport_security import TransportSecurityMiddleware
from starlette.requests import Request

from mcp_server import build_mcp


class FakeTelegram:
    client = None
    authorized = False
    has_session = False

    async def call(self, function, *args, **kwargs):
        return await function(*args, **kwargs)


class FakeResources:
    max_file_size = 1024

    def _encode_candidate(self, dialog_id, message_id, file_id):
        return f"{dialog_id}:{message_id}:{file_id}"


class FakeDatabase:
    dialog_record = {
        "id": 7, "telegram_id": 700, "name": "搜书群", "username": "books",
        "dialog_type": "超级群", "is_group": True, "is_broadcast": False,
        "last_message_id": 10, "last_message_at": datetime(2026, 10, 30, tzinfo=timezone.utc), "status": "normal",
    }

    def resource_dialogs(self):
        return []

    def accessible_dialogs(self, query="", kind=""):
        row = self.dialog_record
        if query and query.casefold() not in f"{row['name']} {row['username']}".casefold():
            return []
        if kind == "channel":
            return []
        return [row.copy()]

    def dialog(self, dialog_id):
        return self.dialog_record.copy() if dialog_id == 7 else None

    def dialog_by_username(self, username):
        return self.dialog_record.copy() if username.casefold() == "books" else None

    def dialog_by_telegram_id(self, telegram_id):
        return self.dialog_record.copy() if telegram_id == 700 else None


class FakeClient:
    def __init__(self):
        self.sent = []
        self.messages = [SimpleNamespace(
            id=11, raw_text="斗破苍穹 https://t.me/novels", date=datetime(2026, 10, 30, tzinfo=timezone.utc),
            document=SimpleNamespace(id=900), file=SimpleNamespace(name="斗破苍穹.epub", size=2048, mime_type="application/epub+zip"), buttons=None,
        )]

    async def get_entity(self, identifier):
        return SimpleNamespace(id=700, title="搜书群", username="books", megagroup=True, broadcast=False) if identifier == "books" else identifier

    async def send_message(self, entity, text):
        self.sent.append((entity.id, text))
        return SimpleNamespace(id=10, raw_text=text, date=datetime(2026, 10, 30, tzinfo=timezone.utc), document=None, file=None, buttons=None)

    async def iter_messages(self, _entity, search=None, min_id=0, reverse=False, limit=50):
        _ = reverse, limit
        for message in self.messages:
            if message.id > min_id and (not search or search.casefold() in message.raw_text.casefold()):
                yield message


class MCPContractTest(unittest.IsolatedAsyncioTestCase):
    def server(self):
        server, _ = build_mcp(FakeTelegram(), FakeResources(), FakeDatabase())
        return server

    def authorized_server(self):
        telegram = FakeTelegram()
        telegram.client = FakeClient()
        telegram.authorized = True
        server, _ = build_mcp(telegram, FakeResources(), FakeDatabase())
        return server, telegram

    async def test_exposes_fixed_tool_contract(self):
        async with Client(self.server()) as client:
            tools = await client.list_tools()
        self.assertEqual(
            [tool.name for tool in tools.tools],
            [
                "channel_get_info", "channel_get_status", "source_list",
                "dialog_list", "dialog_resolve", "message_send", "message_list", "message_search",
                "source_sync", "resource_search", "resource_inspect", "resource_acquire", "task_get", "task_cancel",
            ],
        )

    async def test_atomic_dialog_and_message_tools(self):
        server, telegram = self.authorized_server()
        async with Client(server) as client:
            dialogs = json.loads((await client.call_tool("dialog_list", {"query": "搜"})).content[0].text)
            sent = json.loads((await client.call_tool("message_send", {"dialog_id": 7, "text": "斗破苍穹"})).content[0].text)
            listed = json.loads((await client.call_tool("message_list", {"dialog_id": 7, "after_message_id": 10})).content[0].text)
            searched = json.loads((await client.call_tool("message_search", {"dialog_id": 7, "query": "斗破苍穹"})).content[0].text)

        self.assertEqual(dialogs["data"]["dialogs"][0]["name"], "搜书群")
        self.assertEqual(sent["data"]["message"]["id"], 10)
        self.assertEqual(telegram.client.sent, [(700, "斗破苍穹")])
        self.assertEqual(listed["data"]["messages"][0]["file"]["candidate_id"], "7:11:900")
        self.assertEqual(listed["data"]["messages"][0]["links"], ["https://t.me/novels"])
        self.assertEqual(searched["data"]["next_message_id"], 11)

    async def test_invalid_task_id_returns_contract_error(self):
        async with Client(self.server()) as client:
            result = await client.call_tool("task_get", {"task_id": "invalid"})
        payload = json.loads(result.content[0].text)
        self.assertEqual(payload["status"], "failed")
        self.assertEqual(payload["error"]["code"], "invalid_argument")

    async def test_status_does_not_expose_credentials(self):
        async with Client(self.server()) as client:
            result = await client.call_tool("channel_get_status", {})
        payload = json.loads(result.content[0].text)
        self.assertEqual(payload, {"status": "success", "data": {"state": "disabled"}, "error": None})

    async def test_streamable_http_accepts_docker_service_host(self):
        server, _ = build_mcp(FakeTelegram(), FakeResources(), FakeDatabase())
        middleware = TransportSecurityMiddleware(server.session_manager.security_settings)
        request = Request({
            "type": "http",
            "method": "POST",
            "path": "/",
            "headers": [(b"host", b"telegram:8080"), (b"content-type", b"application/json")],
        })
        self.assertIsNone(await middleware.validate_request(request, is_post=True))


if __name__ == "__main__":
    unittest.main()
