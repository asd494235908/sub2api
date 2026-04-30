# 格品API 接入总入口
---

## 先看 30 秒版本

接入任何客户端，本质上都只是在填 3 个值：

- `Base URL`
- `API Key`
- `Model`

最常见的正确填写方式是：

```text
Provider: OpenAI 或 OpenAI Compatible
Base URL: https://接口域名/v1
API Key: sk-xxxxxxxxxxxxxxxx
Model: 真实模型名
```

如果不清楚这 3 个值从哪里来，先看下面这节。

## 一、接入前准备

开始前，至少要准备下面这 3 个信息：

| 项目 | 说明 | 示例 |
| --- | --- | --- |
| `Base URL` | 中转接口地址，不是官网首页地址 | `https://api.example.com/v1` |
| `API Key` | 访问密钥，通常以 `sk-` 开头 | `sk-xxxxxxxxxxxxxxxx` |
| `Model` | 账号实际可用的模型名 | `gpt-4o`、`GPT-5.4` |

可以先把它理解成：

- `Base URL` 决定“请求发到哪里”
- `API Key` 决定“请求使用哪把密钥”
- `Model` 决定“请求调用哪个模型”

### 一定要记住的 3 句话

1. `Base URL` 填接口地址，不要填平台官网地址
2. `Model` 要按后台真实模型名填写，不要自己猜
3. 大多数 OpenAI 兼容接口，地址通常要写到 `/v1`

### 如果接的是 格品API

最短路径通常是：

1. 登录 格品API 控制台
2. 打开左侧菜单 `API 密钥`
3. 创建或复制 `API Key`
4. 在同一页查看或复制 `API 端点`
5. 在后台确认当前账号可用的模型

如果其它子教程里要求填写：

- `API Host`
- `Endpoint`
- `Server URL`

本质上大多都还是在填 `Base URL`。

## 二、先选需要查看的教程

下面这些是当前已经拆开的子文档。

### 推荐先看哪一篇

| 客户端 | 建议去看 |
| --- | --- |
| Codex / Codex CLI / VS Code Codex | [Codex.md](./Codex.md) |
| Chatbox | [chatbox.md](./chatbox.md) |
| OpenCode | [OpenCode.md](./OpenCode.md) |
| OpenClaw | [OpenClaw.md](./OpenClaw.md) |
| Cursor | [Curso.md](./Curso.md) |
| Trae | [Trae.md](./Trae.md) |
| 还没装 Node.js，但教程里要执行 `npm install` | [node.js.md](./node.js.md) |

### 各文档适合什么人

#### 1. Codex / Codex CLI / VS Code Codex

看这里：[Codex.md](./Codex.md)

适合：

- 想把 Codex 全家桶接到同一个中转配置上
- 能接受先改一次 `~/.codex/config.toml`
- 想先配通 CLI，再复用到桌面版和 VS Code 扩展

#### 2. Chatbox

看这里：[chatbox.md](./chatbox.md)

适合：

- 想用图形界面
- 希望尽量通过设置页完成接入
- 想先用 `Check` 按钮做最低成本验证

#### 3. OpenCode

看这里：[OpenCode.md](./OpenCode.md)

适合：

- 想用本地 CLI 工具
- 能接受写配置文件
- 想要一份已经带示例配置和截图的教程

#### 4. OpenClaw

看这里：[OpenClaw.md](./OpenClaw.md)

适合：

- 更喜欢向导式配置
- 想通过 `Custom Provider` 接入
- 希望步骤更偏“照着点就行”

#### 5. Cursor

看这里：[Curso.md](./Curso.md)

适合：

- 已经在用 Cursor
- 想先确认当前版本有没有公开的自定义 `Base URL` 入口
- 愿意按文档先验证“能不能接”，再决定要不要继续

#### 6. Trae

看这里：[Trae.md](./Trae.md)

适合：

- 已经在用 Trae
- 想确认当前公开版本能不能接第三方 OpenAI 兼容接口
- 更需要一份“能配就配，不能配就及时止损”的教程

#### 7. Node.js

看这里：[node.js.md](./node.js.md)

如果在其它文档里看到了这些命令：

- `npm install -g ...`
- `node --version`
- `npm --version`

那就先把 Node.js 装好，再继续后面的客户端教程。

## 三、没有单独教程时，先按通用规则填

如果当前接的是：

- VS Code 里的其它插件
- 还没有独立子文档的客户端
- 其它支持 OpenAI 兼容接口的工具

可以先按这套通用规则来：

| 客户端里常见名字 | 应填写内容 |
| --- | --- |
| `Provider` | `OpenAI` 或 `OpenAI Compatible` |
| `Base URL` / `API Host` / `Endpoint` / `Server URL` | 中转地址，通常到 `/v1` |
| `API Key` / `Access Token` / `Secret Key` | `API Key` |
| `Model` / `Model Name` / `Chat Model` | 真实模型名 |
| `Organization` / `Project` | 一般留空，除非客户端强制要求 |

### 通用复制模板

```text
Provider: OpenAI Compatible
Base URL: https://api.example.com/v1
API Key: sk-xxxxxxxxxxxxxxxx
Model: gpt-4o
```

### 最常见的 3 个填写错误

1. 把 `Base URL` 错填成平台官网地址
2. 把展示名称当成真实模型名
3. 在地址里重复拼接 `/v1/chat/completions`

## 四、怎么判断自己该看哪篇

如果还在犹豫，可以直接按下面选择：

- 想最快接起来：先看 [OpenClaw.md](./OpenClaw.md)
- 想用 Codex 全家桶：看 [Codex.md](./Codex.md)
- 想用图形界面快速测试：看 [chatbox.md](./chatbox.md)
- 想用 CLI、能接受配文件：先看 [OpenCode.md](./OpenCode.md)
- 已经在用 Cursor：看 [Curso.md](./Curso.md)
- 已经在用 Trae：看 [Trae.md](./Trae.md)
- 被 `npm` / `node` 卡住：先看 [node.js.md](./node.js.md)

## 五、统一排错说明

不管接的是哪个客户端，失败时先按这个顺序检查。

### 1. 先查地址

- 填写的是中转接口地址，不是官网首页
- 地址通常建议写到 `/v1`
- 不要在多个输入框里重复拼接接口路径

### 2. 再查 Key

- 确认没有多复制空格和换行
- 确认该 Key 没有被禁用
- 确认该 Key 还有额度或权限

### 3. 再查模型名

- 必须和后台开放名称完全一致
- 注意大小写、连字符、版本号
- 优先复制后台原文，不要手改

### 4. 再查客户端模式

- 优先选择 `OpenAI` 或 `OpenAI Compatible`
- 如果客户端支持多个 Provider，确认当前真的切到了刚配置的那个
- 如果客户端区分聊天、补全、嵌入，请先只测试聊天能力

### 5. 最后查服务端

- 域名是否能访问
- 证书是否正常
- 反向代理是否配置正确

### 常见报错速查

| 报错或现象 | 优先排查项 |
| --- | --- |
| `401 Unauthorized` | `API Key` 是否正确、是否复制了空格 |
| `404` 或接口不存在 | `Base URL` 是否写错、是否重复拼路径 |
| 模型不存在 | 模型名是否是真实可用名称 |
| 能连上但回复失败 | 当前 Key 是否有该模型权限 |
| 保存后没反应 | 客户端是否仍在使用旧配置、是否需要重启 |

## 六、推荐阅读顺序

如果是第一次接触这类接入，推荐按这个顺序阅读：

1. 先读这篇 `guide.md`
2. 如果需要安装运行环境，先读 [node.js.md](./node.js.md)
3. 再读对应客户端的子教程
4. 接入失败时，再回到这篇文档看“统一排错说明”

## 七、这篇主文档以后怎么维护

为了避免一处改了、另一处忘了改，后面建议这样维护：

- `guide.md` 只放总入口、通用规则、统一排错
- 每个客户端的实际步骤，单独维护在自己的子文档里
- 客户端截图、特殊限制、版本差异，也只写在对应子文档里

这样做的好处是：

1. 主文档更短，更适合新手先快速判断方向
2. 子文档可以按客户端单独更新，不容易互相影响
3. 遇到像 Trae 这种“公开支持路径不稳定”的情况，也能单独说明，不会把主文档写乱
