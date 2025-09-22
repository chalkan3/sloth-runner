# 通知模块

`notifications` 模块提供了一种从您的管道向各种通知服务发送消息的简单方法。这对于报告 CI/CD 工作流的成功或失败特别有用。

目前支持以下服务：
- [Slack](#slack)
- [ntfy](#ntfy)

## 配置

在使用该模块之前，您需要将所需的凭据或 URL 添加到您的 `configs/values.yaml` 文件中。该模块将在运行时读取这些值。

```yaml
# configs/values.yaml

notifications:
  slack:
    # 您的 Slack Incoming Webhook URL
    webhook_url: "https://hooks.slack.com/services/..."
  ntfy:
    # 要使用的 ntfy 服务器。可以是公共服务器或自托管服务器。
    server: "https://ntfy.sh"
    # 用于发布通知的主题。
    topic: "your-sloth-runner-topic"
```

## Slack

### `notifications.slack.send(params)`

通过 Incoming Webhook 向 Slack 频道发送消息。

**参数:**

- `params` (table): 一个包含以下字段的表：
    - `webhook_url` (string): **必需。** Slack Incoming Webhook URL。建议从 `values` 模块获取。
    - `message` (string): **必需。** 消息的主要文本。
    - `pipeline` (string): **可选。** 管道的名称，将显示在消息附件中以提供上下文。
    - `error_details` (string): **可选。** 要包含在消息附件中的任何错误详细信息。这对于失败通知很有用。

**返回:**

- 成功时返回 `true`。
- 失败时返回 `false, error_message`。

**示例:**

```lua
local values = require("values")

local slack_webhook = values.get("notifications.slack.webhook_url")

if slack_webhook and slack_webhook ~= "" then
  -- 成功时
  notifications.slack.send({
    webhook_url = slack_webhook,
    message = "✅ 管道成功执行！",
    pipeline = "my-awesome-pipeline"
  })

  -- 失败时
  notifications.slack.send({
    webhook_url = slack_webhook,
    message = "❌ 管道执行失败！",
    pipeline = "my-awesome-pipeline",
    error_details = "无法连接到数据库。"
  })
end
```

## ntfy

### `notifications.ntfy.send(params)`

向 [ntfy.sh](https://ntfy.sh/) 主题发送消息。

**参数:**

- `params` (table): 一个包含以下字段的表：
    - `server` (string): **必需。** ntfy 服务器 URL。
    - `topic` (string): **必需。** 要发送消息的主题。
    - `message` (string): **必需。** 通知的正文。
    - `title` (string): **可选。** 通知的标题。
    - `priority` (string): **可选。** 通知优先级（例如 `high`, `default`, `low`）。
    - `tags` (table): **可选。** 要添加到通知中的标签（表情符号）列表。

**返回:**

- 成功时返回 `true`。
- 失败时返回 `false, error_message`。

**示例:**

```lua
local values = require("values")

local ntfy_server = values.get("notifications.ntfy.server")
local ntfy_topic = values.get("notifications.ntfy.topic")

if ntfy_topic and ntfy_topic ~= "" then
  -- 成功时
  notifications.ntfy.send({
    server = ntfy_server,
    topic = ntfy_topic,
    title = "管道成功",
    message = "管道无错误完成。",
    priority = "default",
    tags = {"tada"}
  })

  -- 失败时
  notifications.ntfy.send({
    server = ntfy_server,
    topic = ntfy_topic,
    title = "管道失败！",
    message = "管道因错误而失败。",
    priority = "high",
    tags = {"skull", "warning"}
  })
end
```
