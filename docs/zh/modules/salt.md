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
    log.info("正在开始 Salt API 流畅测试...")

    -- 测试 1: 在 minion 'keiteguica' 上执行命令
    log.info("正在测试单个目标: keiteguica")
    -- 为目标 'keiteguica' 链式调用 ping() 命令
    salt.target('keiteguica'):ping()

    log.info("--------------------------------------------------")

    -- 测试 2: 使用 globbing 在多个 minion 上执行命令
    log.info("正在测试 glob 目标: vm-gcp-squid-proxy*")
    -- 为匹配模式的目标链式调用 ping() 和 cmd() 命令
    salt.target('vm-gcp-squid-proxy*'):ping():cmd('pkg.upgrade')

    log.info("Salt API 流畅测试完成。")

    log.info("正在通过 Salt 执行 'ls -la' 并处理输出...")
    local result_stdout, result_stderr, result_err = salt.target('keiteguica'):cmd('cmd.run', 'ls -la'):result()

    if result_err ~= nil then
        log.error("通过 Salt 执行 'ls -la' 时出错: " .. result_err)
        log.error("Stderr: " .. result_stderr)
    else
        log.info("通过 Salt 执行 'ls -la' 的输出:")
        -- 如果输出是表 (JSON)，您可以遍历它或将其转换为字符串
        if type(result_stdout) == "table" then
            log.info("JSON 输出 (表): " .. data.to_json(result_stdout))
        else
            log.info(result_stdout)
        end
    end
    log.info("通过 Salt 处理 'ls -la' 输出完成。")

    return true, "Salt API 流畅命令和 'ls -la' 已成功执行。"
end

TaskDefinitions = {
    test_fluent_salt = {
        description = "演示使用 'salt' 模块进行 SaltStack 编排。",
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
[English](../../en/modules/salt.md) | [Português](../../pt/modules/salt.md) | [中文](./salt.md)