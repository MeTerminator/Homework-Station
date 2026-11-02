"""Forward quoted private-message replies to the homework platform."""

import asyncio
import json
from urllib.error import HTTPError
from urllib.parse import urlsplit
from urllib.request import HTTPRedirectHandler, Request, build_opener

from nonebot import on_message
from nonebot.adapters.onebot.v11 import Bot, PrivateMessageEvent
from nonebot.adapters.onebot.v11.permission import PRIVATE
from nonebot.log import logger


class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def validate_platform_url(url: str) -> str:
    parsed = urlsplit(url)
    if (
        parsed.scheme not in ("http", "https")
        or not parsed.netloc
        or parsed.username is not None
        or parsed.password is not None
        or parsed.query
        or parsed.fragment
        or url.endswith("/")
        or any(c.isspace() for c in url)
    ):
        raise ValueError(
            "HOMEWORK_PLATFORM_URL 必须是 HTTP(S) 平台地址，结尾不带 /，不含查询参数或凭据"
        )
    return url


def post_reply(url: str, secret: str, payload: dict) -> tuple[int, dict]:
    request = Request(
        url + "/api/bot/replies",
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": "Bearer " + secret},
        method="POST",
    )
    try:
        response = build_opener(NoRedirect()).open(request, timeout=15)
    except HTTPError as exc:
        response = exc
    with response:
        raw = response.read(4096)
        try:
            body = json.loads(raw)
        except (ValueError, UnicodeDecodeError):
            body = {}
        return response.code, body if isinstance(body, dict) else {}


async def forward_reply(config, bot: Bot, event: PrivateMessageEvent) -> str:
    # The adapter removes the reply segment when resolving event.reply. Reject
    # non-text segments rather than silently dropping attachments or CQ actions.
    if any(segment.type != "text" for segment in event.message):
        return "QQ 回复目前仅支持文字，图片和附件请前往平台上传。"
    content = event.get_plaintext().strip()
    if not content or len(content.encode()) > 2 * 1024 * 1024:
        return "请引用作业评论并填写回复文字，文本不能超过 2MB。"
    payload = {
        "bot_id": bot.self_id,
        "qq": str(event.user_id),
        "message_id": str(event.message_id),
        "reply_message_id": str(event.reply.message_id),
        "content": content,
    }
    try:
        code, body = await asyncio.to_thread(
            post_reply,
            config.homework_platform_url,
            config.homework_notification_secret.get_secret_value(),
            payload,
        )
    except Exception:
        logger.exception("QQ 评论回复请求失败")
        return "暂时无法确认回复是否成功，请到平台查看后再重试。"
    if code == 201:
        return "回复已发送到平台。"
    if code == 401:
        return "回复失败：Bot 与平台的通知 Secret 不一致，请联系管理员。"
    if code in (400, 403, 404, 409):
        return "回复失败：" + str(body.get("error", "请在平台上查看作业状态后重试。"))
    return "暂时无法确认回复是否成功，请到平台查看后再重试。"


def setup_replies(config) -> None:
    validate_platform_url(config.homework_platform_url)
    secret = config.homework_notification_secret.get_secret_value()
    if len(secret.encode()) < 32 or "\r" in secret or "\n" in secret:
        raise ValueError("QQ 回复需要配置至少 32 字节的 HOMEWORK_NOTIFICATION_SECRET")

    async def quoted_notification(bot: Bot, event: PrivateMessageEvent) -> bool:
        return (
            bool(event.reply)
            and str(event.reply.sender.user_id) == bot.self_id
            and (
                not config.homework_notification_bot_id
                or config.homework_notification_bot_id == bot.self_id
            )
        )

    matcher = on_message(
        rule=quoted_notification, permission=PRIVATE, priority=9, block=True
    )

    @matcher.handle()
    async def handle(bot: Bot, event: PrivateMessageEvent) -> None:
        await matcher.finish(await forward_reply(config, bot, event))
