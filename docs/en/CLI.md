# CLI Commands

The `sloth-runner` command-line interface (CLI) is the primary way to interact with your task pipelines. It provides commands to run, list, validate, and manage your workflows.

---

## `sloth-runner run`

Executes tasks defined in a Lua configuration file. This is the most common command you will use.

**Usage:**
```bash
sloth-runner run [flags]
```

**Flags:**

*   `-f, --file string`: **(Required)** Path to the Lua task configuration file.
*   `-g, --group string`: Run tasks only from a specific task group. If not provided, `sloth-runner` will run tasks from all groups.
*   `-t, --tasks string`: A comma-separated list of specific tasks to run (e.g., `task1,task2`). If not provided, all tasks in the specified group (or all groups) will be considered.
*   `-v, --values string`: Path to a YAML file with values to be passed to your Lua scripts. These values are accessible in Lua via the global `values` table.
*   `-d, --dry-run`: Simulates the execution of tasks. It will print the tasks that would be run and in what order, but will not execute their `command`.
*   `--return`: Prints the final output of the executed tasks as a JSON object to stdout. This includes both the return value of the last task and any data passed to the global `export()` function.
*   `-y, --yes`: Bypasses the interactive task selection prompt when no specific tasks are provided with `-t`.
*   `--interactive`: Enable interactive mode for task execution, prompting for user input before each task.

**Examples:**

*   Run all tasks in a specific group:
    ```bash
    sloth-runner run -f examples/basic_pipeline.lua -g my_group
    ```
*   Run a single, specific task:
    ```bash

    sloth-runner run -f examples/basic_pipeline.lua -g my_group -t my_task
    ```
*   Run multiple tasks and get their combined output as JSON:
    ```bash
    sloth-runner run -f examples/export_example.lua -t export-data-task --return
    ```

---

## `sloth-runner list`

Lists all available task groups and tasks defined in a Lua configuration file, along with their descriptions and dependencies.

**Usage:**
```bash
sloth-runner list [flags]
```

**Flags:**

*   `-f, --file string`: **(Required)** Path to the Lua task configuration file.
*   `-v, --values string`: Path to a YAML values file, in case your task definitions depend on it.

---

## `sloth-runner new`

Generates a new boilerplate Lua task definition file from a template.

**Usage:**
```bash
sloth-runner new <group-name> [flags]
```

**Arguments:**

*   `<group-name>`: The name of the main task group to be created in the file.

**Flags:**

*   `-t, --template string`: The template to use. Default is `simple`. Run `sloth-runner template list` to see all available options.
*   `-o, --output string`: The path to the output file. If not provided, the generated content will be printed to stdout.
*   `--set key=value`: Pass key-value pairs to the template for dynamic content generation.

**Example:**
```bash
sloth-runner new my-python-pipeline -t python -o my_pipeline.lua
```

---

## `sloth-runner validate`

Validates the syntax and basic structure of a Lua task file without executing any tasks.

**Usage:**
```bash
sloth-runner validate [flags]
```

**Flags:**

*   `-f, --file string`: **(Required)** Path to the Lua task configuration file to validate.
*   `-v, --values string`: Path to a YAML values file, if needed for validation.

---

## `sloth-runner test`

Executes a Lua-based test file against a workflow file. (This is an advanced feature).

**Usage:**
```bash
sloth-runner test [flags]
```

**Flags:**

*   `-w, --workflow string`: **(Required)** Path to the Lua workflow file to be tested.
*   `-f, --file string`: **(Required)** Path to the Lua test file.

---

## `sloth-runner template list`

Lists all available templates that can be used with the `sloth-runner new` command.

**Usage:**
```bash
sloth-runner template list
```

---

## `sloth-runner artifacts`

Manages task artifacts, which are files or directories produced by tasks.

**Subcommands:**

*   `sloth-runner artifacts list`: Lists all collected artifacts.
*   `sloth-runner artifacts get <artifact_path>`: Downloads a specific artifact.
*   `sloth-runner artifacts clean`: Cleans up old or unwanted artifacts.

---

### `sloth-runner version`

Displays the current version of `sloth-runner`.

```bash
sloth-runner version
```

### `sloth-runner scheduler`

Manages the `sloth-runner` task scheduler, allowing you to enable, disable, list, and delete scheduled tasks.

For detailed information on scheduler commands and configuration, refer to the [Task Scheduler documentation](scheduler.md).

**Subcommands:**

*   `sloth-runner scheduler enable`: Starts the scheduler as a background process.
*   `sloth-runner scheduler disable`: Stops the running scheduler process.
*   `sloth-runner scheduler list`: Lists all configured scheduled tasks.
*   `sloth-runner scheduler delete <task_name>`: Deletes a specific scheduled task.

