# Task Scheduler

The `sloth-runner` now includes a built-in task scheduler, allowing you to automate the execution of your Lua-defined tasks at specified intervals using cron syntax.

## Features

*   **Background Process:** The scheduler runs as a persistent background process, independent of your terminal session.
*   **Cron-based Scheduling:** Define task schedules using flexible cron strings.
*   **Persistence:** Scheduled tasks are loaded from a configuration file, ensuring they resume after restarts.
*   **Integration with Existing Tasks:** The scheduler leverages the existing `sloth-runner run` command to execute your tasks.

## Configuration: `scheduler.yaml`

Scheduled tasks are defined in a YAML file, typically named `scheduler.yaml`. This file specifies the tasks to run, their schedule, and the Lua file, group, and task name.

```yaml
scheduled_tasks:
  - name: "my_daily_backup"
    schedule: "0 0 * * *" # Every day at midnight
    task_file: "examples/my_workflow.lua"
    task_group: "backup_group"
    task_name: "perform_backup"
  - name: "hourly_report_generation"
    schedule: "0 * * * *" # Every hour
    task_file: "examples/reporting.lua"
    task_group: "reports"
    task_name: "generate_report"
```

**Fields:**

*   `name` (string, required): A unique name for the scheduled task.
*   `schedule` (string, required): The cron string defining when the task should run. Supports standard cron syntax and some predefined schedules (e.g., `@every 1h`, `@daily`). Refer to [robfig/cron documentation](https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format) for details.
*   `task_file` (string, required): The path to the Lua task definition file.
*   `task_group` (string, required): The name of the task group within the Lua file.
*   `task_name` (string, required): The name of the specific task to execute within the task group.

## CLI Commands

### `sloth-runner scheduler enable`

Starts the `sloth-runner` scheduler as a background process. This command ensures the scheduler is running and ready to process scheduled tasks.

```bash
sloth-runner scheduler enable --scheduler-config scheduler.yaml
```

*   `--scheduler-config` (or `-c`): Specifies the path to your `scheduler.yaml` configuration file. Defaults to `scheduler.yaml` in the current directory.

Upon execution, the command will print the PID of the background scheduler process. The scheduler will continue to run even if your terminal session is closed.

### `sloth-runner scheduler disable`

Stops the running `sloth-runner` scheduler background process.

```bash
sloth-runner scheduler disable
```

This command will attempt to gracefully terminate the scheduler process. If successful, it will remove the PID file created by the `enable` command.

### `sloth-runner scheduler list`

Lists all scheduled tasks defined in the `scheduler.yaml` configuration file. This command provides an overview of your configured tasks, their schedules, and associated Lua task details.

```bash
sloth-runner scheduler list --scheduler-config scheduler.yaml
```

*   `--scheduler-config` (or `-c`): Specifies the path to your `scheduler.yaml` configuration file. Defaults to `scheduler.yaml` in the current directory.

**Example Output:**

```
# Configured Scheduled Tasks

NAME                     | SCHEDULE    | FILE                     | GROUP        | TASK
my_daily_backup          | 0 0 * * *   | examples/my_workflow.lua | backup_group | perform_backup
hourly_report_generation | 0 * * * *   | examples/reporting.lua   | reports      | generate_report
```

### `sloth-runner scheduler delete <task_name>`

Deletes a specific scheduled task from the `scheduler.yaml` configuration file. This command removes the task definition, and the scheduler will no longer execute it.

```bash
sloth-runner scheduler delete my_daily_backup --scheduler-config scheduler.yaml
```

*   `<task_name>` (string, required): The unique name of the scheduled task to delete.
*   `--scheduler-config` (or `-c`): Specifies the path to your `scheduler.yaml` configuration file. Defaults to `scheduler.yaml` in the current directory.

**Important:** This command modifies your `scheduler.yaml` file. Ensure you have a backup if necessary. If the scheduler is currently running, you may need to disable and re-enable it for the changes to take effect immediately.

## Logging and Error Handling

The scheduler logs its activities and the execution status of scheduled tasks to standard output and standard error. It's recommended to redirect these outputs to a log file when running in a production environment.

If a scheduled task fails, the scheduler will log the error and continue with other scheduled tasks. It will not stop due to individual task failures.

## Example

1.  Create a `scheduler.yaml` file:

    ```yaml
    scheduled_tasks:
      - name: "my_test_task"
        schedule: "@every 1m"
        task_file: "examples/basic_pipeline.lua"
        task_group: "basic_pipeline"
        task_name: "fetch_data"
    ```

2.  Enable the scheduler:

    ```bash
    sloth-runner scheduler enable --scheduler-config scheduler.yaml
    ```

3.  Observe the output. Every minute, you should see messages indicating the execution of `my_test_task`.

4.  To stop the scheduler:

    ```bash
    sloth-runner scheduler disable
    ```
