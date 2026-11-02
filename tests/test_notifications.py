import importlib.util
import json
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock, patch

from nonebot.drivers import ASGIMixin, Request
from pydantic import SecretStr

spec = importlib.util.spec_from_file_location(
    "homework_notifications",
    Path(__file__).parents[1] / "nonebot_plugin_homework/notifications.py",
)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class NotificationTests(unittest.IsolatedAsyncioTestCase):
    def setUp(self):
        self.config = SimpleNamespace(
            homework_notification_secret=SecretStr("s" * 32),
            homework_notification_path="/homework/notifications",
            homework_notification_bot_id="",
        )
        self.driver = Mock(spec=ASGIMixin)
        module.setup_notifications(self.driver, self.config)
        self.handler = self.driver.setup_http_server.call_args.args[0].handle_func

    async def request(self, data=None, auth=True):
        headers = {"Authorization": "Bearer " + "s" * 32} if auth else {}
        request = Request(
            "POST",
            "http://localhost/homework/notifications",
            headers=headers,
            json=data if data is not None else {"qq": "123456789", "message": "feedback"},
        )
        return await self.handler(request)

    async def test_auth_and_validation(self):
        with patch.object(module, "get_bots") as get_bots:
            self.assertEqual((await self.request(auth=False)).status_code, 401)
            self.assertEqual(
                (await self.request({"qq": "bad", "message": "feedback"})).status_code,
                400,
            )
            get_bots.assert_not_called()

    async def test_friend_delivery_and_plain_text(self):
        class FakeBot:
            self_id = "10000"
            get_friend_list = AsyncMock(return_value=[{"user_id": 123456789}])
            send_private_msg = AsyncMock()

        bot = FakeBot()
        with (
            patch.object(module, "Bot", FakeBot),
            patch.object(module, "get_bots", return_value={"10000": bot}),
        ):
            response = await self.request(
                {"qq": "123456789", "message": "[CQ:at,qq=all]"}
            )
            self.assertEqual(json.loads(response.content)["status"], "sent")
            bot.send_private_msg.assert_awaited_once_with(
                user_id=123456789, message="[CQ:at,qq=all]", auto_escape=True
            )
            bot.send_private_msg.reset_mock()
            bot.get_friend_list.return_value = []
            response = await self.request()
            self.assertEqual(json.loads(response.content)["status"], "not_friend")
            bot.send_private_msg.assert_not_awaited()
            self.config.homework_notification_bot_id = "20000"
            self.assertEqual((await self.request()).status_code, 503)

    async def test_offline_and_api_failure(self):
        with patch.object(module, "get_bots", return_value={}):
            self.assertEqual((await self.request()).status_code, 503)

        class FailingBot:
            self_id = "10000"
            get_friend_list = AsyncMock(side_effect=RuntimeError("offline"))

        with (
            patch.object(module, "Bot", FailingBot),
            patch.object(module, "get_bots", return_value={"10000": FailingBot()}),
            patch.object(module, "logger"),
        ):
            self.assertEqual((await self.request()).status_code, 503)

    def test_invalid_config(self):
        self.config.homework_notification_secret = SecretStr("short")
        with self.assertRaises(ValueError):
            module.setup_notifications(self.driver, self.config)
