# Log Module

The `log` module provides a simple and essential interface for logging messages from within your Lua scripts to the `sloth-runner` console. Using this module is the standard way to provide feedback and debug information during a task's execution.

---

## `log.info(message)`

Logs a message at the INFO level. This is the standard level for general, informative messages.

*   **Parameters:**
    *   `message` (string): The message to log.

---

## `log.warn(message)`

Logs a message at the WARN level. This is suitable for non-critical issues that should be brought to the user's attention.

*   **Parameters:**
    *   `message` (string): The message to log.

---

## `log.error(message)`

Logs a message at the ERROR level. This should be used for significant errors that might cause a task to fail.

*   **Parameters:**
    *   `message` (string): The message to log.

---

## `log.debug(message)`

Logs a message at the DEBUG level. These messages are typically hidden unless the runner is in a verbose or debug mode. They are useful for detailed diagnostic information.

*   **Parameters:**
    *   `message` (string): The message to log.

### Example

```lua
command = function()
  -- The log module is globally available and does not need to be required.
  
  log.info("Starting the logging example task.")
  
  local user_name = "Sloth"
  log.debug("Current user is: " .. user_name)

  if user_name ~= "Sloth" then
    log.warn("The user is not the expected one.")
  end

  log.info("Task is performing its main action...")
  
  local success = true -- Simulate a successful operation
  if not success then
    log.error("The main action failed unexpectedly!")
    return false, "Main action failed"
  end

  log.info("Logging example task finished successfully.")
  return true, "Logging demonstrated."
end
```
