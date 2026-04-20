# OpenClaw 接入 OpenAI 兼容 API 教程

如果你的目标是“把第三方 OpenAI 兼容接口接到本地工具里”，OpenClaw 是比较适合新手的一种方式，因为它官方文档明确支持：

- 安装向导
- Onboarding 向导
- Custom Provider
- `Base URL`、`API Key`、`Model ID`

你开始前准备好：

- `Base URL`：这里填 `https://token.gepinapi.com/`
- `API Key`：例如 `sk-xxxx`
- `Model`：例如 `GPT-5.4`

## 官方网站

- OpenClaw 官网：https://openclaw.ai/
- OpenClaw 文档：https://docs.openclaw.ai/
- OpenClaw 安装说明：https://docs.openclaw.ai/install
- Node.js 教程：[node.js.md](./node.js.md)

## 推荐你记住的最简单流程

1. 安装 OpenClaw
2. 运行 `openclaw onboard`
3. 选择 `Custom Provider`
4. 选择 `OpenAI-compatible`
5. 输入 `Base URL`、`API Key`、`Model ID`
6. 打开控制面板测试

安装方式上，这篇文档统一优先使用 Node.js + `npm`，这样三端步骤最一致；官方安装脚本保留为备用方案。

## 如果安装脚本失败，先装 Node.js 再重试

如果你不想用 npm，或者 npm 安装失败，可以再改用官方安装脚本。大多数情况下你不需要手动研究更多依赖。

但如果你遇到下面这些问题：

- 安装脚本中途失败
- 终端提示缺少运行时
- `openclaw` 命令装完后不能正常启动

可以先补装 Node.js `LTS` 版本，再重新执行 OpenClaw 安装命令。

请先看：

- [Node.js 安装教程](./node.js.md)

确认 `node --version` 和 `npm --version` 都能正常返回版本号后，再重新安装 OpenClaw。

---

## Windows

### 第 1 步：安装 OpenClaw

OpenClaw 官方说明，Windows 原生可用，但 `WSL2` 会更稳定。为了让步骤更统一，这篇文档优先使用 npm 安装。

打开 PowerShell，执行：

```bash
npm install -g openclaw@latest
```

安装完成后，执行：

```powershell
openclaw --version
```

如果能看到版本号，说明安装成功。

如果 npm 安装失败，再改用官方 PowerShell 安装脚本：

```powershell
iwr -useb https://openclaw.ai/install.ps1 | iex
```

### 第 2 步：运行初始化向导

```powershell
openclaw onboard --install-daemon
```

如果你不想装后台服务，也可以先用：

```powershell
openclaw onboard
```

### 第 3 步：在向导里选择自定义 Provider

当向导问你选哪个模型提供方时：

1. 选择 `Custom Provider`
2. 选择 `OpenAI-compatible`

然后按提示依次填写：

1. `Base URL`
2. `API Key`
3. `Model ID`
4. 可选别名

示例：

- `Base URL`：`https://token.gepinapi.com/`
- `API Key`：`sk-xxxx`
- `Model ID`：`GPT-5.4`

### 第 4 步：打开控制面板测试

执行：

```powershell
openclaw dashboard
```

浏览器打开后，输入一句测试话术：

```text
请只回复：OpenClaw Windows 接口已连接成功
```

如果有正常回复，说明已经接通。

---

## macOS

### 第 1 步：安装 OpenClaw

打开终端，执行：

```bash
npm install -g openclaw@latest
```

安装完成后确认版本：

```bash
openclaw --version
```

如果 npm 安装失败，再改用官方安装脚本：

```bash
curl -fsSL --proto '=https' --tlsv1.2 https://openclaw.ai/install.sh | bash
```

### 第 2 步：运行初始化向导

```bash
openclaw onboard --install-daemon
```

### 第 3 步：选择 Custom Provider

在向导里：

1. 选择 `Custom Provider`
2. 选择 `OpenAI-compatible`
3. 输入你的 `Base URL`
4. 输入你的 `API Key`
5. 输入你的 `Model ID`

### 第 4 步：打开 Dashboard 测试

```bash
openclaw dashboard
```

如果浏览器能正常打开控制界面，再发一条简单消息测试即可。

---

## Linux

### 第 1 步：安装 OpenClaw

Linux 上这篇文档优先使用 npm 安装：

```bash
npm install -g openclaw@latest
```

如果你只是想改用官方安装脚本，也可以执行：

```bash
curl -fsSL --proto '=https' --tlsv1.2 https://openclaw.ai/install.sh | bash
```

安装后确认版本：

```bash
openclaw --version
```

### 第 2 步：运行向导

```bash
openclaw onboard --install-daemon
```

### 第 3 步：选择兼容模式

按向导一步一步填：

1. 选择 `Custom Provider`
2. 兼容模式选 `OpenAI-compatible`
3. 填 `Base URL`
4. 填 `API Key`
5. 填 `Model ID`

### 第 4 步：打开网页控制面板

```bash
openclaw dashboard
```

打开后测试一句：

```text
请只回复：OpenClaw Linux 已连接成功
```

---

## 如果向导没配成功，可以手动写配置

这部分不是必须，只是给你留个备用方案。

OpenClaw 官方文档也支持通过自定义 provider 配置 `baseUrl`。

示例思路如下：

```json
{
  "models": {
    "providers": {
      "custom-proxy": {
        "baseUrl": "https://token.gepinapi.com/",
        "apiKey": "你的API_KEY",
        "api": "openai-completions",
        "models": [
          {
            "id": "GPT-5.4",
            "name": "GPT-5.4",
            "reasoning": false,
            "input": ["text"]
          }
        ]
      }
    }
  }
}
```

如果你是第一次接触 OpenClaw，建议先把向导跑通，再研究手动配置。

---

## 常见问题

### 1. Base URL 要不要带 `/v1`

大多数 OpenAI 兼容接口都需要。最稳妥的办法是直接看你服务商给你的示例地址。

### 2. 向导里没看到我要的服务商名字

没关系，直接选 `Custom Provider` 即可。

### 3. 为什么我接的是第三方接口，却要选 `OpenAI-compatible`

因为你接的不是“OpenAI 官方平台”，而是“兼容 OpenAI 接口格式”的服务。

### 4. 打开 dashboard 后没有回复

重点检查：

- `Base URL` 是否能在浏览器或终端访问
- `API Key` 是否正确
- `Model ID` 是否真实可用

## 官方参考

- [OpenClaw 官网](https://openclaw.ai/)
- [OpenClaw 文档](https://docs.openclaw.ai/)
- [OpenClaw Getting Started](https://docs.openclaw.ai/start/getting-started)
- [OpenClaw Onboarding Overview](https://docs.openclaw.ai/start/onboarding-overview)
- [OpenClaw Install](https://docs.openclaw.ai/install)
- [OpenClaw Installer Internals](https://docs.openclaw.ai/install/installer)
- [OpenClaw OpenAI Provider](https://docs.openclaw.ai/providers/openai)
- [OpenClaw Configuration Reference](https://docs.openclaw.ai/gateway/configuration-reference)
- [Node.js 安装教程](./node.js.md)
