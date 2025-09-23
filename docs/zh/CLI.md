# CLI 命令

`sloth-runner` 命令行界面 (CLI) 是与您的任务管道交互的主要方式。它提供了运行、列出、验证和管理工作流的命令。

---

## `sloth-runner run`

执行在 Lua 配置文件中定义的任务。这是您将使用的最常见的命令。

**用法:**
```bash
sloth-runner run [flags]
```

**标志:**

*   `-f, --file string`: **(必需)** Lua 任务配置文件的路径。
*   `-g, --group string`: 仅运行特定任务组中的任务。如果未提供，`sloth-runner` 将运行所有组中的任务。
*   `-t, --tasks string`: 要运行的特定任务的逗号分隔列表 (例如, `task1,task2`)。如果未提供，将考虑指定组（或所有组）中的所有任务。
*   `-v, --values string`: 包含要传递给 Lua 脚本的值的 YAML 文件的路径。这些值可通过全局 `values` 表在 Lua 中访问。
*   `-d, --dry-run`: 模拟任务的执行。它将打印将要运行的任务及其顺序，但不会执行它们的 `command`。
*   `--return`: 将已执行任务的最终输出作为 JSON 对象打印到标准输出。这包括最后一个任务的返回值和传递给全局 `export()` 函数的任何数据。
*   `-y, --yes`: 在未使用 `-t` 提供特定任务时，绕过交互式任务选择提示。

**示例:**

*   运行特定组中的所有任务:
    ```bash
    sloth-runner run -f examples/basic_pipeline.lua -g my_group
    ```
*   运行单个特定任务:
    ```bash

    sloth-runner run -f examples/basic_pipeline.lua -g my_group -t my_task
    ```
*   运行多个任务并将其组合输出作为 JSON 获取:
    ```bash
    sloth-runner run -f examples/export_example.lua -t export-data-task --return
    ```

---

## `sloth-runner list`

列出在 Lua 配置文件中定义的所有可用任务组和任务，以及它们的描述和依赖关系。

**用法:**
```bash
sloth-runner list [flags]
```

**标志:**

*   `-f, --file string`: **(必需)** Lua 任务配置文件的路径。
*   `-v, --values string`: YAML 值文件的路径，以防您的任务定义依赖于它。

---

## `sloth-runner new`

从模板生成一个新的 Lua 任务定义样板文件。

**用法:**
```bash
sloth-runner new <group-name> [flags]
```

**参数:**

*   `<group-name>`: 要在文件中创建的主任务组的名称。

**标志:**

*   `-t, --template string`: 要使用的模板。默认为 `simple`。运行 `sloth-runner template list` 查看所有可用选项。
*   `-o, --output string`: 输出文件的路径。如果未提供，生成的内容将打印到标准输出。
*   `--set key=value`: 传递键值对到模板，用于动态内容生成。

**示例:**
```bash
sloth-runner new my-python-pipeline -t python -o my_pipeline.lua
```

---

## `sloth-runner validate`

验证 Lua 任务文件的语法和基本结构，而不执行任何任务。

**用法:**
```bash
sloth-runner validate [flags]
```

**标志:**

*   `-f, --file string`: **(必需)** 要验证的 Lua 任务配置文件的路径。
*   `-v, --values string`: 如果验证需要，则为 YAML 值文件的路径。

---

## `sloth-runner test`

对工作流文件执行基于 Lua 的测试文件。(这是一个高级功能)。

**用法:**
```bash
sloth-runner test [flags]
```

**标志:**

*   `-w, --workflow string`: **(必需)** 要测试的 Lua 工作流文件的路径。
*   `-f, --file string`: **(必需)** Lua 测试文件的路径。

---

## `sloth-runner template list`

列出可与 `sloth-runner new` 命令一起使用的所有可用模板。

**用法:**
```bash
sloth-runner template list
```

---

## `sloth-runner version`

打印 `sloth-runner` 应用程序的当前版本。

**用法:**
```bash
sloth-runner version
```

### `sloth-runner scheduler`

管理 `sloth-runner` 任务调度器，允许您启用、禁用、列出和删除调度任务。

有关调度器命令和配置的详细信息，请参阅 [任务调度器文档](scheduler.md)。

**子命令:**

*   `sloth-runner scheduler enable`: 将调度器作为后台进程启动。
*   `sloth-runner scheduler disable`: 停止正在运行的调度器进程。
*   `sloth-runner scheduler list`: 列出所有已配置的调度任务。
*   `sloth-runner scheduler delete <task_name>`: 删除特定的调度任务。

