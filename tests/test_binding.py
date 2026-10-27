import base64
import hashlib
import hmac
import importlib.util
import json
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location(
    "binding_token", Path(__file__).parents[1] / "nonebot_plugin_homework/token.py"
)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class BindingTests(unittest.TestCase):
    def test_token_contract(self):
        secret = "shared-secret-for-tests-32-characters"
        token = module.create_binding_code("123456789", secret, 1793116800)
        head, body, signature = token.split(".")
        decode = lambda s: base64.urlsafe_b64decode(s + "=" * (-len(s) % 4))
        claims = json.loads(decode(body))
        self.assertEqual(json.loads(decode(head))["alg"], "HS256")
        self.assertEqual(claims["exp"] - claims["iat"], 300)
        self.assertEqual(claims["ttl"], 300)
        self.assertEqual(claims["aud"], "homework-qq-bind")
        self.assertEqual(claims["qq"], "123456789")
        expected = hmac.new(
            secret.encode(), f"{head}.{body}".encode(), hashlib.sha256
        ).digest()
        self.assertTrue(hmac.compare_digest(expected, decode(signature)))
        self.assertNotEqual(
            token, module.create_binding_code("123456789", secret, 1793116800)
        )

    def test_reject_invalid_configuration(self):
        with self.assertRaises(ValueError):
            module.create_binding_code("123456", "short")
        with self.assertRaises(ValueError):
            module.create_binding_code("abc", "x" * 32)
