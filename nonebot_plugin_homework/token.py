"""HS256 binding token contract shared with the Gin backend."""

import base64
import hashlib
import hmac
import json
import secrets
import time


def create_binding_code(qq: str, secret: str, now: int | None = None) -> str:
    if not qq.isdigit() or not 5 <= len(qq) <= 15 or qq.startswith("0"):
        raise ValueError("QQ 号格式不正确")
    if len(secret.encode()) < 32:
        raise ValueError("绑定码 secret 至少需要 32 字节")
    issued_at = int(time.time()) if now is None else now
    payload = {
        "qq": qq,
        "ttl": 300,
        "iat": issued_at,
        "exp": issued_at + 300,
        "aud": "homework-qq-bind",
        "jti": secrets.token_hex(16),
    }

    def encode(value: dict) -> str:
        raw = json.dumps(value, separators=(",", ":")).encode()
        return base64.urlsafe_b64encode(raw).decode().rstrip("=")

    message = f"{encode({'alg': 'HS256', 'typ': 'JWT'})}.{encode(payload)}"
    signature = hmac.new(secret.encode(), message.encode(), hashlib.sha256).digest()
    return message + "." + base64.urlsafe_b64encode(signature).decode().rstrip("=")
