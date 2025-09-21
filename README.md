[English](./README.md) | [Português](./README.pt.md) | [中文](./README.zh.md)

# 🦥 Sloth Runner 🚀

A flexible and extensible task runner application written in Go, powered by Lua scripting. `sloth-runner` allows you to define complex workflows, manage task dependencies, and integrate with external systems, all through simple Lua scripts.

[![Go CI](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml/badge.svg)](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml)

---

## ✨ Features

*   **📜 Lua Scripting:** Define tasks and workflows using the power and flexibility of Lua scripts.
*   **🔗 Dependency Management:** Specify dependencies between tasks to ensure the ordered execution of complex pipelines.
*   **⚡ Asynchronous Task Execution:** Run tasks concurrently for better performance.
*   **🪝 Pre/Post-Execution Hooks:** Define custom Lua functions to be executed before and after task commands.
*   **⚙️ Rich Lua API:** Access system functionalities directly from your Lua tasks:
    *   **`exec` module:** Execute shell commands.
    *   **`fs` module:** Perform file system operations (read, write, append, check existence, create directory, remove, remove recursively, list).
    *   **`net` module:** Make HTTP requests (GET, POST) and download files.
    *   **`data` module:** Parse and serialize data in JSON and YAML format.
    *   **`log` module:** Log messages with different severity levels (info, warn, error, debug).
    *   **`salt` module:** Execute SaltStack commands (`salt`, `salt-call`) directly.
    *   **`gcp` module:** Execute Google Cloud (`gcloud`) CLI commands.
*   **📝 `values.yaml` Integration:** Pass configuration values to your Lua tasks through a `values.yaml` file, similar to Helm.
*   **💻 Command-Line Interface (CLI):**
    *   `run`: Execute tasks from a Lua configuration file.
    *   `list`: List all available task groups and tasks with their descriptions and dependencies.


## 📚 Complete Documentation

For more detailed documentation, usage guides, and advanced examples, visit our [Complete Documentation](./docs/index.md).

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

To list the tasks in a file:

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

---

## 📜 Defining Tasks in Lua

Tasks are defined in Lua files, typically within a `TaskDefinitions` table. Each task can have a `name`, `description`, `command` (either a string for a shell command or a Lua function), `async` (boolean), `pre_exec` (Lua function hook), `post_exec` (Lua function hook), and `depends_on` (a string or a table of strings).

Example (`examples/basic_pipeline.lua`):

```lua
-- Import reusable tasks from another file. The path is relative.
local docker_tasks = import("examples/shared/docker.lua")

TaskDefinitions = {
    full_pipeline_demo = {
        description = "A comprehensive pipeline demonstrating various features.",
        tasks = {
            -- Task 1: Fetches data, runs asynchronously.
            fetch_data = {
                name = "fetch_data",
                description = "Fetches raw data from an API.",
                async = true,
                command = function(params)
                    log.info("Fetching data...")
                    -- Simulates an API call
                    return true, "echo 'Fetched raw data'", { raw_data = "api_data" }
                end,
            },

            -- Task 2: A flaky task that retries on failure.
            flaky_task = {
                name = "flaky_task",
                description = "This task fails intermittently and will retry.",
                retries = 3,
                command = function()
                    if math.random() > 0.5 then
                        log.info("Flaky task succeeded.")
                        return true, "echo 'Success!'"
                    else
                        log.error("Flaky task failed, will retry...")
                        return false, "Random failure"
                    end
                end,
            },

            -- Task 3: Processes data, depends on the successful completion of fetch_data and flaky_task.
            process_data = {
                name = "process_data",
                description = "Processes the fetched data.",
                depends_on = { "fetch_data", "flaky_task" },
                command = function(params, deps)
                    local raw_data = deps.fetch_data.raw_data
                    log.info("Processing data: " .. raw_data)
                    return true, "echo 'Processed data'", { processed_data = "processed_" .. raw_data }
                end,
            },

            -- Task 4: A long-running task with a timeout.
            long_running_task = {
                name = "long_running_task",
                description = "A task that will be terminated if it runs too long.",
                timeout = "5s",
                command = "echo 'Starting long task...'; sleep 10; echo 'This will not be printed.';",
            },

            -- Task 5: A cleanup task that runs if the long_running_task fails.
            cleanup_on_fail = {
                name = "cleanup_on_fail",
                description = "Runs only if the long-running task fails.",
                next_if_fail = "long_running_task",
                command = "echo 'Cleanup task executed due to previous failure.'",
            },

            -- Task 6: Uses a reusable task from the imported docker.lua module.
            build_image = {
                uses = docker_tasks.build,
                description = "Builds the application's Docker image.",
                params = {
                    image_name = "my-awesome-app",
                    tag = "v1.2.3",
                    context = "./app_context"
                }
            },

            -- Task 7: A conditional task that only runs if a file exists.
            conditional_deploy = {
                name = "conditional_deploy",
                description = "Deploys the application only if the build artifact exists.",
                depends_on = "build_image",
                run_if = "test -f ./app_context/artifact.txt", -- Shell command condition
                command = "echo 'Deploying application...'",
            },

            -- Task 8: This task will abort the entire workflow if a condition is met.
            gatekeeper_check = {
                name = "gatekeeper_check",
                description = "Aborts the workflow if a critical condition is not met.",
                abort_if = function(params, deps)
                    -- Lua function condition
                    log.warn("Checking gatekeeper condition...")
                    if params.force_proceed ~= "true" then
                        log.error("Gatekeeper check failed. Aborting workflow.")
                        return true -- Abort
                    end
                    return false -- Do not abort
                end,
                command = "echo 'This command will not be executed if aborted.'"
            }
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
                        return false, "A random failure occurred"
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

You can control task execution based on conditions using `run_if` and `abort_if`. These can be a shell command or a Lua function.

*   `run_if`: The task will only be executed if the condition is met.
*   `abort_if`: The entire execution will be aborted if the condition is met.

#### Using Shell Commands

The shell command is executed, and its exit code determines the result. An exit code of `0` means the condition was met (success).

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
                command = "echo 'The conditional task is running because the condition was met.'"
            },
            {
                name = "check_abort_condition",
                description = "This task will be aborted if a specific file exists.",
                abort_if = "test -f /tmp/sloth_runner_abort_condition",
                command = "echo 'This will not be executed if the abort condition is met.'"
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
                description = "This task provides the output for the conditional task.",
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
                        log.info("Condition met, the task will run.")
                        return true
                    end
                    log.info("Condition not met, the task will be skipped.")
                    return false
                end,
                command = "echo 'The conditional task is running because the function returned true.'"
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
                        log.info("Abort condition met, execution will be stopped.")
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

You can create libraries of reusable tasks and import them into your main workflow file. This is useful for sharing common tasks (like building Docker images, deploying applications, etc.) across multiple projects.

The global `import()` function loads another Lua file and returns the value it returns. The path is resolved relative to the file calling `import`.

**How it works:**
1.  Create a module (e.g., `shared/docker.lua`) that defines a table of tasks and returns it.
2.  In your main file, call `import("shared/docker.lua")` to load the module.
3.  Reference the imported tasks in your main `TaskDefinitions` table using the `uses` field. `sloth-runner` will automatically merge the imported task with any local overrides you provide (like `description` or `params`).

<details>
<summary>Module Example (`examples/shared/docker.lua`):</summary>

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
<summary>Usage Example (`examples/reusable_tasks.lua`):</summary>

```lua
-- examples/reusable_tasks.lua

-- Import the reusable Docker tasks.
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
    app_deployment = {
        description = "A workflow that uses a reusable Docker module.",
        tasks = {
            -- Use the 'build' task from the module and override its parameters.
            build = {
                uses = docker_tasks.build,
                description = "Builds the main application Docker image",
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

*   `-f, --file string`: Path to the Lua task configuration file.
*   `-t, --tasks string`: Comma-separated list of specific tasks to run.
*   `-g, --group string`: Run tasks only from a specific task group.
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks.
*   `-d, --dry-run`: Simulate the execution of tasks without actually running them.

### `sloth-runner list`

Lists all available task groups and tasks defined in a Lua template file.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration file.
*   `-v, --values string`: Path to a YAML file with values.

---

## ⚙️ Lua API

`sloth-runner` exposes several Go functionalities as Lua modules, allowing your tasks to interact with the system and external services.

*   **`exec` module:** Execute shell commands.
*   **`fs` module:** Perform file system operations.
*   **`net` module:** Make HTTP requests and download files.
*   **`data` module:** Parse and serialize data in JSON and YAML format.
*   **`log` module:** Log messages with different severity levels.
*   **`salt` module:** Execute SaltStack commands.

For detailed API usage, please refer to the examples in the `/examples` directory.
