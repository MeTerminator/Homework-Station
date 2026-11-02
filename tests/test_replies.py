import importlib.util
import json
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import Mock, patch

from nonebot.adapters.onebot.v11 import Message, MessageSegment
from pydantic import SecretStr

spec = importlib.util.spec_from_file_location(
    "homework_replies", Path(__file__).parents[1] / "nonebot_plugin_homework/replies.py"
)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class ReplyTests(unittest.IsolatedAsyncioTestCase):
    def setUp(self):
        self.config = SimpleNamespace(
            homework_platform_url="https://homework.example.com",
            homework_notification_secret=SecretStr("s" * 32),
            homework_notification_bot_id="",
        )
        self.bot = SimpleNamespace(self_id="10000")
        self.event = SimpleNamespace(
            user_id=123456789,
            message_id=456,
            reply=SimpleNamespace(message_id=-123),
            message=Message(MessageSegment.text("已修改\n谢谢 [CQ:at,qq=all]")),
        )
        self.event.get_plaintext = self.event.message.extract_plain_text

    async def test_forward_text_and_credentials(self):
        with patch.object(module, "post_reply", return_value=(201, {"id": 7})) as post:
            self.assertEqual(
                await module.forward_reply(self.config, self.bot, self.event),
                "回复已发送到平台。",
            )
        post.assert_called_once_with(
            "https://homework.example.com",
            "s" * 32,
            {
                "bot_id": "10000",
                "qq": "123456789",
                "message_id": "456",
                "reply_message_id": "-123",
                "content": "已修改\n谢谢 [CQ:at,qq=all]",
            },
        )

    async def test_attachment_and_empty_reply_do_not_post(self):
        for message in (
            Message("  "),
            Message(MessageSegment.image("https://example.com/a.png")),
        ):
            self.event.message = message
            self.event.get_plaintext = message.extract_plain_text
            with patch.object(module, "post_reply") as post:
                result = await module.forward_reply(self.config, self.bot, self.event)
                self.assertTrue(result)
                post.assert_not_called()

    async def test_platform_errors_and_network_failure(self):
        for code in (400, 403, 404, 409):
            with patch.object(
                module, "post_reply", return_value=(code, {"error": "已撤回"})
            ):
                self.assertIn(
                    "已撤回",
                    await module.forward_reply(self.config, self.bot, self.event),
                )
        with patch.object(module, "post_reply", return_value=(401, {})):
            self.assertIn(
                "Secret", await module.forward_reply(self.config, self.bot, self.event)
            )
        with (
            patch.object(module, "post_reply", side_effect=TimeoutError),
            patch.object(module, "logger"),
        ):
            self.assertIn(
                "确认", await module.forward_reply(self.config, self.bot, self.event)
            )

    async def test_matcher_only_handles_quotes_to_selected_bot(self):
        matcher = Mock()
        with patch.object(module, "on_message", return_value=matcher) as on_message:
            module.setup_replies(self.config)
        rule = on_message.call_args.kwargs["rule"]
        self.event.reply.sender = SimpleNamespace(user_id=10000)
        self.assertTrue(await rule(self.bot, self.event))
        self.config.homework_notification_bot_id = "20000"
        self.assertFalse(await rule(self.bot, self.event))
        self.config.homework_notification_bot_id = ""
        self.event.reply.sender.user_id = 99999
        self.assertFalse(await rule(self.bot, self.event))
        self.event.reply = None
        self.assertFalse(await rule(self.bot, self.event))

    def test_platform_url_validation(self):
        for url in (
            "https://example.com/",
            "file:///tmp/x",
            "https://u:p@example.com",
            "https://example.com?x=1",
            "https://example.com/#a",
            "",
        ):
            with self.subTest(url=url), self.assertRaises(ValueError):
                module.validate_platform_url(url)
        self.assertEqual(
            module.validate_platform_url("http://localhost:8080"), "http://localhost:8080"
        )

    def test_http_contract(self):
        response = Mock(code=201)
        response.read.return_value = b'{"id":7}'
        response.__enter__ = Mock(return_value=response)
        response.__exit__ = Mock(return_value=False)
        opener = Mock()
        opener.open.return_value = response
        with patch.object(module, "build_opener", return_value=opener):
            self.assertEqual(
                module.post_reply(
                    self.config.homework_platform_url, "s" * 32, {"content": "回复"}
                ),
                (201, {"id": 7}),
            )
        request = opener.open.call_args.args[0]
        self.assertEqual(request.full_url, "https://homework.example.com/api/bot/replies")
        self.assertEqual(request.get_header("Authorization"), "Bearer " + "s" * 32)
        self.assertEqual(json.loads(request.data), {"content": "回复"})
