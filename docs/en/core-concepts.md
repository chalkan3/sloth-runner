# Core Concepts

This document explains the fundamental concepts of `sloth-runner`, helping you understand how to define and orchestrate complex workflows.

---

## The `TaskDefinitions` Table

The entry point for any `sloth-runner` workflow is a Lua file that returns a global table named `TaskDefinitions`. This table is a dictionary where each key is a **Task Group** name.

```lua
-- my_pipeline.lua
TaskDefinitions = {
  -- Task Groups are defined here
}
```

---

## Task Groups

A Task Group is a collection of related tasks. It can also define properties that affect all tasks within it.

**Group Properties:**

*   `description` (string): A description of what the group does.
*   `tasks` (table): A list of individual task tables.
*   `create_workdir_before_run` (boolean): If `true`, a temporary working directory is created for the group before any task runs. This directory is passed to each task.
*   `clean_workdir_after_run` (function): A Lua function that decides if the temporary workdir should be deleted after the group finishes. It receives the final result of the group (`{success = true/false, ...}`). Returning `true` deletes the directory.

**Example:**
```lua
TaskDefinitions = {
  my_group = {
    description = "A group that manages its own temporary directory.",
    create_workdir_before_run = true,
    clean_workdir_after_run = function(result)
      if not result.success then
        log.warn("Group failed. Workdir will be kept for debugging.")
      end
      return result.success -- Only clean up if everything succeeded
    end,
    tasks = {
      -- Tasks go here
    }
  }
}
```

---

## Individual Tasks

A task is a single unit of work. It's defined as a table with several available properties to control its behavior.

### Basic Properties

*   `name` (string): The unique name of the task within its group.
*   `description` (string): A brief description of what the task does.
*   `command` (string or function): The core action of the task.
    *   **As a string:** It's executed as a shell command.
    *   **As a function:** The Lua function is executed. It receives two arguments: `params` (a table of its parameters) and `deps` (a table containing the outputs of its dependencies). The function must return:
        1.  `boolean`: `true` for success, `false` for failure.
        2.  `string`: A message describing the result.
        3.  `table` (optional): A table of outputs that other tasks can depend on.

### Dependency and Execution Flow

*   `depends_on` (string or table): A list of task names that must complete successfully before this task can run.
*   `next_if_fail` (string or table): A list of task names to run *only if* this task fails. This is useful for cleanup or notification tasks.
*   `async` (boolean): If `true`, the task runs in the background, and the runner does not wait for it to complete before starting the next task in the execution order.

### Error Handling and Robustness

*   `retries` (number): The number of times to retry a task if it fails. Default is `0`.
*   `timeout` (string): A duration (e.g., `"10s"`, `"1m"`) after which the task will be terminated if it's still running.

### Conditional Execution

*   `run_if` (string or function): The task will be skipped unless this condition is met.
    *   **As a string:** A shell command. An exit code of `0` means the condition is met.
    *   **As a function:** A Lua function that returns `true` if the task should run.
*   `abort_if` (string or function): The entire workflow will be aborted if this condition is met.
    *   **As a string:** A shell command. An exit code of `0` means abort.
    *   **As a function:** A Lua function that returns `true` to abort.

### Lifecycle Hooks

*   `pre_exec` (function): A Lua function that runs *before* the main `command`.
*   `post_exec` (function): A Lua function that runs *after* the main `command` has completed successfully.

### Reusability

*   `uses` (table): Specifies a pre-defined task from another file (loaded via `import`) to use as a base. The current task definition can then override properties like `params` or `description`.
*   `params` (table): A dictionary of key-value pairs that can be passed to the task's `command` function.

---

## Global Functions

`sloth-runner` provides global functions in the Lua environment to help orchestrate workflows.

### `import(path)`

Loads another Lua file and returns the value it returns. This is the primary mechanism for creating reusable task modules. The path is relative to the file calling `import`.

**Example (`reusable_tasks.lua`):**
```lua
-- Import a module that returns a table of task definitions
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
  main = {
    tasks = {
      {
        -- Use the 'build' task from the imported module
        uses = docker_tasks.build,
        params = { image_name = "my-app" }
      }
    }
  }
}
```

### `parallel(tasks)`

Executes a list of tasks concurrently and waits for all of them to complete.

*   `tasks` (table): A list of task tables to run in parallel.

**Example:**
```lua
command = function()
  log.info("Starting 3 tasks in parallel...")
  local results, err = parallel({
    { name = "short_task", command = "sleep 1" },
    { name = "medium_task", command = "sleep 2" },
    { name = "long_task", command = "sleep 3" }
  })
  if err then
    return false, "Parallel execution failed"
  end
  return true, "All parallel tasks finished."
end
```

### `export(table)`

Exports data from any point in a script to the CLI. When the `--return` flag is used, all exported tables are merged with the final task's output into a single JSON object.

*   `table`: A Lua table to be exported.

**Example:**
```lua
command = function()
  export({ important_value = "data from the middle of a task" })
  return true, "Task done", { final_output = "some result" }
end
```
Running with `--return` would produce:
```json
{
  "important_value": "data from the middle of a task",
  "final_output": "some result"
}
```
