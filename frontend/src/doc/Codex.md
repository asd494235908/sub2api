# Codex 接入中转站 API 教程

这篇文档把 3 种用法写在一起：

1. `Codex` 桌面版
2. `Codex CLI`
3. `VS Code Codex`

先看一个最重要的结论：

`Codex CLI`、`VS Code Codex`，以及 Codex app 里的大部分本地 agent 配置，核心都围绕同一个配置文件：

```text
~/.codex/config.toml
```

最省事的方式不是 3 个地方分别配置 3 次，而是：

1. 先把 `Codex CLI` 配通
2. 再让桌面版和 VS Code Codex 复用这套配置

## 开始前先准备这 3 个值

- `Base URL`：`https://token.gepinapi.com/v1`
- `API Key`：接口密钥，例如 `sk-xxxx`
- `Model`：建议先用 `GPT-5.4`

如果中转站已经明确给出其它模型名，以后台真实模型名为准。

### 如果不清楚真实模型名的获取位置

最短路径先按这个顺序找：

1. 登录中转站后台
2. 找 `模型列表`、`允许模型`、`可用模型`、`接口文档`
3. 复制其中真正用于请求的 `model id`

需要区分两种名字：

- 展示名称：给人看的，例如“GPT-5.4 高级版”
- 真实模型名：给客户端填的，例如 `GPT-5.4`

如果后台没有模型列表，直接查看接口文档或联系支持，不建议猜测模型名。

## 官方链接

- Codex 官方文档：https://developers.openai.com/codex/
- Codex CLI：https://developers.openai.com/codex/cli
- Codex Authentication：https://developers.openai.com/codex/auth
- Codex IDE Extension：https://developers.openai.com/codex/ide
- Codex IDE Settings：https://developers.openai.com/codex/ide/settings
- Codex App：https://developers.openai.com/codex/app
- Node.js 教程：[node.js.md](./node.js.md)

## 先理解这篇教程的配置思路

Codex 官方文档明确支持：

- 用 API Key 登录
- 用 `~/.codex/config.toml` 配模型和 provider
- CLI 与 IDE 扩展共享同一套登录与配置思路

但官方文档主要是围绕 OpenAI 自己的 API 与账号体系写的。

本文的目标是把这套配置结构改成指向中转站的 OpenAI 兼容地址。

这也是本文优先采用下面顺序的原因：

1. 先把 `Codex CLI` 配好
2. 再复用到 Codex app 和 VS Code Codex

## 第 1 步：如果还没安装 Node.js，先安装 Node.js

如果执行下面命令会报错：

```bash
node --version
npm --version
```

那就先去看：

- [node.js.md](./node.js.md)

确认 `node --version` 和 `npm --version` 都能正常显示版本号后，再继续下面的步骤。

## 第 2 步：安装 Codex CLI

打开终端，执行：

```bash
npm install -g @openai/codex
```

安装完成后，执行：

```bash
codex --version
```

如果能正常显示版本号，说明 `Codex CLI` 已经装好了。

## 第 3 步：先用 API Key 登录

最简单的做法是直接把 Key 通过 stdin 交给 Codex：

```bash
printf 'sk-请替换成真实密钥' | codex login --with-api-key
```

登录完成后，再执行：

```bash
codex login status
```

如果看到类似“Logged in using an API key”的结果，说明登录已经成功。

### 这一步要注意什么

1. 这里填的是中转站 `API Key`
2. 不要多复制空格和换行
3. 如果之前已经登录过别的 Key，可以重新执行一次覆盖

## 第 4 步：编辑 `~/.codex/config.toml`

这是整篇文档最关键的一步。

打开：

```text
~/.codex/config.toml
```

如果文件不存在，就手动创建。

第一次接入，最省事的写法可以直接参考下面这份：

```toml
model_provider = "gepinapi"
model = "gpt-5.4"
review_model = "gpt-5.4"

[model_providers.gepinapi]
name = "拓垦API"
base_url = "https://token.gepinapi.com/v1"
wire_api = "responses"
requires_openai_auth = true
```

### 这几行分别是什么意思

- `model_provider = "gepinapi"`
  说明默认走这里定义的 provider
- `model = "gpt-5.4"`
  说明默认使用这个模型
- `review_model = "gpt-5.4"`
  说明代码评审等场景也先走这个模型
- `base_url`
  指向中转地址
- `wire_api = "responses"`
  这是当前 Codex 配置里更稳妥的写法
- `requires_openai_auth = true`
  让这个 provider 继续使用前面 `codex login` 缓存下来的认证

### 最容易误解的两行

#### 1. `requires_openai_auth = true` 到底是什么意思

这里不是要求再次登录新的 OpenAI 账号。

可以直接理解成：

`继续使用执行 codex login --with-api-key 之后，本地已经缓存下来的那份认证。`

也就是说：

1. 前面那条 `codex login --with-api-key` 会产出可复用的认证
2. 这里会继续使用那次登录得到的认证头
3. 可以把它理解成“继续沿用前面录入的 Key 所生成的认证”

#### 2. `wire_api = "responses"` 为什么这样写

这是当前 Codex 配置里更稳的写法，而且当前版本已经不再推荐旧的 `chat` 写法。

如果这里报错，不建议立刻改成别的值，优先检查：

1. `base_url` 是否正确
2. 中转站是否真的兼容当前 Codex 所走的接口
3. `model` 是否是真实可用模型名

### 如果需要更换模型

只需要把：

```toml
model = "gpt-5.4"
review_model = "gpt-5.4"
```

改成真实模型名，例如：

```toml
model = "gpt-5.3-codex"
review_model = "gpt-5.3-codex"
```

前提是中转站后台确实支持这个模型。

## 第 5 步：先用 Codex CLI 做最短验证

进入任意一个测试目录，执行：

```bash
codex
```

然后给它发一句最简单的话：

```text
请只回复：Codex CLI 已连接成功
```

如果能正常回复，说明下面这整套链路至少已经通了：

- 登录
- `config.toml`
- `Base URL`
- `API Key`
- `Model`

### 如果 CLI 还没通，先不要急着配置另外两个

因为：

- Codex app
- VS Code Codex

后面基本都建立在这套配置之上。

最稳的顺序如下：

1. 先通 CLI
2. 再去看桌面版
3. 最后看 VS Code 扩展

### 如果 CLI 失败，先按这个顺序排查

#### 情况 1：`codex login status` 就不正常

先重做：

```bash
printf 'sk-请替换成真实密钥' | codex login --with-api-key
```

然后再执行：

```bash
codex login status
```

#### 情况 2：登录正常，但一发消息就报错

先检查：

1. `~/.codex/config.toml` 是否已经保存
2. `base_url` 是否写成了完整接口地址
3. `model_provider` 是否真的写成了 `gepinapi`
4. `model` 是否是真实模型名

#### 情况 3：提示 401 或鉴权失败

优先检查：

1. 登录时使用的是否是当前这把 Key
2. Key 有没有多复制空格或换行
3. 这把 Key 在后台是否还启用

#### 情况 4：提示 404、接口不存在、unknown endpoint

优先检查：

1. `base_url` 是否写错
2. 地址里有没有重复拼路径
3. 中转站是否兼容 Codex 当前走的接口格式

#### 情况 5：提示模型不存在

优先检查：

1. 填写的是否是展示名称
2. 当前 Key 是否有这个模型权限
3. 模型名大小写、连接符、版本号是否抄错

---

## 一、Codex 桌面版怎么用

### 第 1 步：打开 Codex app

如果还没安装桌面版，可以先执行：

```bash
codex app
```

或者直接从 Codex 官方入口打开桌面版。

### 第 2 步：如果 app 需要登录，优先继续用 API Key 方式

如果目标是把本地工作区接到中转站，最简单的方式仍然是：

1. 先用上面那套 `codex login --with-api-key`
2. 再让 app 读取同一套认证和配置

如果前面已经把 CLI 配通，最稳的动作如下：

1. 先完全退出 Codex app
2. 确认 `~/.codex/config.toml` 已经保存
3. 再重新打开 app

这样更容易让 app 读到最新配置。

### 第 3 步：确认 app 走的是刚才的配置

进入 app 后，重点确认两件事：

1. 当前打开的是本地工作区，而不是云端任务
2. 当前 agent 使用的是 `~/.codex/config.toml` 里配置的模型/provider

最简单的测试方式：

在本地项目里让它回复：

```text
请只回复：Codex App 已连接成功
```

### 这一步的实话提醒

如果使用的是 Codex Cloud 或依赖 ChatGPT 账户套餐的那套能力，不要把本文的“中转 API 接入”思路和 ChatGPT 账号计费思路混在一起。

最简单的理解方式如下：

- 本文主要覆盖“本地工作区 + API Key + 自定义 provider”
- 如果目标是使用 ChatGPT 套餐内置额度，则按官方账号体系处理

---

## 二、Codex CLI 怎么用

如果前面的验证已经成功，`Codex CLI` 实际上已经可以直接使用。

最常见的两种方式：

### 方式 A：直接进入交互模式

```bash
codex
```

### 方式 B：在当前项目目录里直接提需求

```bash
codex "请检查这个项目里有没有明显的接口配置问题"
```

### CLI 场景最常见的 3 个坑

1. 只登录了 API Key，但没有改 `config.toml`
2. `base_url` 写成官网首页地址，而不是接口地址
3. `wire_api` 不是 `responses`

---

## 三、VS Code Codex 怎么用

### 第 1 步：安装 VS Code Codex 扩展

在 VS Code 里打开扩展市场，搜索并安装官方 Codex 扩展。

### 第 2 步：优先复用前面已经配好的 CLI 配置

对小白来说，最稳的方式不是在 VS Code 里重新配一遍，而是直接复用这两样：

1. `codex login` 已经缓存好的登录状态
2. `~/.codex/config.toml`

如果前面的 `Codex CLI` 已经通了，VS Code Codex 通常也更容易一起通。

第一次接入时，建议按下面顺序处理：

1. 先完全退出 VS Code
2. 确认 `~/.codex/config.toml` 已经保存
3. 再重新打开 VS Code
4. 然后再打开 Codex 扩展去测试

### 第 3 步：在 VS Code 里做最小测试

打开一个项目后，唤起 Codex 扩展，然后发送：

```text
请只回复：VS Code Codex 已连接成功
```

如果可以正常回复，说明这套共享配置已经被扩展吃到了。

### 如果 VS Code 扩展没通，但 CLI 已经通了

先按这个顺序查：

1. VS Code 是否已经完全重启
2. 扩展是不是读取了当前用户下的 `~/.codex/config.toml`
3. 当前 VS Code 登录态是不是还是旧配置
4. 扩展当前使用的 provider/model 是否还是默认值

---

## 常见问题

### 1. 为什么这篇文档先讲 CLI，不先讲桌面版

因为 CLI 最容易验证。

只要 CLI 通了，后面的桌面版和 VS Code 扩展通常都更好判断问题出在哪。

### 2. 为什么已经登录成功了，还是调不通

因为“登录成功”只代表本地已经缓存了 Key。

它不代表：

- `Base URL` 已经写对
- `model_provider` 已经切到中转 provider
- `model` 已经写成可用模型名

### 3. `Base URL` 要不要写 `/v1`

大多数 OpenAI 兼容接口都建议写到 `/v1`。

如果后台已经给了完整地址，就以后台显示为准。

### 4. Codex CLI 提示接口错误，先查什么

先查这 4 项：

1. `codex login status` 是否正常
2. `~/.codex/config.toml` 是否已经保存
3. `base_url` 是否正确
4. `model` 是否是真实可用模型名

### 5. 哪一种方式更适合首次接入

推荐顺序：

1. 先配 `Codex CLI`
2. 再试 `VS Code Codex`
3. 最后再看 `Codex` 桌面版

这是最稳、最不容易迷路的顺序。
