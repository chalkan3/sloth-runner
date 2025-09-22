--
-- notifications_example.lua
--
-- This example demonstrates how to use the notifications module to send
-- messages to Slack and ntfy based on the outcome of a pipeline.
--
-- To run this example:
-- 1. Edit `configs/values.yaml` and add your Slack webhook URL and ntfy topic.
--    notifications:
--      slack:
--        webhook_url: "https://hooks.slack.com/services/..."
--      ntfy:
--        server: "https://ntfy.sh"
--        topic: "your-sloth-runner-topic"
--
-- 2. Run the pipeline from your terminal:
--    go run ./cmd/sloth-runner -f examples/notifications_example.lua
--

local values = require("values")
local log = require("log")

-- Get notification configurations from the values file
local slack_webhook = values.get("notifications.slack.webhook_url")
local ntfy_server = values.get("notifications.ntfy.server")
local ntfy_topic = values.get("notifications.ntfy.topic")

-- This function simulates the main logic of your pipeline.
-- To test the failure notification, simply uncomment the 'error' line.
local function run_pipeline_logic()
  log.info("Starting the example pipeline...")
  -- Simulate some work being done
  log.info("Task 1: Building project... done.")
  log.info("Task 2: Running tests... done.")

  -- Uncomment the following line to simulate a pipeline failure
  -- error("Something went terribly wrong during deployment!")

  log.info("Task 3: Deploying to production... done.")
  log.info("Pipeline finished successfully!")
end

-- We wrap the main pipeline logic in a 'pcall' (protected call).
-- This allows us to catch any errors and handle the notification logic gracefully.
local ok, err = pcall(run_pipeline_logic)

-- After the pipeline runs, send notifications based on the outcome.
if ok then
  log.info("Sending success notifications...")

  -- Send Slack notification on success
  if slack_webhook and slack_webhook ~= "" then
    notifications.slack.send({
      webhook_url = slack_webhook,
      message = "✅ Pipeline executed successfully!",
      pipeline = "notifications-example"
    })
    log.info("Slack success notification sent.")
  end

  -- Send ntfy notification on success
  if ntfy_topic and ntfy_topic ~= "" then
    notifications.ntfy.send({
      server = ntfy_server,
      topic = ntfy_topic,
      title = "Pipeline Success",
      message = "The 'notifications-example' pipeline finished without errors.",
      priority = "default"
    })
    log.info("ntfy success notification sent.")
  end
else
  log.error("Pipeline failed. Sending failure notifications...")
  local error_message = tostring(err)

  -- Send Slack notification on failure
  if slack_webhook and slack_webhook ~= "" then
    notifications.slack.send({
      webhook_url = slack_webhook,
      message = "❌ Pipeline execution failed!",
      pipeline = "notifications-example",
      error_details = error_message
    })
    log.info("Slack failure notification sent.")
  end

  -- Send ntfy notification on failure
  if ntfy_topic and ntfy_topic ~= "" then
    notifications.ntfy.send({
      server = ntfy_server,
      topic = ntfy_topic,
      title = "Pipeline Failed!",
      message = "The 'notifications-example' pipeline failed.",
      priority = "high",
      tags = {"warning", "skull"}
    })
    log.info("ntfy failure notification sent.")
  end
end
