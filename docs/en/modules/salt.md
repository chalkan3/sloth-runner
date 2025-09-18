# Salt Module

The `salt` module in Sloth-Runner provides a fluent API to interact with SaltStack directly from your Lua scripts. This allows you to automate server orchestration and configuration, integrating the power of Salt into your Sloth-Runner workflows.

## Common Use Cases

*   **Configuration Automation:** Apply Salt states (`state.apply`) to specific minions.
*   **Status Verification:** Perform pings (`test.ping`) to check connectivity with minions.
*   **Remote Command Execution:** Execute arbitrary commands (`cmd.run`) on one or more minions.
*   **Deployment Orchestration:** Coordinate application deployments using Salt functions.

## API Reference

### `salt.target(target_string)`

Defines the target (minion or minion group) for subsequent Salt operations.

*   `target_string` (string): The minion ID, glob, list, or other target type supported by Salt.

**Returns:**
*   `SaltTargeter` (userdata): An instance of the `SaltTargeter` object for the specified target.

### `SaltTargeter` Object Methods (Chainable)

All methods below are called on the `SaltTargeter` instance (e.g., `minion:ping()`) and return the `SaltTargeter` instance itself to allow method chaining. To get the result of the last operation, use the `:result()` method.

#### `target:ping()`

Executes the `test.ping` command on the defined target.

#### `target:cmd(function, ...args)`

Executes an arbitrary Salt function on the target.

*   `function` (string): The name of the Salt function to execute (e.g., "state.apply", "cmd.run", "pkg.upgrade").
*   `...args` (variadic): Additional arguments to be passed to the Salt function.

#### `target:result()`

Returns the result of the last Salt operation executed on the `SaltTargeter` instance.

**Returns:**
*   `result` (Lua table): A table containing:
    *   `success` (boolean): `true` if the operation was successful, `false` otherwise.
    *   `stdout` (string or Lua table): The standard output of the Salt command. If Salt returns valid JSON, it will be a Lua table.
    *   `stderr` (string): The standard error output of the Salt command.
    *   `error` (string or `nil`): A Go error message if the command execution failed.

## Usage Examples

### Basic Salt Orchestration Example

This example demonstrates how to use the fluent Salt API to perform pings and execute commands on minions.

```lua
-- examples/fluent_salt_api_test.lua

command = function()
    log.info("Starting Salt API fluent test...")

    -- Test 1: Executing commands on minion 'keiteguica'
    log.info("Testing single target: keiteguica")
    -- Chain the ping() command for target 'keiteguica'
    salt.target('keiteguica'):ping()

    log.info("--------------------------------------------------")

    -- Test 2: Executing commands on multiple minions using globbing
    log.info("Testing glob target: vm-gcp-squid-proxy*")
    -- Chain ping() and cmd() commands for targets matching the pattern
    salt.target('vm-gcp-squid-proxy*'):ping():cmd('pkg.upgrade')

    log.info("Salt API fluent test completed.")

    log.info("Executing 'ls -la' via Salt and processing output...")
    local result_stdout, result_stderr, result_err = salt.target('keiteguica'):cmd('cmd.run', 'ls -la'):result()

    if result_err ~= nil then
        log.error("Error executing 'ls -la' via Salt: " .. result_err)
        log.error("Stderr: " .. result_stderr)
    else
        log.info("Output of 'ls -la' via Salt:")
        -- If the output is a table (JSON), you can iterate over it or convert it to string
        if type(result_stdout) == "table" then
            log.info("JSON Output (table): " .. data.to_json(result_stdout))
        else
            log.info(result_stdout)
        end
    end
    log.info("Processing 'ls -la' output via Salt completed.")

    return true, "Salt API fluent commands and 'ls -la' executed successfully."
end

TaskDefinitions = {
    test_fluent_salt = {
        description = "Demonstrates using the 'salt' module for SaltStack orchestration.",
        tasks = {
            {
                name = "run_salt_orchestration",
                command = command
            }
        }
    }
}
```

---
**Available Languages:**
[English](./salt.md) | [Português](../../pt/modules/salt.md) | [中文](../../zh/modules/salt.md)