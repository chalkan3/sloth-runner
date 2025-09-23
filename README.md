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
*   **⏰ Task Scheduler:** Automate the execution of your Lua tasks at specified intervals using cron syntax, running as a persistent background process.
*   **📝 `values.yaml` Integration:** Pass configuration values to your Lua tasks through a `values.yaml` file, similar to Helm.
*   **💻 Command-Line Interface (CLI):**
    *   `run`: Execute tasks from a Lua configuration file.
    *   `list`: List all available task groups and tasks with their descriptions and dependencies.
    *   `validate`: Validates the syntax and structure of a Lua task file.
    *   `test`: Executes a Lua test file for a task workflow.
    *   `repl`: Starts an interactive REPL session.
    *   `version`: Print the version number of sloth-runner.
    *   `template list`: Lists all available templates.
    *   `new`: Generates a new task definition file from a template.
    *   `check dependencies`: Checks for required external CLI tools.


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

## 📄 Templates

`sloth-runner` provides several templates to quickly scaffold new task definition files.

| Template Name      | Description                                                                    |
| :----------------- | :----------------------------------------------------------------------------- |
| `simple`           | Generates a single group with a 'hello world' task. Ideal for getting started. |
| `python`           | Creates a pipeline to set up a Python environment, install dependencies, and run a script. |
| `parallel`         | Demonstrates how to run multiple tasks concurrently.                           |
| `python-pulumi`    | Pipeline to deploy Pulumi infrastructure managed with Python.                  |
| `python-pulumi-salt` | Provisions infrastructure with Pulumi and configures it using SaltStack.       |
| `git-python-pulumi` | CI/CD Pipeline: Clones a repo, sets up the environment, and deploys with Pulumi. |
| `dummy`            | Generates a dummy task that does nothing.                                      |

---

## 💻 CLI Commands

`sloth-runner` provides a simple and powerful command-line interface.

### `sloth-runner run`

Executes tasks defined in a Lua template file.

**Usage:** `sloth-runner run [flags]`

**Description:**
The run command executes tasks defined in a Lua template file.
You can specify the file, environment variables, and target specific tasks or groups.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file (default: "examples/basic_pipeline.lua")
*   `-e, --env string`: Environment for the tasks (e.g., Development, Production) (default: "Development")
*   `-p, --prod`: Set to true for production environment (default: false)
*   `--shards string`: Comma-separated list of shard numbers (e.g., 1,2,3) (default: "1,2,3")
*   `-t, --tasks string`: Comma-separated list of specific tasks to run (e.g., task1,task2)
*   `-g, --group string`: Run tasks only from a specific task group
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks
*   `-d, --dry-run`: Simulate the execution of tasks without actually running them (default: false)
*   `--return`: Return the output of the target tasks as JSON (default: false)
*   `-y, --yes`: Bypass interactive task selection and run all tasks (default: false)

### `sloth-runner list`

Lists all available task groups and tasks.

**Usage:** `sloth-runner list [flags]`

**Description:**
The list command displays all task groups and their respective tasks, along with their descriptions and dependencies.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file (default: "examples/basic_pipeline.lua")
*   `-e, --env string`: Environment for the tasks (e.g., Development, Production) (default: "Development")
*   `-p, --prod`: Set to true for production environment (default: false)
*   `--shards string`: Comma-separated list of shard numbers (e.g., 1,2,3) (default: "1,2,3")
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks

### `sloth-runner validate`

Validates the syntax and structure of a Lua task file.

**Usage:** `sloth-runner validate [flags]`

**Description:**
The validate command checks a Lua task file for syntax errors and ensures that the TaskDefinitions table is correctly structured.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file (default: "examples/basic_pipeline.lua")
*   `-e, --env string`: Environment for the tasks (e.g., Development, Production) (default: "Development")
*   `-p, --prod`: Set to true for production environment (default: false)
*   `--shards string`: Comma-separated list of shard numbers (e.g., 1,2,3) (default: "1,2,3")
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks

### `sloth-runner test`

Executes a Lua test file for a task workflow.

**Usage:** `sloth-runner test -w <workflow-file> -f <test-file>`

**Description:**
The test command runs a specified Lua test file against a workflow.
Inside the test file, you can use the 'test' and 'assert' modules to validate task behaviors.

**Flags:**

*   `-f, --file string`: Path to the Lua test file (required)
*   `-w, --workflow string`: Path to the Lua workflow file to be tested (required)

### `sloth-runner repl`

Starts an interactive REPL session.

**Usage:** `sloth-runner repl [flags]`

**Description:**
The repl command starts an interactive Read-Eval-Print Loop that allows you
to execute Lua code and interact with all the built-in sloth-runner modules.
You can optionally load a workflow file to have its context available.

**Flags:**

*   `-f, --file string`: Path to a Lua workflow file to load into the REPL session

### `sloth-runner scheduler`

Manages the background task scheduler.

**Usage:** `sloth-runner scheduler [command]`

**Description:**
The scheduler command provides subcommands to enable, disable, list, and delete the sloth-runner background task scheduler.

#### `sloth-runner scheduler enable`

Starts the sloth-runner scheduler in the background.

**Usage:** `sloth-runner scheduler enable [flags]`

**Description:**
The enable command starts the sloth-runner scheduler as a persistent background process.
It will monitor and execute tasks defined in the scheduler configuration file.

**Flags:**

*   `-c, --scheduler-config string`: Path to the scheduler configuration file (default: "scheduler.yaml")

#### `sloth-runner scheduler disable`

Stops the running sloth-runner scheduler.

**Usage:** `sloth-runner scheduler disable`

**Description:**
The disable command stops the background sloth-runner scheduler process.

#### `sloth-runner scheduler list`

Lists all scheduled tasks.

**Usage:** `sloth-runner scheduler list [flags]`

**Description:**
The list command displays all scheduled tasks defined in the scheduler configuration file.

**Flags:**

*   `-c, --scheduler-config string`: Path to the scheduler configuration file (default: "scheduler.yaml")

#### `sloth-runner scheduler delete <task_name>`

Deletes a specific scheduled task.

**Usage:** `sloth-runner scheduler delete <task_name> [flags]`

**Description:**
The delete command removes a specific scheduled task from the scheduler configuration file.

**Arguments:**

*   `<task_name>`: The name of the scheduled task to delete.

**Flags:**

*   `-c, --scheduler-config string`: Path to the scheduler configuration file (default: "scheduler.yaml")

### `sloth-runner version`

Print the version number of sloth-runner.

**Usage:** `sloth-runner version`

**Description:**
All software has versions. This is sloth-runner's

### `sloth-runner template list`

Lists all available templates.

**Usage:** `sloth-runner template list`

**Description:**
Displays a table of all available templates that can be used with the 'new' command.

### `sloth-runner new <group-name>`

Generates a new task definition file from a template.

**Usage:** `sloth-runner new <group-name> [flags]`

**Description:**
The new command creates a boilerplate Lua task definition file.
You can choose from different templates and specify an output file.
Run 'sloth-runner template list' to see all available templates.

**Arguments:**

*   `<group-name>`: The name of the task group to generate.

**Flags:**

*   `-o, --output string`: Output file path (default: stdout)
*   `-t, --template string`: Template to use. See `template list` for options. (default: "simple")
*   `--set key=value`: Pass key-value pairs to the template for dynamic content generation.

### `sloth-runner check dependencies`

Checks for required external CLI tools.

**Usage:** `sloth-runner check dependencies`

**Description:**
Verifies that all external command-line tools used by the various modules (e.g., docker, aws, doctl) are installed and available in the system's PATH.

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
