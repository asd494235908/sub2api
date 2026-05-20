import importlib.util
import json
from pathlib import Path
import unittest


SCRIPT_PATH = Path(__file__).resolve().parent / "check_openai_401_accounts.py"


def load_module():
    spec = importlib.util.spec_from_file_location("check_openai_401_accounts", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CheckOpenAI401AccountsTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.mod = load_module()

    def test_parse_sse_success_event(self):
        body = "\n".join(
            [
                'data: {"type":"test_start","model":"gpt-5.5"}',
                'data: {"type":"content","text":"pong"}',
                'data: {"type":"test_complete","success":true}',
            ]
        )

        result = self.mod.parse_test_sse(body)

        self.assertTrue(result["complete"])
        self.assertEqual(result["text"], "pong")
        self.assertEqual(result["error"], "")

    def test_classify_401_token_invalidated(self):
        error = json.dumps(
            {
                "error": {
                    "message": "Your authentication token has been invalidated.",
                    "code": "token_invalidated",
                },
                "status": 401,
            }
        )

        self.assertEqual(self.mod.classify_error(error), "401_unauthorized")

    def test_classify_401_token_revoked(self):
        error = "Encountered invalidated oauth token for user, code token_revoked, status 401"

        self.assertEqual(self.mod.classify_error(error), "401_unauthorized")

    def test_classify_429_not_401(self):
        error = '{"error":{"type":"usage_limit_reached","message":"The usage limit has been reached"},"status":429}'

        self.assertEqual(self.mod.classify_error(error), "429_rate_limit")
        self.assertNotEqual(self.mod.classify_error(error), "401_unauthorized")

    def test_is_401_result_requires_failed_401(self):
        self.assertTrue(self.mod.is_401_result({"result": "FAILED", "class": "401_unauthorized"}))
        self.assertFalse(self.mod.is_401_result({"result": "OK", "class": ""}))
        self.assertFalse(self.mod.is_401_result({"result": "FAILED", "class": "429_rate_limit"}))


if __name__ == "__main__":
    unittest.main()
