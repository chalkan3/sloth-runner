# 核心概念

本文档解释了 Sloth-Runner 的核心概念，帮助您了解如何定义和执行任务。

## Lua 中的任务定义

Sloth-Runner 中的任务在 Lua 文件中定义，通常在一个名为 `TaskDefinitions` 的全局表中。此表是一个映射，其中键是任务组名称，值是组表。

### 任务组结构

每个任务组都有：
*   `description`：任务组的文本描述。
*   `tasks`：包含单个任务定义的表。

### 单个任务结构

每个单独的任务可以具有以下字段：

*   `name` (字符串)：任务在其组中的唯一名称。
*   `description` (字符串)：任务功能的简要描述。
*   `command` (字符串或 Lua 函数)：
    *   如果为 `string`，它将作为 shell 命令执行。
    *   如果为 `Lua function`，此函数将执行。它可以接收 `params`（任务参数）和 `deps`（来自依赖任务的输出）。该函数应返回 `true` 表示成功，`false` 表示失败，并可选择返回一条消息和输出表。
*   `async` (布尔值，可选)：如果为 `true`，任务将异步执行。默认为 `false`。
*   `pre_exec` (Lua 函数，可选)：在任务主 `command` 之前执行的 Lua 函数。
*   `post_exec` (Lua 函数，可选)：在任务主 `command` 之后执行的 Lua 函数。
*   `depends_on` (字符串或字符串表，可选)：必须成功完成才能运行此任务的任务名称。
*   `retries` (数字，可选)：如果任务失败，将重试的次数。默认为 `0`。
*   `timeout` (字符串，可选)：任务在仍在运行时将被终止的持续时间（例如，“10s”，“1m”）。
*   `run_if` (字符串或 Lua 函数，可选)：仅当此条件为真时才执行任务。可以是 shell 命令（退出代码 0 表示成功）或 Lua 函数（返回 `true` 表示成功）。
*   `abort_if` (字符串或 Lua 函数，可选)：如果此条件为真，则整个工作流执行将中止。可以是 shell 命令（退出代码 0 表示成功）或 Lua 函数（返回 `true` 表示成功）。
*   `next_if_fail` (字符串或字符串表，可选)：如果此任务失败，将执行的任务名称。

### `TaskDefinitions` 结构示例

```lua
TaskDefinitions = {
    my_first_group = {
        description = "一个示例任务组。",
        tasks = {
            my_first_task = {
                name = "my_first_task",
                description = "一个执行 shell 命令的简单任务。",
                command = "echo 'Hello from Sloth-Runner!'"
            },
            my_second_task = {
                name = "my_second_task",
                description = "一个依赖于第一个任务并使用 Lua 函数的任务。",
                depends_on = "my_first_task",
                command = function(params, deps)
                    log.info("正在执行第二个任务。")
                    -- 您可以通过 'deps' 访问以前任务的输出
                    -- local output_from_first = deps.my_first_task.some_output
                    return true, "echo 'Second task completed!'"
                end
            }
        }
    }
}
```

## 参数和输出

*   **参数 (`params`)：** 可以通过命令行传递给任务，或在任务本身中定义。`command` 函数和 `run_if`/`abort_if` 函数可以访问它们。
*   **输出 (`deps`)：** Lua `command` 函数可以返回一个输出表。依赖于此任务的任务可以通过 `deps` 参数访问这些输出。

## 内置模块

Sloth-Runner 将各种 Go 功能公开为 Lua 模块，允许您的任务与系统和外部服务进行交互。除了基本模块（`exec`、`fs`、`net`、`data`、`log`、`import`、`parallel`）之外，Sloth-Runner 现在还包括用于 Git、Pulumi 和 Salt 的高级模块。

这些模块提供了流畅直观的 API，可实现复杂的自动化。

*   **`exec` 模块：** 用于执行任意 shell 命令。
*   **`fs` 模块：** 用于文件系统操作（读取、写入等）。
*   **`net` 模块：** 用于发出 HTTP 请求和下载。
*   **`data` 模块：** 用于解析和序列化 JSON 和 YAML。
*   **`log` 模块：** 用于将消息记录到 Sloth-Runner 控制台。
*   **`import` 函数：** 用于导入其他 Lua 文件和重用任务。
*   **`parallel` 函数：** 用于并行执行任务。
*   **`git` 模块：** 用于与 Git 存储库交互。
*   **`pulumi` 模块：** 用于编排 Pulumi 堆栈。
*   **`salt` 模块：** 用于执行 SaltStack 命令。

有关每个模块的详细信息，请参阅文档中的相应部分。

---
[English](../en/core-concepts.md) | [Português](../pt/core-concepts.md) | [中文](./core-concepts.md)