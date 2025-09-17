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

## 💻 CLI Commands

`sloth-runner` provides a simple and powerful command-line interface.

### `sloth-runner run`

Executes tasks defined in a Lua template file.

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file.
*   `-t, --tasks string`: Comma-separated list of specific tasks to run.
*   `-g, --group string`: Run tasks only from a specific task group.
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks.

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
