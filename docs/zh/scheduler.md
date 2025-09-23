# 任务调度器

`sloth-runner` 现在包含一个内置的任务调度器，允许您使用 cron 语法在指定的时间间隔自动执行您的 Lua 定义任务。

## 功能

*   **后台进程:** 调度器作为持久的后台进程运行，独立于您的终端会话。
*   **基于 Cron 的调度:** 使用灵活的 cron 字符串定义任务调度。
*   **持久性:** 调度任务从配置文件加载，确保在重启后恢复。
*   **与现有任务集成:** 调度器利用现有的 `sloth-runner run` 命令来执行您的任务。

## 配置: `scheduler.yaml`

调度任务在 YAML 文件中定义，通常命名为 `scheduler.yaml`。此文件指定要运行的任务、它们的调度以及 Lua 文件、组和任务名称。

```yaml
scheduled_tasks:
  - name: "my_daily_backup"
    schedule: "0 0 * * *" # 每天午夜
    task_file: "examples/my_workflow.lua"
    task_group: "backup_group"
    task_name: "perform_backup"
  - name: "hourly_report_generation"
    schedule: "0 * * * *" # 每小时
    task_file: "examples/reporting.lua"
    task_group: "reports"
    task_name: "generate_report"
```

**字段:**

*   `name` (字符串, 必填): 调度任务的唯一名称。
*   `schedule` (字符串, 必填): 定义任务何时运行的 cron 字符串。支持标准 cron 语法和一些预定义调度 (例如, `@every 1h`, `@daily`)。有关详细信息，请参阅 [robfig/cron 文档](https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format)。
*   `task_file` (字符串, 必填): Lua 任务定义文件的路径。
*   `task_group` (字符串, 必填): Lua 文件中的任务组名称。
*   `task_name` (字符串, 必填): 在任务组中执行的特定任务名称。

## CLI 命令

### `sloth-runner scheduler enable`

将 `sloth-runner` 调度器作为后台进程启动。此命令确保调度器正在运行并准备好处理调度任务。

```bash
sloth-runner scheduler enable --scheduler-config scheduler.yaml
```

*   `--scheduler-config` (或 `-c`): 指定 `scheduler.yaml` 配置文件的路径。默认为当前目录中的 `scheduler.yaml`。

执行后，命令将打印后台调度器进程的 PID。即使您的终端会话关闭，调度器也将继续运行。

### `sloth-runner scheduler disable`

停止正在运行的 `sloth-runner` 调度器后台进程。

```bash
sloth-runner scheduler disable
```

此命令将尝试优雅地终止调度器进程。如果成功，它将删除由 `enable` 命令创建的 PID 文件。

### `sloth-runner scheduler list`

列出 `scheduler.yaml` 配置文件中定义的所有调度任务。此命令提供已配置任务、其调度和相关 Lua 任务详细信息的概述。

```bash
sloth-runner scheduler list --scheduler-config scheduler.yaml
```

*   `--scheduler-config` (或 `-c`): 指定 `scheduler.yaml` 配置文件的路径。默认为当前目录中的 `scheduler.yaml`。

**示例输出:**

```
# Configured Scheduled Tasks

NAME                     | SCHEDULE    | FILE                     | GROUP        | TASK
my_daily_backup          | 0 0 * * *   | examples/my_workflow.lua | backup_group | perform_backup
hourly_report_generation | 0 * * * *   | examples/reporting.lua   | reports      | generate_report
```

### `sloth-runner scheduler delete <task_name>`

从 `scheduler.yaml` 配置文件中删除特定的调度任务。此命令将删除任务定义，调度器将不再执行它。

```bash
sloth-runner scheduler delete my_daily_backup --scheduler-config scheduler.yaml
```

*   `<task_name>` (字符串, 必填): 要删除的调度任务的唯一名称。
*   `--scheduler-config` (或 `-c`): 指定 `scheduler.yaml` 配置文件的路径。默认为当前目录中的 `scheduler.yaml`。

**重要:** 此命令会修改您的 `scheduler.yaml` 文件。如有必要，请确保您有备份。如果调度器当前正在运行，您可能需要禁用并重新启用它才能使更改立即生效。

## 日志和错误处理

调度器将其活动和调度任务的执行状态记录到标准输出和标准错误。建议在生产环境中运行时将这些输出重定向到日志文件。

如果调度任务失败，调度器将记录错误并继续执行其他调度任务。它不会因单个任务失败而停止。

## 示例

1.  创建 `scheduler.yaml` 文件:

    ```yaml
    scheduled_tasks:
      - name: "my_test_task"
        schedule: "@every 1m"
        task_file: "examples/basic_pipeline.lua"
        task_group: "basic_pipeline"
        task_name: "fetch_data"
    ```

2.  启用调度器:

    ```bash
    sloth-runner scheduler enable --scheduler-config scheduler.yaml
    ```

3.  观察输出。每分钟，您应该会看到指示 `my_test_task` 执行的消息。

4.  停止调度器:

    ```bash
    sloth-runner scheduler disable
    ```
