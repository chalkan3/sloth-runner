# sloth-runner

A flexible and extensible task runner application written in Go, powered by Lua scripting. `sloth-runner` allows you to define complex workflows, manage task dependencies, and integrate with external systems, all through simple Lua scripts.

## Features

*   **Lua Scripting:** Define tasks and workflows using powerful and flexible Lua scripts.
*   **Dependency Management:** Specify task dependencies to ensure ordered execution of complex pipelines.
*   **Asynchronous Task Execution:** Run tasks concurrently for improved performance.
*   **Pre/Post Execution Hooks:** Define custom Lua functions to run before and after task commands.
*   **Rich Lua API:** Access system functionalities directly from Lua tasks:
    *   **`exec` module:** Execute shell commands.
    *   **`fs` module:** Perform file system operations (read, write, append, exists, mkdir, rm, rm_r, ls).
    *   **`net` module:** Make HTTP requests (GET, POST) and download files.
    *   **`data` module:** Parse and serialize JSON and YAML data.
    *   **`log` module:** Log messages with different severity levels (info, warn, error, debug).
    *   **`salt` module:** Execute SaltStack commands (`salt`, `salt-call`) directly.
*   **`values.yaml` Integration:** Pass configuration values to your Lua tasks via a `values.yaml` file, similar to Helm.
*   **Command-Line Interface (CLI):**
    *   `run`: Execute tasks from a Lua configuration file.
    *   `list`: List all available task groups and tasks with their descriptions and dependencies.

## 📚 Documentation

To help you get the most out of Sloth Runner, we've prepared detailed documentation:

-   **[Getting Started Tutorial](docs/TUTORIAL.md):** A step-by-step guide to creating your first tasks.
-   **[Lua API Reference](docs/LUA_API.md):** Detailed documentation for all the available Lua modules (`fs`, `net`, `exec`, etc.).
-   **[Examples Guide](docs/EXAMPLES.md):** An explanation of what each of the provided examples demonstrates.
-   **[Contributing Guide](CONTRIBUTING.md):** Guidelines for contributing to the Sloth Runner project.

## Getting Started

## Installation

To install `sloth-runner` on your system, you can use the provided `install.sh` script. This script automatically detects your operating system and architecture, downloads the latest release from GitHub, and places the `sloth-runner` executable in `/usr/local/bin`.

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**Note:** The `install.sh` script requires `sudo` privileges to move the executable to `/usr/local/bin`.

To run the `sloth-runner` application:

```bash
go run ./cmd/sloth-runner
```

### Basic Usage

To run a Lua task file:

```bash
go run ./cmd/sloth-runner run -f examples/basic_pipeline.lua
```

To list tasks in a file:

```bash
go run ./cmd/sloth-runner list -f examples/basic_pipeline.lua
```

## Defining Tasks in Lua

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
                post_exec = function(params, output)
                    print("Lua Hook: fetch_data completed. Raw data: " .. (output.raw_data or "N/A"))
                    return true, "fetch_data post_exec successful"
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
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: process_data preparing. Input source: " .. (input_from_dependency.fetch_data.source or "unknown"))
                    return true, "process_data pre_exec successful"
                end,
            },
            {
                name = "store_result",
                description = "Stores the final processed data",
                depends_on = "process_data", -- Dependency
                command = function(params, input_from_dependency)
                    local final_data = input_from_dependency.process_data.processed_data
                    print("Lua: Executing store_result with final data: " .. final_data)
                    return true, "echo 'Result stored'", { final_result = final_data, timestamp = os.time() }
                end,
            }
        }
    }
}
```

## Dependency Management

Tasks can declare dependencies using the `depends_on` field. This field can be a single task name (string) or a list of task names (table of strings). `sloth-runner` ensures that dependent tasks are executed only after their dependencies have completed successfully. The output of a dependency is passed as `input_from_dependency` to the dependent task's `command`, `pre_exec`, and `post_exec` functions.

Example with multiple dependencies (`examples/complex_workflow.lua`):

```lua
-- ... (Task definitions) ...
            {
                name = "generate_report",
                description = "Generates final report based on staged and enriched data",
                depends_on = {"load_to_staging", "enrich_data"}, -- Multiple dependencies
                command = function(params, input_from_dependency)
                    local staging_id = input_from_dependency.load_to_staging.staging_id
                    local enriched_info = input_from_dependency.enrich_data.enriched_info
                    print("Lua: Generating report for staging_id: " .. staging_id .. " with enriched info: " .. enriched_info .. "...")
                    return true, "echo 'Report generated'", { report_url = "http://reports.example.com/" .. staging_id .. "_" .. os.time() .. ".pdf" }
                end,
                async = false,
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: generate_report preparing. Staging ID: " .. (input_from_dependency.load_to_staging.staging_id or "N/A") .. ", Enriched Info: " .. (input_from_dependency.enrich_data.enriched_info or "N/A"))
                end,
            }
-- ...
```

## CLI Commands

`sloth-runner` provides the following command-line interface:

### `sloth-runner run`

Executes tasks defined in a Lua template file.

```bash
go run ./cmd/sloth-runner run [flags]
```

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file (default: `examples/basic_pipeline.lua`)
*   `-e, --env string`: Environment for the tasks (e.g., `Development`, `Production`) (default: `Development`)
*   `-p, --prod`: Set to true for production environment
*   `--shards string`: Comma-separated list of shard numbers (e.g., `1,2,3`) (default: `1,2,3`)
*   `-t, --tasks string`: Comma-separated list of specific tasks to run (e.g., `task1,task2`). If omitted, all tasks in the group are run.
*   `-g, --group string`: Run tasks only from a specific task group.
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks (e.g., `configs/my_values.yaml`).

### `sloth-runner list`

Lists all available task groups and tasks defined in a Lua template file.

```bash
go run ./cmd/sloth-runner list [flags]
```

**Flags:**

*   `-f, --file string`: Path to the Lua task configuration template file (default: `examples/basic_pipeline.lua`)
*   `-e, --env string`: Environment for the tasks (e.g., `Development`, `Production`) (default: `Development`)
*   `-p, --prod`: Set to true for production environment
*   `--shards string`: Comma-separated list of shard numbers (e.g., `1,2,3`) (default: `1,2,3`)
*   `-v, --values string`: Path to a YAML file with values to be passed to Lua tasks.

## Lua Extensions (API)

`sloth-runner` exposes several Go functionalities as Lua modules, allowing your tasks to interact with the system and external services.

### `exec` module

Execute shell commands.

*   `exec.command(cmd, arg1, arg2, ...)`: Executes a shell command. Returns `stdout`, `stderr`, `error_message`.

    ```lua
    local stdout, stderr, err = exec.command("ls", "-l", "/tmp")
    if err then
        log.error("Command failed: " .. err)
    else
        log.info("Output: " .. stdout)
    end
    ```

### `fs` module

Perform file system operations.

*   `fs.read(path)`: Reads file content. Returns `content`, `error_message`.
*   `fs.write(path, content)`: Writes content to a file. Returns `error_message`.
*   `fs.append(path, content)`: Appends content to a file. Returns `error_message`.
*   `fs.exists(path)`: Checks if a file/directory exists. Returns `boolean`.
*   `fs.mkdir(path)`: Creates a directory (and parents). Returns `error_message`.
*   `fs.rm(path)`: Removes a file or empty directory. Returns `error_message`.
*   `fs.rm_r(path)`: Recursively removes a directory. Returns `error_message`.
*   `fs.ls(path)`: Lists files/directories in a path. Returns `table_of_names`, `error_message`.

    ```lua
    local content, err = fs.read("/etc/hostname")
    if err then
        log.error("Failed to read file: " .. err)
    else
        log.info("Hostname: " .. content)
    end
    ```

### `net` module

Make HTTP requests and download files.

*   `net.http_get(url)`: Performs an HTTP GET request. Returns `body`, `status_code`, `headers_table`, `error_message`.
*   `net.http_post(url, body, headers_table)`: Performs an HTTP POST request. Returns `body`, `status_code`, `headers_table`, `error_message`.
*   `net.download(url, destination_path)`: Downloads a file. Returns `error_message`.

    ```lua
    local body, status, headers, err = net.http_get("https://api.github.com/zen")
    if err then
        log.error("HTTP GET failed: " .. err)
    else
        log.info("GitHub Zen: " .. body .. " (Status: " .. status .. ")")
    end
    ```

### `data` module

Parse and serialize JSON and YAML data.

*   `data.parse_json(json_string)`: Parses JSON string to Lua table. Returns `lua_table`, `error_message`.
*   `data.to_json(lua_table)`: Converts Lua table to JSON string. Returns `json_string`, `error_message`.
*   `data.parse_yaml(yaml_string)`: Parses YAML string to Lua table. Returns `lua_table`, `error_message`.
*   `data.to_yaml(lua_table)`: Converts Lua table to YAML string. Returns `yaml_string`, `error_message`.

    ```lua
    local json_str = '{"name": "sloth", "speed": "slow"}'
    local parsed_data, err = data.parse_json(json_str)
    if err then
        log.error("Failed to parse JSON: " .. err)
    else
        log.info("Parsed name: " .. parsed_data.name)
    end
    ```

### `log` module

Log messages with different severity levels.

*   `log.info(message)`
*   `log.warn(message)`
*   `log.error(message)`
*   `log.debug(message)`

    ```lua
    log.info("This is an informational message.")
    log.warn("Something might be wrong here.")
    ```

### `salt` module

Execute SaltStack commands directly.

*   `salt.cmd(command_type, arg1, arg2, ...)`: Executes a SaltStack command. `command_type` can be `"salt"` or `"salt-call"`. Arguments are passed directly to the Salt command. Returns `stdout`, `stderr`, `error_message`.

    ```lua
    -- Ping a specific minion
    local stdout, stderr, err = salt.cmd("salt", "my_minion", "test.ping")
    if err then
        log.error("Salt ping failed: " .. err)
    else
        log.info("Salt ping output: " .. stdout)
    end

    -- Run a local state.highstate
    local stdout_call, stderr_call, err_call = salt.cmd("salt-call", "state.highstate")
    if err_call then
        log.error("Salt-call failed: " .. err_call)
    end
    ```

## `values.yaml` Integration

You can pass a `values.yaml` file to your `sloth-runner` execution using the `-v` or `--values` flag. The content of this YAML file will be parsed and made available as a global Lua table named `values` within your Lua tasks.

Example `configs/my_values.yaml`:

```yaml
app:
  name: MyAwesomeApp
  version: 1.0.0
database:
  host: production.db
  port: 5432
```

Accessing values in Lua:

```lua
-- In your Lua task
if values then
    log.info("App Name from values: " .. values.app.name)
    log.info("DB Host from values: " .. values.database.host)
else
    log.warn("No values.yaml loaded.")
end
```

To run with `values.yaml`:

```bash
go run ./cmd/sloth-runner run -f examples/my_workflow.lua -v configs/my_values.yaml
```