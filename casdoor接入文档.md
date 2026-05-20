# Token 产品接入 Casdoor 登录对接文档

## 一、接入目标

Token 产品通过 Casdoor 实现统一登录。

- Token 产品域名：`https://token.gepinkeji.com`
- Casdoor 登录域名：`https://login.gepinkeji.com`
- Casdoor 应用名称：`token-gepinkeji-com`
- Casdoor Client ID：`48ee9380353a95f3b463`
- OAuth 回调地址：`https://token.gepinkeji.com/api/auth/callback`
- 产品鉴权方式：继续使用 Token 产品自己的 JWT

Casdoor 只负责登录认证。用户登录成功后，Token 产品后端用 Casdoor 返回的用户信息生成 Token 产品自己的 JWT。

## 二、整体登录流程

1. 用户在 Token 产品点击登录。
2. 前端跳转到 Token 后端接口：`/api/auth/login`。
3. Token 后端生成 `state`，然后跳转到 Casdoor 登录页。
4. 用户在 Casdoor 使用密码或短信验证码登录。
5. Casdoor 登录成功后回调：`/api/auth/callback`。
6. Token 后端用 `code` 换 Casdoor access token。
7. Token 后端请求 Casdoor `/api/userinfo` 获取用户信息。
8. Token 后端绑定或创建本地用户。
9. Token 后端生成一次性 `login_ticket`。
10. 前端用 `login_ticket` 换取 Token 产品自己的 JWT。
11. 后续所有 Token 产品接口继续使用产品自己的 JWT。

## 三、后端接口

### 1. 登录入口

```http
GET /api/auth/login
```

用途：开始 Casdoor 登录流程。

后端处理逻辑：

- 生成随机 `state`
- 保存 `state` 到临时 cookie
- 读取可选参数 `redirect`
- 跳转到 Casdoor OAuth 授权地址

Casdoor 授权地址格式：

```text
https://login.gepinkeji.com/login/oauth/authorize?client_id=48ee9380353a95f3b463&response_type=code&redirect_uri=https%3A%2F%2Ftoken.gepinkeji.com%2Fapi%2Fauth%2Fcallback&scope=openid%20profile%20email&state=<state>
```

### 2. OAuth 回调

```http
GET /api/auth/callback
```

用途：接收 Casdoor 登录成功后的回调。

后端处理逻辑：

- 读取 URL 参数 `code`
- 读取 URL 参数 `state`
- 校验 `state` 是否和临时 cookie 一致
- 使用 `code` 换取 Casdoor access token
- 请求 Casdoor `/api/userinfo`
- 绑定或创建 Token 产品用户
- 创建一次性 `login_ticket`
- 跳转到前端成功页：

```text
https://token.gepinkeji.com/auth/success?ticket=<login_ticket>
```

注意：不要把产品 JWT 直接放到 URL 中。

### 3. 用 ticket 换 JWT

```http
POST /api/auth/exchange-ticket
Content-Type: application/json
```

请求：

```json
{
  "ticket": "一次性登录票据"
}
```

响应：

```json
{
  "token": "Token产品自己的JWT",
  "expires_in": 604800,
  "user": {
    "id": "产品用户ID",
    "name": "用户名",
    "email": "邮箱",
    "phone": "手机号"
  }
}
```

要求：

- `ticket` 只能使用一次
- `ticket` 有效期建议 2 分钟
- 兑换成功后立即删除 `ticket`

### 4. 当前用户信息

```http
GET /api/auth/me
Authorization: Bearer <产品JWT>
```

未登录返回：

```http
401 Unauthorized
```

已登录返回：

```json
{
  "id": "产品用户ID",
  "name": "用户名",
  "email": "邮箱",
  "phone": "手机号"
}
```

### 5. 退出登录

```http
POST /api/auth/logout
Authorization: Bearer <产品JWT>
```

如果 Token 产品是无状态 JWT，后端可以只返回成功，由前端删除本地 JWT。

如果产品已有 JWT 黑名单机制，则后端把当前 JWT 加入黑名单。

## 四、后端配置

Token 产品后端新增环境变量：

```bash
CASDOOR_ISSUER=https://login.gepinkeji.com
CASDOOR_CLIENT_ID=48ee9380353a95f3b463
CASDOOR_CLIENT_SECRET=<从 Casdoor token-gepinkeji-com 应用页面复制>
CASDOOR_REDIRECT_URI=https://token.gepinkeji.com/api/auth/callback
CASDOOR_SCOPE="openid profile email"
```

继续使用 Token 产品现有 JWT 配置：

```bash
JWT_SECRET=<Token产品现有JWT密钥>
JWT_EXPIRE_SECONDS=604800
```

`CASDOOR_CLIENT_SECRET` 只能放在后端环境变量中，不能写入前端代码。

## 五、Casdoor 接口

### 1. 用 code 换 token

```http
POST https://login.gepinkeji.com/api/login/oauth/access_token
Content-Type: application/x-www-form-urlencoded
```

参数：

```text
grant_type=authorization_code
client_id=48ee9380353a95f3b463
client_secret=<CASDOOR_CLIENT_SECRET>
code=<callback code>
redirect_uri=https://token.gepinkeji.com/api/auth/callback
```

### 2. 获取用户信息

```http
GET https://login.gepinkeji.com/api/userinfo
Authorization: Bearer <casdoor_access_token>
```

返回中重点使用字段：

```json
{
  "sub": "Casdoor用户唯一ID",
  "name": "用户名",
  "email": "邮箱",
  "phone": "手机号"
}
```

## 六、用户绑定规则

Token 产品用户表建议增加或确认存在以下字段：

```text
casdoor_sub
email
phone
```

绑定顺序：

1. 先按 `casdoor_sub` 查找用户。
2. 找不到时按 `email` 查找用户。
3. 仍找不到时按 `phone` 查找用户。
4. 如果匹配到旧用户，补写 `casdoor_sub`。
5. 如果都找不到，则创建 Token 产品新用户。

产品 JWT 建议包含：

```json
{
  "user_id": "产品用户ID",
  "casdoor_sub": "Casdoor用户唯一ID",
  "email": "用户邮箱",
  "phone": "用户手机号",
  "exp": 过期时间
}
```

不要保存 Casdoor 密码。不要把 Casdoor access token 当作 Token 产品 JWT 使用。

## 七、前端页面改造

### 1. 登录按钮

登录按钮点击后跳转：

```text
/api/auth/login
```

如果需要登录后回到当前页面：

```text
/api/auth/login?redirect=/当前页面路径
```

### 2. 新增登录成功页

新增前端页面：

```text
/auth/success
```

页面逻辑：

1. 从 URL 读取 `ticket`
2. 调用：

```http
POST /api/auth/exchange-ticket
```

3. 获取 Token 产品自己的 JWT
4. 保存到 Token 产品当前使用的位置，例如 `localStorage`
5. 跳转到首页或原访问页面

### 3. 页面初始化

页面加载时：

1. 从本地读取产品 JWT
2. 调用：

```http
GET /api/auth/me
Authorization: Bearer <产品JWT>
```

3. 如果返回 `200`，展示已登录状态
4. 如果返回 `401`，清除本地 JWT，展示登录按钮

### 4. 退出登录

退出按钮逻辑：

1. 调用：

```http
POST /api/auth/logout
```

2. 删除前端保存的产品 JWT
3. 跳转到首页或登录入口

## 八、安全要求

- 不把 Casdoor client secret 写入前端。
- 不把产品 JWT 放到 URL。
- 不把 Casdoor access token 保存到前端。
- `state` 必须校验，防止 CSRF。
- `login_ticket` 必须短时有效、一次性使用。
- 产品最终登录态继续使用自己的 JWT。
- HTTPS 域名下使用正式回调地址，不使用 localhost。

## 九、测试验收

### 正常流程

- 打开 `https://token.gepinkeji.com/api/auth/login`
- 应跳转到 Casdoor 登录页
- 使用密码登录成功
- 应回到 Token 产品
- 使用短信验证码登录成功
- 应回到 Token 产品
- `/auth/success?ticket=...` 能换到产品 JWT
- 带产品 JWT 请求 `/api/auth/me` 能返回当前用户
- 原有 Token 产品接口继续可用

### 异常流程

- callback 缺少 `code`，应拒绝登录
- `state` 不匹配，应拒绝登录
- `ticket` 重复使用，应失败
- `ticket` 过期，应失败
- Casdoor token exchange 失败时，不创建产品 JWT
- 未登录访问 `/api/auth/me`，应返回 `401`

## 十、不需要修改的部分

- 不需要修改 Nginx。
- 不需要修改 Casdoor 源码。
- 不需要修改 Casdoor 数据库。
- 不需要把 Casdoor access token 作为产品 JWT。
- 不需要使用 localhost 回调地址。
