# Chatbox 接入中转站 API 教程

适合下面这些场景：

- 不想折腾命令行
- 想用图形界面聊天
- 想尽量少改配置

Chatbox 通常是比较适合首次接入的一种方式。

因为它本身就支持：

- 自定义模型服务商
- OpenAI API compatible
- 手动填写 `API Host`、`API Key`、`Model`

## 开始前先准备这 3 个值

- `Base URL / API Host`：`https://token.gepinkeji.com/v1`
- `API Key`：接口密钥，例如 `sk-xxxx`
- `Model`：建议先用 `GPT-5.4`

先提醒一句：

Chatbox 界面里有时叫：

- `API Host`
- `API Base URL`
- `API 地址`

它本质上大多都在指这里的 `Base URL`。

### 如果不清楚真实模型名的获取位置

最短路径先按这个顺序找：

1. 登录中转站后台
2. 找 `模型列表`、`允许模型`、`可用模型`
3. 复制真正用于请求的 `model id`

如果后台只给了展示名称，没有给真实模型名，不建议猜测。

## 官方链接

- Chatbox 官网：https://chatboxai.app/
- Chatbox 下载页：https://chatboxai.app/zh
- Chatbox Provider 文档：https://docs.chatboxai.app/en/guides/providers

## 第 1 步：安装并打开 Chatbox

先安装 Chatbox 客户端。

打开后，不要急着开始聊天，先去设置页。

按照 Chatbox 官方文档的常见路径，优先这样找：

1. 打开 Chatbox
2. 点击左下角 `Settings`
3. 进入 `Model Provider`

如果当前界面是中文，也可能看到：

- `设置`
- `模型服务商`

## 第 2 步：新增一个自定义 Provider

进入 `Model Provider` 后，按这个顺序走：

1. 点击 `Add`
2. 给这个 provider 起一个名字，比如 `GPAPI`
3. 类型选择 `OpenAI API compatible`

如果当前版本没有看到完全一样的文字，就找这些接近的选项：

- `OpenAI`
- `OpenAI API`
- `OpenAI Compatible`
- `OpenAI API compatible`

## 第 3 步：填写 API 信息

这一步直接按下面填：

| 字段 | 建议填写 |
| --- | --- |
| Provider Name | `GPAPI` |
| API Host | `https://token.gepinkeji.com/v1` |
| API Key | 真实密钥 |
| Model | `GPT-5.4` |

### 最容易卡住的一点

有些版本会把“路径”单独拆开。

这时可以先这样理解：

- 如果它只让填写 `API Host`，优先直接填完整的 `https://token.gepinkeji.com/v1`
- 如果它已经默认拼 `/v1/chat/completions`，不要再额外重复拼一遍

### 两种最常见界面怎么填

#### 情况 1：只有一个 `API Host` 输入框

直接填：

```text
https://token.gepinkeji.com/v1
```

#### 情况 2：`Host` 和 `Path` 分成两个框

可以先按这个思路填：

```text
Host:
https://token.gepinkeji.com

Path:
/v1
```

如果当前版本已经在内部固定拼了 `/v1/chat/completions`，不要再把完整路径重复塞进 `Host`。

## 第 4 步：至少新增 1 个模型

按 Chatbox 官方文档的通用配置流程，保存前一般要至少加 1 个模型。

所以如果界面里有“添加模型”这一步，请至少加一个：

```text
GPT-5.4
```

如果中转后台给出的真实模型名不是这个，就以后台为准。

## 第 5 步：点击 Check 做第一次验证

保存前后，Chatbox 常见都会有一个：

```text
Check
```

或者：

```text
检查连接
```

的按钮。

先点击这个按钮。

如果显示连接成功，说明下面这些值至少已经基本没填错：

- `API Host`
- `API Key`
- `Model`

## 第 6 步：回到首页发一条最简单的测试消息

创建一个新对话，然后直接发：

```text
请只回复：Chatbox 已连接成功
```

如果能正常回复，说明 Chatbox 已经接通。

### 如果 `Check` 失败，先按这个顺序看

#### 1. 提示 401 或鉴权失败

优先查：

1. `API Key` 是否完整
2. 是否多复制了空格
3. 这把 Key 是否启用

#### 2. 提示 404 或接口不存在

优先查：

1. `API Host` 是否写错
2. 是否重复拼了 `/v1` 或 `/chat/completions`
3. `Host` 和 `Path` 是否填反了

#### 3. 提示模型不存在

优先查：

1. 填写的是不是展示名称
2. 当前模型名是否和后台完全一致
3. 当前 Key 是否有这个模型权限

## 常见问题

### 1. 为什么地址已经填写，还是连不上

最常见的原因是：

1. 官网地址被填进去了
2. 少写了 `/v1`
3. 两个地方重复拼了接口路径

### 2. 为什么 Check 能过，但聊天还是不对

优先查这 3 个地方：

1. 当前对话是不是已经切到了刚添加的 provider
2. 当前模型是不是刚填写的模型
3. Chatbox 有没有缓存旧配置

如果怀疑缓存了旧配置，最简单的处理方式如下：

1. 关闭 Chatbox
2. 重新打开
3. 新建一个全新的聊天窗口再测

### 3. 模型不存在怎么办

优先检查：

1. 是不是把展示名称当成真实模型名了
2. 这个 Key 当前有没有这个模型权限
3. 模型名大小写、连接符、版本号有没有抄错

### 4. API Key 正确，为什么还提示无权限

这时不要先怀疑 Chatbox，先回到中转站后台看：

- Key 是否启用
- Key 是否还有额度
- Key 是否已分组
- 当前分组是否允许这个模型

### 5. Chatbox 最适合怎么接

最推荐的小白顺序是：

1. 先新增一个 `OpenAI API compatible` provider
2. 先只填一组最简单配置
3. 先点 `Check`
4. 再开新对话测试

不要一上来就同时开多个 provider、多组模型、多套高级参数。
