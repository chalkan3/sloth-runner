# Salt 模块

Sloth-Runner 中的 `salt` 模块提供了一个流畅的 API，可以直接从您的 Lua 脚本与 SaltStack 进行交互。这使您能够自动化服务器编排和配置，将 Salt 的强大功能集成到您的 Sloth-Runner 工作流中。

## 常见用例

*   **配置自动化：** 将 Salt 状态 (`state.apply`) 应用于特定的 minion。
*   **状态验证：** 执行 ping (`test.ping`) 以检查与 minion 的连接。
*   **远程命令执行：** 在一个或多个 minion 上执行任意命令 (`cmd.run`)。
*   **部署编排：** 使用 Salt 函数协调应用程序部署。

## API 参考

### `salt.target(target_string)`

定义后续 Salt 操作的目标（minion 或 minion 组）。

*   `target_string` (字符串)：minion ID、glob、列表或 Salt 支持的其他目标类型。

**返回：**
*   `SaltTargeter` (用户数据)：指定目标的 `SaltTargeter` 对象的实例。

### `SaltTargeter` 对象方法（可链式调用）

以下所有方法都在 `SaltTargeter` 实例上调用（例如，`minion:ping()`），并返回 `SaltTargeter` 实例本身以允许方法链式调用。要获取上次操作的结果，请使用 `:result()` 方法。

#### `target:ping()`

在定义的目标上执行 `test.ping` 命令。

#### `target:cmd(function, ...args)`

在目标上执行任意 Salt 函数。

*   `function` (字符串)：要执行的 Salt 函数的名称（例如，“state.apply”、“cmd.run”、“pkg.upgrade”）。
*   `...args` (可变参数)：要传递给 Salt 函数的其他参数。

#### `target:result()`

返回在 `SaltTargeter` 实例上执行的上次 Salt 操作的结果。

**返回：**
*   `result` (Lua 表)：包含以下内容的表：
    *   `success` (布尔值)：如果操作成功，则为 `true`；否则为 `false`。
    *   `stdout` (字符串或 Lua 表)：Salt 命令的标准输出。如果 Salt 返回有效的 JSON，它将是一个 Lua 表。
    *   `stderr` (字符串)：Salt 命令的标准错误输出。
    *   `error` (字符串或 `nil`)：如果命令执行失败，则为 Go 错误消息。

## 使用示例

### 基本 Salt 编排示例

此示例演示如何使用流畅的 Salt API 执行 ping 并对 minion 执行命令。

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
**可用语言：**
[English](../en/modules/salt.md) | [Português ../../pt/modules/salt.md) | [中文](./salt.md)