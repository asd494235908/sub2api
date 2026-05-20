#!/usr/bin/env python3
"""Check remote sub2api OpenAI accounts and list accounts returning 401."""

from __future__ import annotations

import argparse
import base64
import json
import os
import shutil
import subprocess
import sys
import textwrap
from collections import Counter
from pathlib import Path
from typing import Any


DEFAULT_HOST = "101.96.208.132"
DEFAULT_USER = "root"
DEFAULT_MODEL = "gpt-5.5"
DEFAULT_TIMEOUT = 35
DEFAULT_REMOTE_BASE_URL = "http://127.0.0.1:8080"


def parse_test_sse(body: str) -> dict[str, Any]:
    texts: list[str] = []
    error = ""
    complete = False
    events: list[str] = []

    for raw in body.splitlines():
        line = raw.strip()
        if not line.startswith("data:"):
            continue
        data = line[5:].strip()
        if not data or data == "[DONE]":
            continue
        try:
            event = json.loads(data)
        except json.JSONDecodeError:
            continue

        event_type = str(event.get("type", ""))
        events.append(event_type)
        if event_type == "content" and event.get("text"):
            texts.append(str(event["text"]))
        elif event_type == "error":
            error = str(event.get("error", ""))
        elif event_type == "test_complete" and event.get("success") is True:
            complete = True

    return {
        "complete": complete,
        "text": "".join(texts).strip(),
        "error": error,
        "events": events,
    }


def classify_error(error: str) -> str:
    lowered = (error or "").lower()
    if not lowered:
        return ""
    if (
        "401" in lowered
        or "unauthorized" in lowered
        or "token_invalidated" in lowered
        or "token revoked" in lowered
        or "token_revoked" in lowered
        or "invalidated oauth token" in lowered
        or "authentication token has been invalidated" in lowered
    ):
        return "401_unauthorized"
    if "403" in lowered or "forbidden" in lowered:
        return "403_forbidden"
    if "429" in lowered or "rate limit" in lowered or "usage_limit_reached" in lowered:
        return "429_rate_limit"
    if "502" in lowered or "upstream" in lowered:
        return "upstream_or_502"
    if (
        "404" in lowered
        or "not found" in lowered
        or "unsupported" in lowered
        or "not support" in lowered
    ):
        return "model_or_endpoint_unsupported"
    if "timeout" in lowered or "timed out" in lowered:
        return "timeout"
    return "other"


def is_401_result(result: dict[str, Any]) -> bool:
    return result.get("result") == "FAILED" and result.get("class") == "401_unauthorized"


def build_remote_python(model: str, timeout: int, limit: int | None, base_url: str) -> str:
    model_json = json.dumps(model, ensure_ascii=False)
    base_url_json = json.dumps(base_url, ensure_ascii=False)
    limit_value = "None" if limit is None else str(limit)

    return f"""#!/usr/bin/env python3
import json
import subprocess
import sys
import time
import urllib.error
import urllib.request
from collections import Counter

MODEL = {model_json}
TIMEOUT = {int(timeout)}
LIMIT = {limit_value}
BASE_URL = {base_url_json}


def psql(sql):
    cmd = [
        "docker", "exec", "-i", "sub2api-postgres",
        "psql", "-U", "sub2api", "-d", "sub2api",
        "-At", "-F", "\\t", "-v", "ON_ERROR_STOP=1", "-c", sql,
    ]
    return subprocess.check_output(cmd, text=True)


def parse_test_sse(body):
    texts = []
    error = ""
    complete = False
    events = []
    for raw in body.splitlines():
        line = raw.strip()
        if not line.startswith("data:"):
            continue
        data = line[5:].strip()
        if not data or data == "[DONE]":
            continue
        try:
            event = json.loads(data)
        except json.JSONDecodeError:
            continue
        event_type = str(event.get("type", ""))
        events.append(event_type)
        if event_type == "content" and event.get("text"):
            texts.append(str(event["text"]))
        elif event_type == "error":
            error = str(event.get("error", ""))
        elif event_type == "test_complete" and event.get("success") is True:
            complete = True
    return {{"complete": complete, "text": "".join(texts).strip(), "error": error, "events": events}}


def classify_error(error):
    lowered = (error or "").lower()
    if not lowered:
        return ""
    if (
        "401" in lowered
        or "unauthorized" in lowered
        or "token_invalidated" in lowered
        or "token revoked" in lowered
        or "token_revoked" in lowered
        or "invalidated oauth token" in lowered
        or "authentication token has been invalidated" in lowered
    ):
        return "401_unauthorized"
    if "403" in lowered or "forbidden" in lowered:
        return "403_forbidden"
    if "429" in lowered or "rate limit" in lowered or "usage_limit_reached" in lowered:
        return "429_rate_limit"
    if "502" in lowered or "upstream" in lowered:
        return "upstream_or_502"
    if "404" in lowered or "not found" in lowered or "unsupported" in lowered or "not support" in lowered:
        return "model_or_endpoint_unsupported"
    if "timeout" in lowered or "timed out" in lowered:
        return "timeout"
    return "other"


def query_accounts():
    sql = '''
with rows as (
  select a.id, a.name, a.type, a.status, a.schedulable,
         coalesce(string_agg(ag.group_id::text, ',' order by ag.group_id), '') as groups,
         coalesce(a.last_used_at::text, '') as last_used_at,
         coalesce(a.temp_unschedulable_until::text, '') as temp_unschedulable_until,
         left(coalesce(a.error_message, ''), 240) as db_error_message
  from accounts a
  left join account_groups ag on ag.account_id = a.id
  where a.platform = 'openai' and a.deleted_at is null
  group by a.id
  order by a.id
)
select coalesce(json_agg(rows order by id), '[]'::json)::text from rows;
'''
    accounts = json.loads(psql(sql).strip() or "[]")
    if LIMIT is not None:
        accounts = accounts[:LIMIT]
    return accounts


def load_admin_key():
    key = psql("select value from settings where key='admin_api_key';").strip()
    if not key:
        raise RuntimeError("settings.admin_api_key is empty or missing")
    return key


def test_account(account, admin_key):
    payload = json.dumps({{"model_id": MODEL, "prompt": "请用一句话回复：pong", "mode": "default"}}).encode("utf-8")
    headers = {{"Content-Type": "application/json", "Accept": "text/event-stream", "x-api-key": admin_key}}
    url = f"{{BASE_URL}}/api/v1/admin/accounts/{{account['id']}}/test"
    start = time.time()
    status = None
    text = ""
    error = ""

    try:
        req = urllib.request.Request(url, data=payload, headers=headers, method="POST")
        with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
            status = resp.status
            body = resp.read().decode("utf-8", errors="replace")
        parsed = parse_test_sse(body)
        text = parsed["text"]
        error = parsed["error"]
        ok = status == 200 and parsed["complete"] and not error
        if status != 200 and not error:
            error = f"HTTP {{status}}"
        if status == 200 and not parsed["complete"] and not error:
            error = "SSE ended without test_complete"
    except urllib.error.HTTPError as exc:
        status = exc.code
        body = exc.read().decode("utf-8", errors="replace")
        parsed = parse_test_sse(body)
        text = parsed["text"]
        error = parsed["error"] or f"HTTP {{exc.code}}: {{body[:300]}}"
        ok = False
    except Exception as exc:
        error = f"{{type(exc).__name__}}: {{exc}}"
        ok = False

    return {{
        **account,
        "result": "OK" if ok else "FAILED",
        "http_status": status,
        "latency_ms": int((time.time() - start) * 1000),
        "class": classify_error(error),
        "error": error[:1000],
        "text": text[:200],
    }}


def main():
    started = time.time()
    accounts = query_accounts()
    print(f"OpenAI accounts found: {{len(accounts)}}", flush=True)
    admin_key = load_admin_key()
    print(f"Admin key loaded: {{admin_key[:8]}}...{{admin_key[-4:]}}", flush=True)

    results = []
    for index, account in enumerate(accounts, 1):
        result = test_account(account, admin_key)
        results.append(result)
        error_preview = result["error"].replace("\\n", " ")[:180]
        print(
            f"[{{index:03d}}/{{len(accounts)}}] {{result['result']}} "
            f"id={{result['id']}} sched={{str(result['schedulable']).lower()}} "
            f"status={{result['status']}} http={{result['http_status']}} "
            f"{{result['latency_ms']}}ms class={{result['class']}} err={{error_preview}}",
            flush=True,
        )

    unauthorized = [item for item in results if item["result"] == "FAILED" and item["class"] == "401_unauthorized"]
    unauthorized_schedulable = [item for item in unauthorized if item["schedulable"]]
    failed = [item for item in results if item["result"] != "OK"]
    counts = Counter(item["class"] for item in failed)

    payload = {{
        "summary": {{
            "total": len(results),
            "ok": len(results) - len(failed),
            "failed": len(failed),
            "unauthorized_401": len(unauthorized),
            "unauthorized_401_schedulable": len(unauthorized_schedulable),
            "elapsed_seconds": int(time.time() - started),
            "failure_classes": dict(sorted(counts.items())),
        }},
        "unauthorized_401": unauthorized,
        "unauthorized_401_schedulable": unauthorized_schedulable,
        "results": results,
    }}

    print("\\n=== SUMMARY ===")
    print(json.dumps(payload["summary"], ensure_ascii=False, sort_keys=True))
    print("\\n=== 401_ACCOUNTS ===")
    if not unauthorized:
        print("(none)")
    for item in unauthorized:
        print(
            f"id={{item['id']}} name={{item['name']}} status={{item['status']}} "
            f"schedulable={{item['schedulable']}} groups={{item['groups']}} "
            f"err={{item['error'].replace(chr(10), ' ')[:260]}}"
        )

    print("\\n=== 401_SCHEDULABLE_TRUE ===")
    if not unauthorized_schedulable:
        print("(none)")
    for item in unauthorized_schedulable:
        print(
            f"id={{item['id']}} name={{item['name']}} status={{item['status']}} "
            f"schedulable={{item['schedulable']}} groups={{item['groups']}} "
            f"err={{item['error'].replace(chr(10), ' ')[:260]}}"
        )

    print("\\nJSON_RESULT_BEGIN")
    print(json.dumps(payload, ensure_ascii=False))
    print("JSON_RESULT_END")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"FATAL: {{type(exc).__name__}}: {{exc}}", file=sys.stderr)
        sys.exit(1)
"""


def mask_secret(value: str) -> str:
    if not value:
        return ""
    if len(value) <= 12:
        return "***"
    return f"{value[:8]}...{value[-4:]}"


def build_expect_script(host: str, user: str, password: str, remote_python: str) -> str:
    encoded = base64.b64encode(remote_python.encode("utf-8")).decode("ascii")
    remote_command = (
        "python3 -c "
        + json.dumps(
            "import base64;"
            f"code=base64.b64decode('{encoded}').decode('utf-8');"
            "exec(compile(code, '<remote-sub2api-openai-401-check>', 'exec'))"
        )
    )
    ssh_target = f"{user}@{host}"
    return textwrap.dedent(
        f"""
        set timeout -1
        log_user 1
        spawn -noecho ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/tmp/sub2api_openai_401_known_hosts {ssh_target} {json.dumps(remote_command)}
        expect {{
          -re "(?i)password:" {{
            send -- {json.dumps(password + chr(13))}
            exp_continue
          }}
          eof
        }}
        catch wait result
        exit [lindex $result 3]
        """
    ).strip()


def extract_remote_json(output: str) -> dict[str, Any] | None:
    start_marker = "JSON_RESULT_BEGIN"
    end_marker = "JSON_RESULT_END"
    start = output.find(start_marker)
    end = output.find(end_marker)
    if start == -1 or end == -1 or end <= start:
        return None
    raw = output[start + len(start_marker) : end].strip()
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return None


def strip_json_payload_from_output(output: str) -> str:
    start_marker = "JSON_RESULT_BEGIN"
    end_marker = "JSON_RESULT_END"
    start = output.find(start_marker)
    end = output.find(end_marker)
    if start == -1 or end == -1 or end <= start:
        return output
    return (output[:start] + output[end + len(end_marker) :]).strip() + "\n"


def run_remote(args: argparse.Namespace) -> tuple[int, str, dict[str, Any] | None]:
    if shutil.which("expect") is None:
        raise RuntimeError("expect is required but was not found in PATH")

    password = args.password or os.environ.get("SUB2API_SSH_PASSWORD", "")
    if not password:
        raise RuntimeError("SSH password is required via --password or SUB2API_SSH_PASSWORD")

    remote_python = build_remote_python(
        model=args.model,
        timeout=args.timeout,
        limit=args.limit,
        base_url=args.remote_base_url,
    )
    expect_script = build_expect_script(args.host, args.user, password, remote_python)
    process = subprocess.run(
        ["expect", "-c", expect_script],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        check=False,
    )
    output = process.stdout
    parsed = extract_remote_json(output)
    return process.returncode, output, parsed


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="SSH to a remote sub2api server, test all OpenAI accounts, and list 401 accounts."
    )
    parser.add_argument("--host", default=DEFAULT_HOST)
    parser.add_argument("--user", default=DEFAULT_USER)
    parser.add_argument("--password", default="")
    parser.add_argument("--model", default=DEFAULT_MODEL)
    parser.add_argument("--timeout", type=int, default=DEFAULT_TIMEOUT)
    parser.add_argument("--limit", type=int, default=None, help="Test only the first N accounts.")
    parser.add_argument("--remote-base-url", default=DEFAULT_REMOTE_BASE_URL)
    parser.add_argument("--output-json", default="", help="Write full parsed result JSON to this local path.")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    returncode, output, parsed = run_remote(args)
    print(strip_json_payload_from_output(output), end="")

    if args.output_json:
        if parsed is None:
            print("Unable to find JSON result in remote output; not writing --output-json.", file=sys.stderr)
        else:
            output_path = Path(args.output_json)
            output_path.parent.mkdir(parents=True, exist_ok=True)
            output_path.write_text(json.dumps(parsed, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            print(f"Wrote JSON result to {output_path}")

    return returncode


if __name__ == "__main__":
    raise SystemExit(main())
