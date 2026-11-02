"""Authenticated comment delivery through NoneBot's portable HTTP routing layer."""

import asyncio
import hmac
import json
import re

from nonebot import get_bots
from nonebot.adapters.onebot.v11 import Bot
from nonebot.drivers import URL, ASGIMixin, HTTPServerSetup, Request, Response
from nonebot.log import logger


async def deliver_notification(bots: list[Bot], qq: str, message: str) -> str:
    failed = False
    # Prefer an existing friend connection across all eligible bot accounts.
    for bot in bots:
        try:
            friends = await bot.get_friend_list()
            if any(str(friend["user_id"]) == qq for friend in friends):
                await bot.send_private_msg(
                    user_id=int(qq), message=message, auto_escape=True
                )
                return "sent"
        except Exception:
            failed = True
            logger.exception("评论好友私聊通知失败")

    # A group is only a source for a temporary PRIVATE session. Never post the
    # homework comment to the group itself. Stop after the first successful send.
    for bot in bots:
        try:
            groups = await bot.get_group_list()
        except Exception:
            failed = True
            logger.exception("评论通知获取共同群列表失败")
            continue
        for group in groups:
            try:
                group_id = int(group["group_id"])
                members = await bot.get_group_member_list(group_id=group_id)
                if not any(str(member["user_id"]) == qq for member in members):
                    continue
                # group_id is an implementation extension for temporary sessions;
                # call_api forwards it without relying on the adapter's type stub.
                await bot.call_api(
                    "send_private_msg",
                    user_id=int(qq),
                    group_id=group_id,
                    message=message,
                    auto_escape=True,
                )
                return "sent"
            except Exception:
                failed = True
                logger.exception("评论共同群临时私聊失败，继续尝试其他群")
    # Unknown membership or rejected temporary sessions must remain retryable.
    return "delivery_failed" if failed else "not_friend"


def setup_notifications(driver, config) -> None:
    secret = config.homework_notification_secret.get_secret_value()
    if len(secret.encode()) < 32 or "\r" in secret or "\n" in secret:
        raise ValueError("HOMEWORK_NOTIFICATION_SECRET 至少 32 字节且不能包含换行")
    path = config.homework_notification_path
    if (
        not path.startswith("/")
        or path.startswith("//")
        or any(c in path for c in "?#\\")
    ):
        raise ValueError("HOMEWORK_NOTIFICATION_PATH 必须是站内绝对路径")
    if not isinstance(driver, ASGIMixin):
        raise ValueError("评论通知需要支持 ASGI 的 NoneBot 驱动器")
    lock = asyncio.Lock()

    def response(code: int, status: str) -> Response:
        return Response(
            code,
            content=json.dumps({"status": status}),
            headers={"Content-Type": "application/json"},
        )

    async def handle(request: Request) -> Response:
        supplied = request.headers.get("Authorization", "")
        if not hmac.compare_digest(supplied.encode(), ("Bearer " + secret).encode()):
            return response(401, "unauthorized")
        data = request.json
        if not isinstance(data, dict):
            return response(400, "invalid_request")
        qq, message = data.get("qq"), data.get("message")
        if (
            not isinstance(qq, str)
            or not re.fullmatch(r"[1-9][0-9]{4,14}", qq)
            or not isinstance(message, str)
            or not message.strip()
            or len(message.encode()) > 3 * 1024 * 1024
        ):
            return response(400, "invalid_request")
        async with lock:
            bots = [
                bot
                for bot in get_bots().values()
                if isinstance(bot, Bot)
                and (
                    not config.homework_notification_bot_id
                    or bot.self_id == config.homework_notification_bot_id
                )
            ]
            if not bots:
                return response(503, "bot_offline")
            status = await deliver_notification(bots, qq, message)
            return response(503 if status == "delivery_failed" else 200, status)

    driver.setup_http_server(
        HTTPServerSetup(
            path=URL(path),
            method="POST",
            name="homework_notifications",
            handle_func=handle,
        )
    )
