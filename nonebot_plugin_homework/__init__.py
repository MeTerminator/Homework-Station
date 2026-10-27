"""NoneBot2 + OneBot v11: private /绑定码 command."""

from nonebot import get_plugin_config, on_command
from nonebot.adapters.onebot.v11 import PrivateMessageEvent
from nonebot.adapters.onebot.v11.permission import PRIVATE
from nonebot.plugin import PluginMetadata
from pydantic import BaseModel, Field, SecretStr

from .token import create_binding_code


class Config(BaseModel):
    # You may hardcode the same secret as the admin settings here, but environment
    # configuration keeps it out of source control.
    homework_binding_secret: SecretStr = Field(...)


__plugin_meta__ = PluginMetadata(
    name="作业驿站绑定码",
    description="通过私聊签发 5 分钟有效的 QQ 绑定 JWT",
    usage="私聊发送 /绑定码",
    type="application",
    homepage="https://github.com/nonebot/nonebot2",
    config=Config,
    supported_adapters={"~onebot.v11"},
)

config = get_plugin_config(Config)
secret = config.homework_binding_secret.get_secret_value()
if len(secret.encode()) < 32:
    raise ValueError(
        "HOMEWORK_BINDING_SECRET 至少需要 32 字节，与作业驿站后台设置保持一致"
    )

binding = on_command("绑定码", permission=PRIVATE, priority=10, block=True)


@binding.handle()
async def handle_binding(event: PrivateMessageEvent) -> None:
    code = create_binding_code(str(event.user_id), secret)
    await binding.finish(
        f"您的绑定码为：\n\n{code}\n\n绑定码 5 分钟内有效，请勿发送给他人"
    )
