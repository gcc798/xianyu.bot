import json
import unittest

from mcp import Client

from mcp_server import build_mcp


class FakeTelegram:
    client = None
    authorized = False
    has_session = False


class FakeResources:
    max_file_size = 1024


class FakeDatabase:
    def resource_dialogs(self):
        return []


class MCPContractTest(unittest.IsolatedAsyncioTestCase):
    def server(self):
        server, _ = build_mcp(FakeTelegram(), FakeResources(), FakeDatabase())
        return server

    async def test_exposes_fixed_tool_contract(self):
        async with Client(self.server()) as client:
            tools = await client.list_tools()
        self.assertEqual(
            [tool.name for tool in tools.tools],
            [
                "channel_get_info", "channel_get_status", "source_list", "source_sync",
                "resource_search", "resource_inspect", "resource_acquire", "task_get", "task_cancel",
            ],
        )

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


if __name__ == "__main__":
    unittest.main()
