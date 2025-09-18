# Core Concepts

This document explains the fundamental concepts of Sloth-Runner, helping you understand how tasks are defined and executed.

## Defining Tasks in Lua

Tasks in Sloth-Runner are defined in Lua files, typically within a global table called `TaskDefinitions`. This table is a map where keys are task group names and values are group tables.

### Task Group Structure

Each task group has:
*   `description`: A textual description of the group.
*   `tasks`: A table containing individual task definitions.

### Individual Task Structure

Each individual task can have the following fields:

*   `name` (string): The unique name of the task within its group.
*   `description` (string): A brief description of what the task does.
*   `command` (string or Lua function):
    *   If a `string`, it will be executed as a shell command.
    *   If a `Lua function`, this function will be executed. It can receive `params` (task parameters) and `deps` (outputs from dependent tasks). The function should return `true` for success, `false` for failure, and optionally a message and an outputs table.
*   `async` (boolean, optional): If `true`, the task will be executed asynchronously. Default is `false`.
*   `pre_exec` (Lua function, optional): A Lua function to be executed before the task's main `command`.
*   `post_exec` (Lua function, optional): A Lua function to be executed after the task's main `command`.
*   `depends_on` (string or table of strings, optional): Names of tasks that must complete successfully before this task can run.
*   `retries` (number, optional): The number of times the task will be retried if it fails. Default is `0`.
*   `timeout` (string, optional): A duration (e.g., "10s", "1m") after which the task will be terminated if still running.
*   `run_if` (string or Lua function, optional): The task will only be executed if this condition is true. Can be a shell command (exit code 0 for success) or a Lua function (returns `true` for success).
*   `abort_if` (string or Lua function, optional): If this condition is true, the entire workflow execution will be aborted. Can be a shell command (exit code 0 for success) or a Lua function (returns `true` for success).
*   `next_if_fail` (string or table of strings, optional): Names of tasks to be executed if this task fails.

### Example `TaskDefinitions` Structure

```lua
TaskDefinitions = {
    my_first_group = {
        description = "An example task group.",
        tasks = {
            my_first_task = {
                name = "my_first_task",
                description = "A simple task that executes a shell command.",
                command = "echo 'Hello from Sloth-Runner!'"
            },
            my_second_task = {
                name = "my_second_task",
                description = "A task that depends on the first and uses a Lua function.",
                depends_on = "my_first_task",
                command = function(params, deps)
                    log.info("Executing the second task.")
                    -- You can access outputs from previous tasks via 'deps'
                    -- local output_from_first = deps.my_first_task.some_output
                    return true, "echo 'Second task completed!'"
                end
            }
        }
    }
}
```

## Parameters and Outputs

*   **Parameters (`params`):** Can be passed to tasks via the command line or defined within the task itself. The `command` function and `run_if`/`abort_if` functions can access them.
*   **Outputs (`deps`):** Lua `command` functions can return an outputs table. Tasks that depend on this task can access these outputs through the `deps` argument.

## Built-in Modules

Sloth-Runner exposes various Go functionalities as Lua modules, allowing your tasks to interact with the system and external services. In addition to the basic modules (`exec`, `fs`, `net`, `data`, `log`, `import`, `parallel`), Sloth-Runner now includes advanced modules for Git, Pulumi, and Salt.

These modules offer a fluent and intuitive API for complex automation.

*   **`exec` module:** For executing arbitrary shell commands.
*   **`fs` module:** For file system operations (read, write, etc.).
*   **`net` module:** For making HTTP requests and downloads.
*   **`data` module:** For parsing and serializing JSON and YAML.
*   **`log` module:** For logging messages to the Sloth-Runner console.
*   **`import` function:** For importing other Lua files and reusing tasks.
*   **`parallel` function:** For executing tasks in parallel.
*   **`git` module:** For interacting with Git repositories.
*   **`pulumi` module:** For orchestrating Pulumi stacks.
*   **`salt` module:** For executing SaltStack commands.

For details on each module, please refer to their respective sections in the documentation.

---
[English](./core-concepts.md) | [Português](../pt/core-concepts.md) | [中文](../zh/core-concepts.md)