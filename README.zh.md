[English](./README.md) | [Português](./README.pt.md) | [中文](./README.zh.md)

# 🦥 Sloth Runner 🚀

一个用 Go 编写、由 Lua 脚本驱动的灵活且可扩展的任务运行器应用程序。`sloth-runner` 允许您通过简单的 Lua 脚本定义复杂的工作流、管理任务依赖关系以及与外部系统集成。

[![Go CI](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml/badge.svg)](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml)

---

## ✨ 功能特性

*   **📜 Lua 脚本:** 使用强大而灵活的 Lua 脚本定义任务和工作流。
*   **🔗 依赖管理:** 指定任务依赖关系，确保复杂管道的有序执行。
*   **⚡ 异步任务执行:** 并发运行任务以提高性能。
*   **🪝 执行前后钩子:** 定义自定义 Lua 函数，在任务命令之前和之后运行。
*   **⚙️ 丰富的 Lua API:** 直接从 Lua 任务访问系统功能：
    *   **`exec` 模块:** 执行 shell 命令。
    *   **`fs` 模块:** 执行文件系统操作（读、写、追加、检查存在、创建目录、删除、递归删除、列出）。
    *   **`net` 模块:** 发出 HTTP 请求（GET、POST）和下载文件。
    *   **`data` 模块:** 解析和序列化 JSON 和 YAML 数据。
    *   **`log` 模块:** 以不同的严重级别（info、warn、error、debug）记录消息。
    *   **`salt` 模块:** 直接执行 SaltStack 命令（`salt`、`salt-call`）。
*   **📝 `values.yaml` 集成:** 通过 `values.yaml` 文件将配置值传递给您的 Lua 任务，类似于 Helm。
*   **💻 命令行界面 (CLI):**
    *   `run`: 从 Lua 配置文件执行任务。
    *   `list`: 列出所有可用的任务组和任务及其描述和依赖关系。


## 📚 完整文档

要获取更详细的文档、使用指南和高级示例，请访问我们的[完整文档](./docs/zh/index.md)。

---

## 🚀 开始使用

### 安装

要在您的系统上安装 `sloth-runner`，您可以使用提供的 `install.sh` 脚本。该脚本会自动检测您的操作系统和架构，从 GitHub 下载最新的发布版本，并将 `sloth-runner` 可执行文件放置在 `/usr/local/bin` 中。

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**注意:** `install.sh` 脚本需要 `sudo` 权限才能将可执行文件移动到 `/usr/local/bin`。

### 基本用法

要运行一个 Lua 任务文件：

```bash
sloth-runner run -f examples/basic_pipeline.lua
```

要列出文件中的任务：

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

---

## 📜 在 Lua 中定义任务

任务在 Lua 文件中定义，通常在一个 `TaskDefinitions` 表中。每个任务可以有 `name`、`description`、`command`（shell 命令字符串或 Lua 函数）、`async`（布尔值）、`pre_exec`（Lua 函数钩子）、`post_exec`（Lua 函数钩子）和 `depends_on`（字符串或字符串表）。

示例 (`examples/basic_pipeline.lua`):

```lua
-- 从另一个文件导入可重用的任务。路径是相对的。
local docker_tasks = import("examples/shared/docker.lua")

TaskDefinitions = {
    full_pipeline_demo = {
        description = "一个演示各种功能的综合管道。",
        tasks = {
            -- 任务 1: 获取数据，异步运行。
            fetch_data = {
                name = "fetch_data",
                description = "从 API 获取原始数据。",
                async = true,
                command = function(params)
                    log.info("正在获取数据...")
                    -- 模拟 API 调用
                    return true, "echo '获取了原始数据'", { raw_data = "api_data" }
                end,
            },

            -- 任务 2: 一个不稳定的任务，失败时会重试。
            flaky_task = {
                name = "flaky_task",
                description = "这个任务会间歇性失败，并且会重试。",
                retries = 3,
                command = function()
                    if math.random() > 0.5 then
                        log.info("不稳定的任务成功。")
                        return true, "echo '成功!'"
                    else
                        log.error("不稳定的任务失败，将重试...")
                        return false, "随机失败"
                    end
                end,
            },

            -- 任务 3: 处理数据，依赖于 fetch_data 和 flaky_task 的成功完成。
            process_data = {
                name = "process_data",
                description = "处理获取的数据。",
                depends_on = { "fetch_data", "flaky_task" },
                command = function(params, deps)
                    local raw_data = deps.fetch_data.raw_data
                    log.info("正在处理数据: " .. raw_data)
                    return true, "echo '处理了数据'", { processed_data = "processed_" .. raw_data }
                end,
            },

            -- 任务 4: 一个带超时的长时间运行任务。
            long_running_task = {
                name = "long_running_task",
                description = "一个如果运行时间过长将被终止的任务。",
                timeout = "5s",
                command = "echo '开始长任务...'; sleep 10; echo '这不会被打印出来。';",
            },

            -- 任务 5: 一个清理任务，如果 long_running_task 失败则运行。
            cleanup_on_fail = {
                name = "cleanup_on_fail",
                description = "仅在长时间运行的任务失败时运行。",
                next_if_fail = "long_running_task",
                command = "echo '由于先前的失败，清理任务已执行。'",
            },

            -- 任务 6: 使用从导入的 docker.lua 模块中可重用的任务。
            build_image = {
                uses = docker_tasks.build,
                description = "构建应用程序的 Docker 镜像。",
                params = {
                    image_name = "my-awesome-app",
                    tag = "v1.2.3",
                    context = "./app_context"
                }
            },

            -- 任务 7: 一个条件任务，仅在文件存在时运行。
            conditional_deploy = {
                name = "conditional_deploy",
                description = "仅在构建产物存在时部署应用程序。",
                depends_on = "build_image",
                run_if = "test -f ./app_context/artifact.txt", -- Shell 命令条件
                command = "echo '正在部署应用程序...'",
            },

            -- 任务 8: 如果满足条件，此任务将中止整个工作流。
            gatekeeper_check = {
                name = "gatekeeper_check",
                description = "如果关键条件未满足，则中止工作流。",
                abort_if = function(params, deps)
                    -- Lua 函数条件
                    log.warn("正在检查守门员条件...")
                    if params.force_proceed ~= "true" then
                        log.error("守门员检查失败。正在中止工作流。")
                        return true -- 中止
                    end
                    return false -- 不中止
                end,
                command = "echo '如果中止，此命令将不会运行。'"
            }
        }
    }
}
```

---

## 高级功能

`sloth-runner` 提供了几个高级功能，用于对任务执行进行精细控制。

### 任务重试和超时

您可以通过为不稳定的任务指定重试次数和为长时间运行的任务指定超时来使您的工作流更加健壮。

*   `retries`: 如果任务失败，重试的次数。
*   `timeout`: 一个持续时间字符串（例如 "10s", "1m"），超过该时间后任务将被终止。

<details>
<summary>示例 (`examples/retries_and_timeout.lua`):</summary>

```lua
TaskDefinitions = {
    robust_workflow = {
        description = "一个演示重试和超时的工作流",
        tasks = {
            {
                name = "flaky_task",
                description = "这个任务有 50% 的几率失败",
                retries = 3,
                command = function()
                    if math.random() < 0.5 then
                        log.error("模拟随机失败！")
                        return false, "发生随机失败"
                    end
                    return true, "echo '不稳定的任务成功！'", { result = "success" }
                end
            },
            {
                name = "long_running_task",
                description = "这个任务模拟一个将超时的长进程",
                timeout = "2s",
                command = "sleep 5 && echo '这不应该被打印出来'"
            }
        }
    }
}
```
</details>

### 条件执行: `run_if` 和 `abort_if`

您可以使用 `run_if` 和 `abort_if` 根据条件控制任务执行。这些可以是 shell 命令或 Lua 函数。

*   `run_if`: 只有在满足条件时才会执行任务。
*   `abort_if`: 如果满足条件，整个执行过程将被中止。

#### 使用 Shell 命令

执行 shell 命令，其退出代码决定结果。退出代码 `0` 表示条件满足（成功）。

<details>
<summary>示例 (`examples/conditional_execution.lua`):</summary>

```lua
TaskDefinitions = {
    conditional_workflow = {
        description = "一个使用 run_if 和 abort_if 演示条件执行的工作流。",
        tasks = {
            {
                name = "check_condition_for_run",
                description = "这个任务创建一个文件，下一个任务会检查该文件。",
                command = "touch /tmp/sloth_runner_run_condition"
            },
            {
                name = "conditional_task",
                description = "这个任务只有在条件文件存在时才运行。",
                depends_on = "check_condition_for_run",
                run_if = "test -f /tmp/sloth_runner_run_condition",
                command = "echo '条件任务正在运行，因为条件已满足。'"
            },
            {
                name = "check_abort_condition",
                description = "如果特定文件存在，此任务将中止。",
                abort_if = "test -f /tmp/sloth_runner_abort_condition",
                command = "echo '如果中止条件满足，这不会运行。'"
            }
        }
    }
}
```
</details>

#### 使用 Lua 函数

对于更复杂的逻辑，您可以使用 Lua 函数。该函数接收任务的 `params` 和 `deps`（来自依赖项的输出）。它必须返回 `true` 才能满足条件。

<details>
<summary>示例 (`examples/conditional_functions.lua`):</summary>

```lua
TaskDefinitions = {
    conditional_functions_workflow = {
        description = "一个使用 Lua 函数演示条件执行的工作流。",
        tasks = {
            {
                name = "setup_task",
                description = "此任务为条件任务提供输出。",
                command = function()
                    return true, "设置完成", { should_run = true }
                end
            },
            {
                name = "conditional_task_with_function",
                description = "此任务仅在 run_if 函数返回 true 时运行。",
                depends_on = "setup_task",
                run_if = function(params, deps)
                    log.info("正在检查 conditional_task_with_function 的 run_if 条件...")
                    if deps.setup_task and deps.setup_task.should_run == true then
                        log.info("条件满足，任务将运行。")
                        return true
                    end
                    log.info("条件不满足，任务将被跳过。")
                    return false
                end,
                command = "echo '条件任务正在运行，因为函数返回了 true。'"
            },
            {
                name = "abort_task_with_function",
                description = "如果 abort_if 函数返回 true，此任务将中止执行。",
                params = {
                    abort_execution = "true"
                },
                abort_if = function(params, deps)
                    log.info("正在检查 abort_task_with_function 的 abort_if 条件...")
                    if params.abort_execution == "true" then
                        log.info("中止条件满足，执行将停止。")
                        return true
                    end
                    log.info("中止条件不满足。")
                    return false
                end,
                command = "echo '这不应该被执行。'"
            }
        }
    }
}
```
</details>

### 使用 `import` 的可重用任务模块

您可以创建可重用的任务库，并将它们导入到您的主工作流文件中。这对于在多个项目之间共享通用任务（如构建 Docker 镜像、部署应用程序等）非常有用。

全局 `import()` 函数加载另一个 Lua 文件并返回其返回值。路径相对于调用 `import` 的文件进行解析。

**工作原理:**
1.  创建一个模块（例如 `shared/docker.lua`），定义一个任务表并返回它。
2.  在您的主文件中，调用 `import("shared/docker.lua")` 来加载模块。
3.  在您的主 `TaskDefinitions` 表中使用 `uses` 字段引用导入的任务。`sloth-runner` 将自动将导入的任务与您提供的任何本地覆盖（如 `description` 或 `params`）合并。

<details>
<summary>模块示例 (`examples/shared/docker.lua`):</summary>

```lua
-- examples/shared/docker.lua
-- 一个用于 Docker 任务的可重用模块。

local TaskDefinitions = {
    build = {
        name = "build",
        description = "构建一个 Docker 镜像",
        params = {
            tag = "latest",
            dockerfile = "Dockerfile",
            context = "."
        },
        command = function(params)
            local image_name = params.image_name or "my-default-image"
            -- ... 构建命令逻辑 ...
            local cmd = string.format("docker build -t %s:%s -f %s %s", image_name, params.tag, params.dockerfile, params.context)
            return true, cmd
        end
    },
    push = {
        name = "push",
        description = "将 Docker 镜像推送到注册表",
        -- ... 推送任务逻辑 ...
    }
}

return TaskDefinitions
```
</details>

<details>
<summary>用法示例 (`examples/reusable_tasks.lua`):</summary>

```lua
-- examples/reusable_tasks.lua

-- 导入可重用的 Docker 任务。
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
    app_deployment = {
        description = "一个使用可重用 Docker 模块的工作流。",
        tasks = {
            -- 使用模块中的 'build' 任务并覆盖其参数。
            build = {
                uses = docker_tasks.build,
                description = "构建主应用程序 Docker 镜像",
                params = {
                    image_name = "my-app",
                    tag = "v1.0.0",
                    context = "./app"
                }
            },
            
            -- 一个依赖于导入的 'build' 任务的常规任务。
            deploy = {
                name = "deploy",
                description = "部署应用程序",
                depends_on = "build",
                command = "echo '正在部署...'"
            }
        }
    }
}
```
</details>

---

## 💻 CLI 命令

`sloth-runner` 提供了一个简单而强大的命令行界面。

### `sloth-runner run`

执行在 Lua 模板文件中定义的任务。

**标志:**

*   `-f, --file string`: Lua 任务配置文件路径。
*   `-t, --tasks string`: 要运行的特定任务的逗号分隔列表。
*   `-g, --group string`: 仅运行特定任务组中的任务。
*   `-v, --values string`: 包含要传递给 Lua 任务的值的 YAML 文件路径。
*   `-d, --dry-run`: 模拟任务执行而不实际运行它们。

### `sloth-runner list`

列出在 Lua 模板文件中定义的所有可用任务组和任务。

**标志:**

*   `-f, --file string`: Lua 任务配置文件路径。
*   `-v, --values string`: 包含值的 YAML 文件路径。

---

## ⚙️ Lua API

`sloth-runner` 将几个 Go 功能作为 Lua 模块公开，允许您的任务与系统和外部服务进行交互。

*   **`exec` 模块:** 执行 shell 命令。
*   **`fs` 模块:** 执行文件系统操作。
*   **`net` 模块:** 发出 HTTP 请求和下载文件。
*   **`data` 模块:** 解析和序列化 JSON 和 YAML 数据。
*   **`log` 模块:** 以不同的严重级别记录消息。
*   **`salt` 模块:** 执行 SaltStack 命令。

有关详细的 API 用法，请参阅 `/examples` 目录中的示例。
