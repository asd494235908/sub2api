# Casdoor 统一登录接入示例

本文给其他产品接入 `https://login.gepinkeji.com` Casdoor 统一登录时参考。示例产品域名使用 `https://product.example.com`，实际接入时替换为自己的产品域名、应用名、Client ID 和 Client Secret。

## 一、接入目标

Casdoor 只负责用户认证。用户在 Casdoor 登录成功后，产品后端拿到 Casdoor 用户信息，再创建或绑定产品自己的本地用户，并签发产品自己的登录态。

推荐原则：

- 产品自己的登录态仍由产品后端维护。
- 不把产品 JWT、session token 放进 URL。
- 不把 Casdoor access token 保存到前端。
- `CASDOOR_CLIENT_SECRET` 只放在后端配置中，不能写入前端。

## 二、Casdoor 应用配置示例

在 Casdoor 后台为每个产品创建独立应用。

示例：

| 配置项 | 示例值 |
| --- | --- |
| Casdoor 地址 | `https://login.gepinkeji.com` |
| 应用名称 | `product-example-com` |
| Client ID | `<Casdoor 分配的 Client ID>` |
| Client Secret | `<Casdoor 应用密钥>` |
| Redirect URI | `https://product.example.com/api/auth/casdoor/callback` |
| Scope | `openid profile email phone` |

Casdoor 应用里必须配置产品后端回调地址：

```text
https://product.example.com/api/auth/casdoor/callback
```

如果产品有多个环境，建议每个环境使用独立 Casdoor 应用，例如：

```text
product-example-com-prod
product-example-com-staging
product-example-com-dev
```

## 三、产品后端配置示例

```bash
CASDOOR_ENABLED=true
CASDOOR_ISSUER=https://login.gepinkeji.com
CASDOOR_CLIENT_ID=<Casdoor 应用 Client ID>
CASDOOR_CLIENT_SECRET=<Casdoor 应用 Client Secret>
CASDOOR_REDIRECT_URI=https://product.example.com/api/auth/casdoor/callback
CASDOOR_SCOPE="openid profile email phone"
CASDOOR_AUTHORIZE_URL=https://login.gepinkeji.com/login/oauth/authorize
CASDOOR_TOKEN_URL=https://login.gepinkeji.com/api/login/oauth/access_token
CASDOOR_USERINFO_URL=https://login.gepinkeji.com/api/userinfo
```

说明：

- `CASDOOR_REDIRECT_URI` 必须和 Casdoor 后台应用配置完全一致。
- 需要手机号时，`CASDOOR_SCOPE` 必须包含 `phone`。
- 修改后端环境变量后，需要重启或重新部署产品后端。

## 四、推荐登录流程

1. 用户在产品登录页点击 `Casdoor 登录`。
2. 前端跳转到产品后端登录入口：
   ```text
   GET /api/auth/casdoor/login?redirect=/dashboard
   ```
3. 产品后端生成 `state`，写入 HttpOnly cookie。
4. 产品后端 302 跳转到 Casdoor 授权地址。
5. 用户在 Casdoor 登录。
6. Casdoor 回调产品后端：
   ```text
   GET /api/auth/casdoor/callback?code=...&state=...
   ```
7. 产品后端校验 `state`。
8. 产品后端用 `code` 换 Casdoor access token。
9. 产品后端调用 Casdoor `/api/userinfo` 获取用户信息。
10. 产品后端绑定或创建本地用户。
11. 产品后端签发产品自己的登录态。
12. 前端进入产品页面。

## 五、授权地址示例

产品后端跳转到 Casdoor 的 URL 示例：

```text
https://login.gepinkeji.com/login/oauth/authorize?client_id=<CLIENT_ID>&response_type=code&redirect_uri=https%3A%2F%2Fproduct.example.com%2Fapi%2Fauth%2Fcasdoor%2Fcallback&scope=openid%20profile%20email%20phone&state=<STATE>
```

参数说明：

| 参数 | 说明 |
| --- | --- |
| `client_id` | Casdoor 应用 Client ID |
| `response_type` | 固定为 `code` |
| `redirect_uri` | 产品后端回调地址 |
| `scope` | 建议 `openid profile email phone` |
| `state` | 产品后端生成的随机字符串，用于防 CSRF |

## 六、后端接口示例

### 1. 登录入口

```http
GET /api/auth/casdoor/login
```

后端处理：

- 生成随机 `state`。
- 保存 `state` 到 HttpOnly cookie 或服务端 session。
- 读取可选 `redirect`，只允许站内路径。
- 302 跳转到 Casdoor 授权地址。

### 2. OAuth 回调

```http
GET /api/auth/casdoor/callback
```

后端处理：

- 校验 URL 参数 `state` 和登录入口保存的 `state` 是否一致。
- 读取 URL 参数 `code`。
- 调用 Casdoor token 接口换取 access token。
- 调用 Casdoor userinfo 接口获取用户信息。
- 按 Casdoor 用户唯一 ID 绑定本地用户。
- 如果未绑定，可按 email 或 phone 匹配已有本地用户。
- 如果仍找不到用户，按产品自己的注册策略创建用户或拒绝登录。

### 3. Token 交换请求

```http
POST https://login.gepinkeji.com/api/login/oauth/access_token
Content-Type: application/x-www-form-urlencoded
```

请求参数：

```text
grant_type=authorization_code
client_id=<CLIENT_ID>
client_secret=<CLIENT_SECRET>
code=<CODE>
redirect_uri=https://product.example.com/api/auth/casdoor/callback
```

### 4. 获取用户信息

```http
GET https://login.gepinkeji.com/api/userinfo
Authorization: Bearer <CASDOOR_ACCESS_TOKEN>
```

常用字段示例：

```json
{
  "sub": "casdoor-user-id",
  "name": "zhangsan",
  "email": "zhangsan@example.com",
  "phone_number": "13800138000"
}
```

字段兼容建议：

- 用户唯一 ID：优先使用 `sub`。
- 邮箱：读取 `email`。
- 手机号：建议兼容 `phone`、`phone_number`、`phoneNumber`、`mobile`。

## 七、前端页面示例

登录页按钮：

```html
<a href="/api/auth/casdoor/login?redirect=/dashboard">
  Casdoor 登录
</a>
```

如果产品后端直接在 callback 中写入产品登录态，可以登录成功后直接跳转到目标页面。

如果产品使用一次性 ticket 方案，推荐流程是：

```text
callback -> /auth/casdoor/success?ticket=...&redirect=/dashboard
```

前端成功页负责：

- 读取 `ticket`。
- 调用产品后端 ticket 兑换接口。
- 保存产品自己的登录态。
- 跳转到 `redirect` 指定的站内页面。

## 八、用户绑定建议

推荐绑定优先级：

1. 优先使用 Casdoor `sub` 绑定本地用户。
2. 如果没有绑定记录，再用 Casdoor `email` 匹配本地用户。
3. 如果仍没有匹配，再用 Casdoor `phone` 或 `phone_number` 匹配本地用户。
4. 如果 email 和 phone 命中不同本地用户，应拒绝登录，避免误绑。
5. 如果创建新用户，密码可随机生成或标记为第三方登录用户，不保存 Casdoor 密码。

推荐保存一张第三方身份绑定表：

```text
provider = "casdoor"
issuer = "https://login.gepinkeji.com"
subject = Casdoor sub
user_id = 产品本地用户 ID
```

## 九、安全要求

- `state` 必须随机生成，并在回调时校验。
- `redirect` 只能允许站内路径，不能允许外部 URL。
- `client_secret` 只能保存在后端。
- Casdoor access token 不返回前端。
- 产品 JWT 或 session token 不放入 URL。
- 建议登录过程中的一次性 ticket 只能使用一次，并设置较短有效期。
- 日志中不要打印完整手机号、access token、refresh token、client secret。

## 十、验收检查

### 1. 检查授权跳转

```bash
curl -sS -D - -o /dev/null \
  "https://product.example.com/api/auth/casdoor/login?redirect=/dashboard" \
  | grep -i location
```

确认 `Location` 中包含：

```text
https://login.gepinkeji.com/login/oauth/authorize
client_id=<CLIENT_ID>
redirect_uri=https%3A%2F%2Fproduct.example.com%2Fapi%2Fauth%2Fcasdoor%2Fcallback
scope=openid%20profile%20email%20phone
state=
```

### 2. 浏览器完整验证

1. 打开产品登录页。
2. 点击 `Casdoor 登录`。
3. 在 Casdoor 页面使用密码或验证码登录。
4. 登录成功后回到产品页面。
5. 产品后端当前用户接口能返回登录用户。
6. 如果需要手机号，确认当前用户信息里已经保存或展示手机号。

## 十一、常见问题

### redirect_uri 不匹配

Casdoor 后台配置、产品后端环境变量、实际授权 URL 中的 `redirect_uri` 必须完全一致。

### 登录后没有手机号

优先检查授权 URL 的 `scope` 是否包含 `phone`。如果没有包含，Casdoor 可能不会在 `/api/userinfo` 返回手机号。

同时确认产品后端兼容读取：

```text
phone
phone_number
phoneNumber
mobile
```

### 登录后回调失败

重点检查：

- `state` cookie 或服务端 session 是否丢失。
- 回调地址是否经过代理改写。
- token 接口是否能用 `code` 换到 access token。
- userinfo 接口是否能拿到 `sub`。

### 用户绑定冲突

如果同一个 Casdoor 用户的 email 和 phone 分别命中两个本地账号，不要自动合并，应提示用户联系管理员或走人工处理流程。
