# 📚 Examples Guide

This guide provides a brief overview of each example file located in the `/examples` directory. These examples are designed to showcase the various features of Sloth Runner.

---

### `api_data_manipulation.lua`

Demonstrates the power of the `net` and `data` modules.
-   Fetches data from a public API (`http_get`).
-   Parses the JSON response (`data.parse_json`).
-   Extracts information and uses it in a subsequent command.

---

### `basic_pipeline.lua`

The "Hello World" of Sloth Runner.
-   Shows a simple, linear dependency chain (`task C` depends on `task B`, which depends on `task A`).
-   Demonstrates passing output from one task as input to the next.

---

### `complex_workflow.lua`

Illustrates more advanced dependency management.
-   A task that depends on multiple other tasks simultaneously.
-   Mixing of asynchronous (`async = true`) and synchronous tasks in the same workflow.

---

### `data_test.lua`

Focuses on the `data` module.
-   Shows how to parse and stringify both JSON and YAML within a Lua script.
-   Useful for tasks that need to read or generate configuration files.

---

### `dynamic_workflow.gotmpl`

This is a Go Template file, not a pure Lua file. It showcases the pre-execution templating feature.
-   Uses Go template syntax (`{{.IsProduction}}`, `{{range .Shards}}`) to dynamically generate the final Lua script.
-   Allows you to create different workflows based on flags like `--prod` or `--shards`.

---

### `exec_test.lua`

Focuses on the `exec` module.
-   Provides a simple example of how to run an external shell command (`ls -la`) and print its output.

---

### `output_manipulation_pipeline.lua`

A practical example of using the `fs` and `exec` modules together.
-   The first task (`get_file_list`) runs `find` to get a list of Go files and processes the `stdout` into a Lua table.
-   The second task (`count_go_files`) receives this table as input and counts the number of files.

---

### `salt_integration.lua`

Demonstrates the direct integration with SaltStack using the `salt` module.
-   Shows how to run `salt` and `salt-call` commands from within a task.
-   Useful for orchestrating infrastructure management tasks.

---

### `values_test.lua`

Shows how to use external configuration files.
-   Demonstrates how a `values.yaml` file passed via the `-v` flag can be accessed globally in Lua via the `values` table.
-   Allows you to separate your task logic from your configuration data.
