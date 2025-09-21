# Salt Module

The `salt` module provides a fluent API to interact with SaltStack, allowing you to run remote execution commands and manage configurations from your `sloth-runner` workflows.

---

## `salt.client([options])`

Creates a Salt client object.

*   **Parameters:**
    *   `options` (table, optional): A table of options.
        *   `config_path` (string): Path to the Salt master configuration file.
*   **Returns:**
    *   `client` (object): A `SaltClient` object.

---

## The `SaltClient` Object

This object represents a client for a Salt master and provides methods for targeting minions.

### `client:target(target_string, [expr_form])`

Specifies the minion(s) to target for a command.

*   **Parameters:**
    *   `target_string` (string): The target expression (e.g., `"*"` for all minions, `"web-server-1"`, or a grain value).
    *   `expr_form` (string, optional): The type of targeting to use (e.g., `"glob"`, `"grain"`, `"list"`). Defaults to glob.
*   **Returns:**
    *   `target` (object): A `SaltTarget` object.

---

## The `SaltTarget` Object

This object represents a specific target and provides chainable methods for executing Salt functions.

### `target:cmd(function, [arg1, arg2, ...])`

Executes a Salt execution module function on the target.

*   **Parameters:**
    *   `function` (string): The name of the function to run (e.g., `"test.ping"`, `"state.apply"`, `"cmd.run"`).
    *   `arg1`, `arg2`, ... (any): Additional arguments to pass to the Salt function.
*   **Returns:**
    *   `result` (table): A table containing `success` (boolean), `stdout` (string or table), and `stderr` (string). If the Salt command returns JSON, `stdout` will be a parsed Lua table.

### Example

This example demonstrates targeting minions to ping them and apply a Salt state.

```lua
command = function()
  local salt = require("salt")

  -- 1. Create a Salt client
  local client = salt.client()

  -- 2. Target all minions and ping them
  log.info("Pinging all minions...")
  local ping_result = client:target("*"):cmd("test.ping")
  if not ping_result.success then
    return false, "Failed to ping minions: " .. ping_result.stderr
  end
  print("Ping Results:")
  print(data.to_yaml(ping_result.stdout)) -- stdout is a table

  -- 3. Target a specific web server and apply a state
  log.info("Applying 'nginx' state to web-server-1...")
  local apply_result = client:target("web-server-1", "glob"):cmd("state.apply", "nginx")
  if not apply_result.success then
    return false, "Failed to apply state: " .. apply_result.stderr
  end
  
  log.info("State applied successfully.")
  return true, "Salt operations complete."
end
```
