# OpenAI 401 账号检测脚本

本文面向 Sub2API 运维人员，说明如何使用 `tools/check_openai_401_accounts.py` 批量检测远程服务器上的 OpenAI 账号，并列出测试返回 401 的账号。

该脚本适用于排查 OpenAI OAuth 账号失效、`token_revoked`、`token_invalidated` 等问题，尤其是定位仍然 `schedulable=true`、但实际请求会返回 401 的账号。

---

## 功能说明

脚本在本机运行，自动完成以下步骤：

1. 通过 SSH 登录远程服务器。
2. 在远程服务器内读取 PostgreSQL 中所有未删除的 `platform='openai'` 账号。
3. 在远程服务器内请求 Sub2API 本机管理接口逐个测试账号：

   ```http
   POST http://127.0.0.1:8080/api/v1/admin/accounts/{id}/test
   ```

4. 解析测试接口返回的 SSE 事件，汇总所有返回 401 的账号。
5. 单独列出 `schedulable=true` 的 401 账号，方便判断哪些坏号仍可能被调度命中。

脚本不会打印完整的用户 API Key、admin API Key、access token 或 refresh token。

---

## 前置条件

本机需要：

- Python 3
- `expect`
- 能通过 SSH 访问目标服务器

远程服务器需要：

- `sub2api` 容器运行中
- `sub2api-postgres` 容器运行中
- 数据库存在 `settings.admin_api_key`
- 管理接口可从服务器本机访问：`http://127.0.0.1:8080`

---

## 基本用法

推荐通过环境变量提供 SSH 密码，避免密码出现在 shell history 中：

```bash
SUB2API_SSH_PASSWORD='your-password' python3 tools/check_openai_401_accounts.py
```

默认连接：

- host：`101.96.208.132`
- user：`root`
- model：`gpt-5.5`
- 单账号请求超时：`35` 秒

也可以通过参数显式指定：

```bash
python3 tools/check_openai_401_accounts.py \
  --host 101.96.208.132 \
  --user root \
  --password 'your-password' \
  --model gpt-5.5 \
  --timeout 35
```

---

## 小样本验证

上线或首次使用前，建议先测试少量账号：

```bash
SUB2API_SSH_PASSWORD='your-password' python3 tools/check_openai_401_accounts.py --limit 3
```

`--limit 3` 表示只测试数据库查询结果中的前 3 个 OpenAI 账号，用于验证 SSH、数据库查询和管理接口调用链路。

---

## 输出说明

脚本会逐账号输出测试结果，例如：

```text
[001/98] OK id=212 sched=true status=active http=200 3198ms class= err=
[002/98] FAILED id=512 sched=true status=error http=200 5677ms class=401_unauthorized err=API returned 401: ...
```

字段含义：

| 字段 | 说明 |
|------|------|
| `OK` / `FAILED` | 账号测试是否成功 |
| `id` | 账号 ID |
| `sched` | 数据库中的 `schedulable` 状态 |
| `status` | 数据库中的账号状态 |
| `http` | 管理测试接口本身返回的 HTTP 状态 |
| `class` | 脚本归类后的错误类型 |
| `err` | 上游错误摘要 |

汇总区示例：

```text
=== SUMMARY ===
{"total": 98, "ok": 52, "failed": 46, "unauthorized_401": 39, "unauthorized_401_schedulable": 1}

=== 401_ACCOUNTS ===
id=512 name=... status=error schedulable=True groups=2 err=API returned 401: ...

=== 401_SCHEDULABLE_TRUE ===
id=512 name=... status=error schedulable=True groups=2 err=API returned 401: ...
```

重点关注：

- `401_ACCOUNTS`：所有测试返回 401 的 OpenAI 账号。
- `401_SCHEDULABLE_TRUE`：仍可调度但测试返回 401 的账号，这类账号最容易导致线上请求命中异常账号。

---

## 保存 JSON 结果

如需保存完整结果：

```bash
SUB2API_SSH_PASSWORD='your-password' python3 tools/check_openai_401_accounts.py \
  --output-json /tmp/openai-401-check.json
```

JSON 中包含：

- `summary`
- `unauthorized_401`
- `unauthorized_401_schedulable`
- `results`

---

## 参数列表

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--host` | `101.96.208.132` | 远程服务器 IP 或域名 |
| `--user` | `root` | SSH 用户 |
| `--password` | 空 | SSH 密码；也可使用 `SUB2API_SSH_PASSWORD` |
| `--model` | `gpt-5.5` | 账号测试使用的模型 |
| `--timeout` | `35` | 单个账号测试超时时间，单位秒 |
| `--limit` | 空 | 只测试前 N 个账号，适合 dry run |
| `--remote-base-url` | `http://127.0.0.1:8080` | 远程服务器内访问 Sub2API 的地址 |
| `--output-json` | 空 | 将完整检测结果写入本地 JSON 文件 |

---

## 注意事项

- 脚本会真实请求上游账号，可能消耗少量额度。
- 管理测试接口可能更新账号测试、错误或探测相关字段。
- 脚本默认串行测试，避免给远程服务和上游造成额外压力。
- 脚本只列出问题账号，不会自动修改 `status` 或 `schedulable`。
- 如果输出中出现大量 `429_rate_limit`，说明账号未必失效，但额度、套餐或限速已经触发。

---

## 常见处理建议

发现 `401_SCHEDULABLE_TRUE` 后，建议先人工确认，再通过管理后台或管理 API 将对应账号停止调度，例如设置 `schedulable=false` 或修复/重新授权该账号。

发现 `status=error` 但 `schedulable=true` 时，应优先处理。这种状态组合表示账号已经被标记异常，但仍可能进入调度候选。
