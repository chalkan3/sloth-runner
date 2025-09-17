# 🦥 Sloth Runner 🚀

A flexible and extensible task runner application written in Go, powered by Lua scripting. `sloth-runner` allows you to define complex workflows, manage task dependencies, and integrate with external systems, all through simple Lua scripts.

[![Go CI](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml/badge.svg)](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml)

---

## ✨ Features

*   **📜 Lua Scripting:** Define tasks and workflows using powerful and flexible Lua scripts.
*   **🔗 Dependency Management:** Specify task dependencies to ensure ordered execution of complex pipelines.
*   **⚡ Asynchronous Task Execution:** Run tasks concurrently for improved performance.
*   **🪝 Pre/Post Execution Hooks:** Define custom Lua functions to run before and after task commands.
*   **⚙️ Rich Lua API:** Access system functionalities directly from Lua tasks:
    *   **`exec` module:** Execute shell commands.
    *   **`fs` module:** Perform file system operations (read, write, append, exists, mkdir, rm, rm_r, ls).
    *   **`net` module:** Make HTTP requests (GET, POST) and download files.
    *   **`data` module:** Parse and serialize JSON and YAML data.
    *   **`log` module:** Log messages with different severity levels (info, warn, error, debug).
    *   **`salt` module:** Execute SaltStack commands (`salt`, `salt-call`) directly.
*   **📝 `values.yaml` Integration:** Pass configuration values to your Lua tasks via a `values.yaml` file, similar to Helm.
*   **💻 Command-Line Interface (CLI):**
    *   `run`: Execute tasks from a Lua configuration file.
    *   `list`: List all available task groups and tasks with their descriptions and dependencies.


## 📚 Documentation

To help you get the most out of Sloth Runner, we've prepared detailed documentation:

-   **[Getting Started Tutorial](docs/TUTORIAL.md):** A step-by-step guide to creating your first tasks.
-   **[Lua API Reference](docs/LUA_API.md):** Detailed documentation for all the available Lua modules (`fs`, `net`, `exec`, etc.).
-   **[Examples Guide](docs/EXAMPLES.md):** An explanation of what each of the provided examples demonstrates.
-   **[Contributing Guide](CONTRIBUTING.md):** Guidelines for contributing to the Sloth Runner project.

---

## 🚀 Getting Started

### Installation

To install `sloth-runner` on your system, you can use the provided `install.sh` script. This script automatically detects your operating system and architecture, downloads the latest release from GitHub, and places the `sloth-runner` executable in `/usr/local/bin`.

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**Note:** The `install.sh` script requires `sudo` privileges to move the executable to `/usr/local/bin`.

### Basic Usage

To run a Lua task file:

```bash
sloth-runner run -f examples/basic_pipeline.lua
```

To list tasks in a file:

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

---

## 📜 Defining Tasks in Lua

Tasks are defined in Lua files, typically within a `TaskDefinitions` table. Each task can have a `name`, `description`, `command` (either a string for shell command or a Lua function), `async` (boolean), `pre_exec` (Lua function hook), `post_exec` (Lua function hook), and `depends_on` (string or table of strings).

Example (`examples/basic_pipeline.lua`):

```lua
TaskDefinitions = {
    basic_pipeline = {
        description = "A simple data processing pipeline",
        tasks = {
            {
                name = "fetch_data",
                description = "Simulates fetching raw data",
                command = function(params)
                    print("Lua: Executing fetch_data...")
                    return true, "echo 'Fetched raw data'", { raw_data = "some_data_from_api", source = "external_api" }
                end,
            },
            {
                name = "process_data",
                description = "Processes the raw data",
                depends_on = "fetch_data", -- Dependency
                command = function(params, input_from_dependency)
                    local raw_data = input_from_dependency.fetch_data.raw_data
                    print("Lua: Executing process_data with input: " .. raw_data)
                    return true, "echo 'Processed data'", { processed_data = "processed_" .. raw_data, status = "success" }
                end,
            },
        }
    }
}
```

---

## Advanced Features

`sloth-runner` provides several advanced features for fine-grained control over task execution.

### Task Retries and Timeouts

You can make your workflows more robust by specifying retries for flaky tasks and timeouts for long-running ones.

*   `retries`: The number of times to retry a task if it fails.
*   `timeout`: A duration string (e.g., "10s", "1m") after which a task will be terminated.

<details>
<summary>Example (`examples/retries_and_timeout.lua`):</summary>

```lua
TaskDefinitions = {
    robust_workflow = {
        description = "A workflow to demonstrate retries and timeouts",
        tasks = {
            {
                name = "flaky_task",
                description = "This task fails 50% of the time",
                retries = 3,
                command = function()
                    if math.random() < 0.5 then
                        log.error("Simulating a random failure!")
                        return false, "Random failure occurred"
                    end
                    return true, "echo 'Flaky task succeeded!'", { result = "success" }
                end
            },
            {
                name = "long_running_task",
                description = "This task simulates a long process that will time out",
                timeout = "2s",
                command = "sleep 5 && echo 'This should not be printed'"
            }
        }
    }
}
```
</details>

### Conditional Execution: `run_if` and `abort_if`

You can control task execution based on conditions using `run_if` and `abort_if`. These can be either a shell command or a Lua function.

*   `run_if`: The task will only be executed if the condition is met.
*   `abort_if`: The entire execution will be aborted if the condition is met.

#### Using Shell Commands

The shell command is executed, and its exit code determines the outcome. A `0` exit code means the condition is met (success).

<details>
<summary>Example (`examples/conditional_execution.lua`):</summary>

```lua
TaskDefinitions = {
    conditional_workflow = {
        description = "A workflow to demonstrate conditional execution with run_if and abort_if.",
        tasks = {
            {
                name = "check_condition_for_run",
                description = "This task creates a file that the next task checks for.",
                command = "touch /tmp/sloth_runner_run_condition"
            },
            {
                name = "conditional_task",
                description = "This task only runs if the condition file exists.",
                depends_on = "check_condition_for_run",
                run_if = "test -f /tmp/sloth_runner_run_condition",
                command = "echo 'Conditional task is running because the condition was met.'"
            },
            {
                name = "check_abort_condition",
                description = "This task will abort if a specific file exists.",
                abort_if = "test -f /tmp/sloth_runner_abort_condition",
                command = "echo 'This will not run if the abort condition is met.'"
            }
        }
    }
}
```
</details>

#### Using Lua Functions

For more complex logic, you can use a Lua function. The function receives the task's `params` and the `deps` (outputs from dependencies). It must return `true` for the condition to be met.

<details>
<summary>Example (`examples/conditional_functions.lua`):</summary>

```lua
TaskDefinitions = {
    conditional_functions_workflow = {
        description = "A workflow to demonstrate conditional execution with Lua functions.",
        tasks = {
            {
                name = "setup_task",
                description = "This task provides output for the conditional task.",
                command = function()
                    return true, "Setup complete", { should_run = true }
                end
            },
            {
                name = "conditional_task_with_function",
                description = "This task only runs if the run_if function returns true.",
                depends_on = "setup_task",
                run_if = function(params, deps)
                    log.info("Checking run_if condition for conditional_task_with_function...")
                    if deps.setup_task and deps.setup_task.should_run == true then
                        log.info("Condition met, task will run.")
                        return true
                    end
                    log.info("Condition not met, task will be skipped.")
                    return false
                end,
                command = "echo 'Conditional task is running because the function returned true.'"
            },
            {
                name = "abort_task_with_function",
                description = "This task will abort the execution if the abort_if function returns true.",
                params = {
                    abort_execution = "true"
                },
                abort_if = function(params, deps)
                    log.info("Checking abort_if condition for abort_task_with_function...")
                    if params.abort_execution == "true" then
                        log.info("Abort condition met, execution will stop.")
                        return true
                    end
                    log.info("Abort condition not met.")
                    return false
                end,
                command = "echo 'This should not be executed.'"
            }
        }
    }
}
```
</details>

### Reusable Task Modules with `import`

You can create reusable libraries of tasks and import them into your main workflow file. This is useful for sharing common tasks (like building Docker images, deploying applications, etc.) across multiple projects.

The global `import()` function loads another Lua file and returns the value it returns. The path is resolved relative to the file calling `import`.

**How it works:**
1.  Create a module (e.g., `shared/docker.lua`) that defines a table of tasks and returns it.
2.  In your main file, call `import("shared/docker.lua")` to load the module.
3.  Reference the imported tasks in your main `TaskDefinitions` table using a `uses` field. `sloth-runner` will automatically merge the imported task with any local overrides you provide (like `description` or `params`).

<details>
<summary>Example Module (`examples/shared/docker.lua`):</summary>

```lua
-- examples/shared/docker.lua
-- A reusable module for Docker tasks.

local TaskDefinitions = {
    build = {
        name = "build",
        description = "Builds a Docker image",
        params = {
            tag = "latest",
            dockerfile = "Dockerfile",
            context = "."
        },
        command = function(params)
            local image_name = params.image_name or "my-default-image"
            -- ... build command logic ...
            local cmd = string.format("docker build -t %s:%s -f %s %s", image_name, params.tag, params.dockerfile, params.context)
            return true, cmd
        end
    },
    push = {
        name = "push",
        description = "Pushes a Docker image to a registry",
        -- ... push task logic ...
    }
}

return TaskDefinitions
```
</details>

<details>
<summary>Example Usage (`examples/reusable_tasks.lua`):</summary>

```lua
-- examples/reusable_tasks.lua

-- Import the reusable Docker tasks.
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
    app_deployment = {
        description = "A workflow that uses a reusable Docker module.",
        tasks = {
            -- Use the 'build' task from the module and override its params.
            build = {
                uses = docker_tasks.build,
                description = "Build the main application Docker image",
                params = {
                    image_name = "my-app",
                    tag = "v1.0.0",
                    context = "./app"
                }
            },
            
            -- A regular task that depends on the imported 'build' task.
            deploy = {
                name = "deploy",
                description = "Deploys the application",
                depends_on = "build",
                command = "echo 'Deploying...'"
            }
        }
    }
}
```
</details>

---

## 💻 CLI Commands

`sloth-runner` provides a simple and powerful command-line interface.

### `sloth-runner run`

Executes tasks defined in a Lua template file.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file.
*   `-t, --tasks string`: Comma-separated list of specific tasks to run.
*   `-g, --group string`: Run tasks only from a specific task group.
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks.
*   `-d, --dry-run`: Simulate the execution of tasks without actually running them.

### `sloth-runner list`

Lists all available task groups and tasks defined in a Lua template file.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file.
*   `-v, --values string`: Path to a YAML file with values.

---

## ⚙️ Lua API

`sloth-runner` exposes several Go functionalities as Lua modules, allowing your tasks to interact with the system and external services.

*   **`exec` module:** Execute shell commands.
*   **`fs` module:** Perform file system operations.
*   **`net` module:** Make HTTP requests and download files.
*   **`data` module:** Parse and serialize JSON and YAML data.
*   **`log` module:** Log messages with different severity levels.
*   **`salt` module:** Execute SaltStack commands.

For detailed API usage, please refer to the examples in the `/examples` directory.