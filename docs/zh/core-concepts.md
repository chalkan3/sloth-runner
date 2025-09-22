# 核心概念

本文档解释了 `sloth-runner` 的基本概念，帮助您理解如何定义和编排复杂的工作流。

---

## `TaskDefinitions` 表

任何 `sloth-runner` 工作流的入口点都是一个返回名为 `TaskDefinitions` 的全局 Lua 表的 Lua 文件。此表是一个字典，其中每个键都是一个 **任务组** 名称。

```lua
-- my_pipeline.lua
TaskDefinitions = {
  -- 在此处定义任务组
}
```

---

## 任务组

任务组是相关任务的集合。它还可以定义影响其中所有任务的属性。

**组属性:**

*   `description` (string): 组功能的描述。
*   `tasks` (table): 单个任务表的列表。
*   `create_workdir_before_run` (boolean): 如果为 `true`，则在任何任务运行之前为该组创建一个临时工作目录。此目录会传递给每个任务。
*   `clean_workdir_after_run` (function): 一个 Lua 函数，用于决定在组完成后是否应删除临时工作目录。它接收组的最终结果 (`{success = true/false, ...}`)。返回 `true` 将删除目录。

**示例:**
```lua
TaskDefinitions = {
  my_group = {
    description = "一个管理自己临时目录的组。",
    create_workdir_before_run = true,
    clean_workdir_after_run = function(result)
      if not result.success then
        log.warn("组失败。工作目录将保留用于调试。")
      end
      return result.success -- 仅在一切成功时清理
    end,
    tasks = {
      -- 任务在此处定义
    }
  }
}
```

---

## 单个任务

任务是工作的单个单元。它被定义为一个具有多个可用属性以控制其行为的表。

### 基本属性

*   `name` (string): 任务在其组中的唯一名称。
*   `description` (string): 任务功能的简要描述。
*   `command` (string 或 function): 任务的核心操作。
    *   **作为字符串:** 作为 shell 命令执行。
    *   **作为函数:** 执行 Lua 函数。它接收两个参数：`params` (其参数表) 和 `deps` (其依赖项的输出表)。该函数必须返回：
        1.  `boolean`: `true` 表示成功，`false` 表示失败。
        2.  `string`: 描述结果的消息。
        3.  `table` (可选): 其他任务可以依赖的输出表。

### 依赖与执行流程

*   `depends_on` (string 或 table): 在此任务运行之前必须成功完成的任务名称列表。
*   `next_if_fail` (string 或 table): *仅当* 此任务失败时才运行的任务名称列表。这对于清理或通知任务很有用。
*   `async` (boolean): 如果为 `true`，任务将在后台运行，运行器不会等待它完成再开始执行顺序中的下一个任务。

### 错误处理与稳健性

*   `retries` (number): 如果任务失败，重试的次数。默认为 `0`。
*   `timeout` (string): 一个持续时间 (例如 `"10s"`, `"1m"`), 如果任务仍在运行，则在此时间后终止。

### 条件执行

*   `run_if` (string 或 function): 除非满足此条件，否则将跳过该任务。
    *   **作为字符串:** 一个 shell 命令。退出代码 `0` 表示条件满足。
    *   **作为函数:** 一个返回 `true` 表示任务应运行的 Lua 函数。
*   `abort_if` (string 或 function): 如果满足此条件，整个工作流将被中止。
    *   **作为字符串:** 一个 shell 命令。退出代码 `0` 表示中止。
    *   **作为函数:** 一个返回 `true` 表示中止的 Lua 函数。

### 生命周期钩子

*   `pre_exec` (function): 在主 `command` *之前* 运行的 Lua 函数。
*   `post_exec` (function): 在主 `command` 成功完成 *之后* 运行的 Lua 函数。

### 可重用性

*   `uses` (table): 指定从另一个文件（通过 `import` 加载）的预定义任务作为基础。然后，当前任务定义可以覆盖 `params` 或 `description` 等属性。
*   `params` (table): 可以传递给任务的 `command` 函数的键值对字典。
*   `artifacts` (string 或 table): 一个文件模式 (glob) 或模式列表，指定成功运行后应将任务 `workdir` 中的哪些文件保存为工件。
*   `consumes` (string 或 table): 前一个任务的工件名称（或名称列表），在运行此任务之前应将其复制到此任务的 `workdir` 中。

---

## 工件管理

Sloth-Runner 允许任务通过工件机制相互共享文件。一个任务可以“生产”一个或多个文件作为工件，后续任务可以“消费”这些工件。

这对于 CI/CD 管道非常有用，其中构建步骤可能会生成一个二进制文件（工件），然后由测试或部署步骤使用。

### 工作原理

1.  **生产工件:** 将 `artifacts` 键添加到您的任务定义中。该值可以是单个文件模式 (例如 `"report.txt"`) 或列表 (例如 `{"*.log", "app.bin"}`)。任务成功运行后，运行器将在任务的 `workdir` 中查找与这些模式匹配的文件，并将它们复制到管道的共享工件存储中。

2.  **消费工件:** 将 `consumes` 键添加到另一个任务的定义中（通常 `depends_on` 生产者任务）。该值应该是您要使用的工件的文件名 (例如 `"report.txt"`)。在此任务运行之前，运行器会将指定的工件从共享存储复制到此任务的 `workdir` 中，使其可用于 `command`。

### 工件示例

```lua
TaskDefinitions = {
  ["ci-pipeline"] = {
    description = "演示工件的使用。",
    create_workdir_before_run = true,
    tasks = {
      {
        name = "build",
        description = "创建一个二进制文件并将其声明为工件。",
        command = "echo 'binary_content' > app.bin",
        artifacts = {"app.bin"}
      },
      {
        name = "test",
        description = "消费二进制文件以运行测试。",
        depends_on = "build",
        consumes = {"app.bin"},
        command = function(params)
          -- 此时, 'app.bin' 存在于此任务的 workdir 中
          local content, err = fs.read(params.workdir .. "/app.bin")
          if content == "binary_content" then
            log.info("成功消费工件！")
            return true
          else
            return false, "工件内容不正确！"
          end
        end
      }
    }
  }
}
```

---

## 全局函数

`sloth-runner` 在 Lua 环境中提供全局函数以帮助编排工作流。

### `import(path)`

加载另一个 Lua 文件并返回其返回的值。这是创建可重用任务模块的主要机制。路径是相对于调用 `import` 的文件的。

**示例 (`reusable_tasks.lua`):**
```lua
-- 导入一个返回任务定义表的模块
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
  main = {
    tasks = {
      {
        -- 使用导入模块中的 'build' 任务
        uses = docker_tasks.build,
        params = { image_name = "my-app" }
      }
    }
  }
}
```

### `parallel(tasks)`

并发执行任务列表，并等待所有任务完成。

*   `tasks` (table): 要并行运行的任务表列表。

**示例:**
```lua
command = function()
  log.info("并行启动3个任务...")
  local results, err = parallel({
    { name = "short_task", command = "sleep 1" },
    { name = "medium_task", command = "sleep 2" },
    { name = "long_task", command = "sleep 3" }
  })
  if err then
    return false, "并行执行失败"
  end
  return true, "所有并行任务已完成。"
end
```

### `export(table)`

将数据从脚本的任何位置导出到 CLI。当使用 `--return` 标志时，所有导出的表都会与最终任务的输出合并成一个 JSON 对象。

*   `table`: 要导出的 Lua 表。

**示例:**
```lua
command = function()
  export({ important_value = "来自任务中间的数据" })
  return true, "任务完成", { final_output = "一些结果" }
end
```
使用 `--return` 运行将产生：
```json
{
  "important_value": "来自任务中间的数据",
  "final_output": "一些结果"
}
```