# Notifications Module

The `notifications` module provides a simple way to send messages to various notification services from your pipelines. This is particularly useful for reporting the success or failure of a CI/CD workflow.

Currently, the following services are supported:
- [Slack](#slack)
- [ntfy](#ntfy)

## Configuration

Before using the module, you need to add the required credentials or URLs to your `configs/values.yaml` file. The module will read these values at runtime.

```yaml
# configs/values.yaml

notifications:
  slack:
    # Your Slack Incoming Webhook URL
    webhook_url: "https://hooks.slack.com/services/..."
  ntfy:
    # The ntfy server to use. Can be the public one or self-hosted.
    server: "https://ntfy.sh"
    # The topic to publish the notification to.
    topic: "your-sloth-runner-topic"
```

## Slack

### `notifications.slack.send(params)`

Sends a message to a Slack channel via an Incoming Webhook.

**Parameters:**

- `params` (table): A table containing the following fields:
    - `webhook_url` (string): **Required.** The Slack Incoming Webhook URL. It's recommended to get this from the `values` module.
    - `message` (string): **Required.** The main text of the message.
    - `pipeline` (string): **Optional.** The name of the pipeline, which will be displayed in the message attachment for context.
    - `error_details` (string): **Optional.** Any error details to include in the message attachment. This is useful for failure notifications.

**Returns:**

- `true` on success.
- `false, error_message` on failure.

**Example:**

```lua
local values = require("values")

local slack_webhook = values.get("notifications.slack.webhook_url")

if slack_webhook and slack_webhook ~= "" then
  -- On success
  notifications.slack.send({
    webhook_url = slack_webhook,
    message = "✅ Pipeline executed successfully!",
    pipeline = "my-awesome-pipeline"
  })

  -- On failure
  notifications.slack.send({
    webhook_url = slack_webhook,
    message = "❌ Pipeline execution failed!",
    pipeline = "my-awesome-pipeline",
    error_details = "Could not connect to database."
  })
end
```

## ntfy

### `notifications.ntfy.send(params)`

Sends a message to an [ntfy.sh](https://ntfy.sh/) topic.

**Parameters:**

- `params` (table): A table containing the following fields:
    - `server` (string): **Required.** The ntfy server URL.
    - `topic` (string): **Required.** The topic to send the message to.
    - `message` (string): **Required.** The body of the notification.
    - `title` (string): **Optional.** The title of the notification.
    - `priority` (string): **Optional.** Notification priority (e.g., `high`, `default`, `low`).
    - `tags` (table): **Optional.** A list of tags (emojis) to add to the notification.

**Returns:**

- `true` on success.
- `false, error_message` on failure.

**Example:**

```lua
local values = require("values")

local ntfy_server = values.get("notifications.ntfy.server")
local ntfy_topic = values.get("notifications.ntfy.topic")

if ntfy_topic and ntfy_topic ~= "" then
  -- On success
  notifications.ntfy.send({
    server = ntfy_server,
    topic = ntfy_topic,
    title = "Pipeline Success",
    message = "The pipeline finished without errors.",
    priority = "default",
    tags = {"tada"}
  })

  -- On failure
  notifications.ntfy.send({
    server = ntfy_server,
    topic = ntfy_topic,
    title = "Pipeline Failed!",
    message = "The pipeline failed with an error.",
    priority = "high",
    tags = {"skull", "warning"}
  })
end
```
