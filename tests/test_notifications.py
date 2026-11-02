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
            get_group_list = AsyncMock(return_value=[])

        bot = FakeBot()
        with (
            patch.object(module, "Bot", FakeBot),
            patch.object(module, "get_bots", return_value={"10000": bot}),
        ):
            response = await self.request(
                {"qq": "123456789", "message": "[CQ:at,qq=all]"}
            )
            self.assertEqual(json.loads(response.content)["status"], "sent")
            bot.get_group_list.assert_not_awaited()
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

    async def test_common_group_temporary_private_delivery(self):
        class GroupBot:
            self_id = "10000"
            get_friend_list = AsyncMock(return_value=[])
            get_group_list = AsyncMock(
                return_value=[{"group_id": 10001}, {"group_id": "10002"}]
            )
            get_group_member_list = AsyncMock(
                side_effect=[[{"user_id": 99999}], [{"user_id": "123456789"}]]
            )
            call_api = AsyncMock()
            send_group_msg = AsyncMock()

        bot = GroupBot()
        with (
            patch.object(module, "Bot", GroupBot),
            patch.object(module, "get_bots", return_value={"10000": bot}),
        ):
            result = await self.request({"qq": "123456789", "message": "[CQ:at,qq=all]"})
        self.assertEqual(json.loads(result.content)["status"], "sent")
        bot.call_api.assert_awaited_once_with(
            "send_private_msg",
            user_id=123456789,
            group_id=10002,
            message="[CQ:at,qq=all]",
            auto_escape=True,
        )
        bot.send_group_msg.assert_not_awaited()

    async def test_temporary_failure_tries_next_group_and_stops_on_success(self):
        bot = SimpleNamespace(
            get_friend_list=AsyncMock(return_value=[]),
            get_group_list=AsyncMock(return_value=[{"group_id": n} for n in (1, 2, 3)]),
            get_group_member_list=AsyncMock(return_value=[{"user_id": 123456789}]),
            call_api=AsyncMock(
                side_effect=[RuntimeError("temporary session denied"), {}]
            ),
        )
        with patch.object(module, "logger"):
            self.assertEqual(
                await module.deliver_notification([bot], "123456789", "text"), "sent"
            )
        self.assertEqual(bot.call_api.await_count, 2)
        self.assertEqual(bot.get_group_member_list.await_count, 2)
        self.assertEqual(bot.call_api.await_args.kwargs["group_id"], 2)

    async def test_friend_on_second_bot_precedes_common_group(self):
        first = SimpleNamespace(
            get_friend_list=AsyncMock(return_value=[]),
            get_group_list=AsyncMock(),
        )
        second = SimpleNamespace(
            get_friend_list=AsyncMock(return_value=[{"user_id": 123456789}]),
            send_private_msg=AsyncMock(),
            get_group_list=AsyncMock(),
        )
        self.assertEqual(
            await module.deliver_notification([first, second], "123456789", "text"),
            "sent",
        )
        first.get_group_list.assert_not_awaited()
        second.get_group_list.assert_not_awaited()

    async def test_group_lookup_and_send_failures_are_retryable(self):
        for stage in ("groups", "members", "send"):
            with self.subTest(stage=stage):
                bot = SimpleNamespace(
                    get_friend_list=AsyncMock(return_value=[]),
                    get_group_list=AsyncMock(return_value=[{"group_id": 1}]),
                    get_group_member_list=AsyncMock(
                        return_value=[{"user_id": 123456789}]
                    ),
                    call_api=AsyncMock(),
                )
                operation = {
                    "groups": bot.get_group_list,
                    "members": bot.get_group_member_list,
                    "send": bot.call_api,
                }[stage]
                operation.side_effect = RuntimeError("API unavailable")
                with patch.object(module, "logger"):
                    self.assertEqual(
                        await module.deliver_notification([bot], "123456789", "text"),
                        "delivery_failed",
                    )

    async def test_no_common_group_skips_without_sending(self):
        bot = SimpleNamespace(
            get_friend_list=AsyncMock(return_value=[]),
            get_group_list=AsyncMock(return_value=[{"group_id": 1}]),
            get_group_member_list=AsyncMock(return_value=[{"user_id": 99999}]),
            call_api=AsyncMock(),
        )
        self.assertEqual(
            await module.deliver_notification([bot], "123456789", "text"), "not_friend"
        )
        bot.call_api.assert_not_awaited()
