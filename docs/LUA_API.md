# 🦥 Sloth Runner - Lua API Reference ⚙️

This document provides a detailed reference for the Lua modules exposed by `sloth-runner`. These modules allow your tasks to interact with the underlying system, manage data, and control execution flow.

---

## `exec` Module

The `exec` module allows you to execute external shell commands.

### `exec.command(command, [arg1, arg2, ...])`

Executes a shell command with the given arguments.

-   **Parameters:**
    -   `command` (string): The command to execute (e.g., `"ls"`, `"docker"`).
    -   `arg...` (string, optional): A variable number of string arguments for the command.
-   **Returns:**
    -   `stdout` (string): The standard output of the command.
    -   `stderr` (string): The standard error output of the command.
    -   `err` (string or nil): An error message if the command fails to execute, otherwise `nil`.

**Example:**

```lua
log.info("Listing files in /tmp...")
local stdout, stderr, err = exec.command("ls", "-la", "/tmp")

if err then
    log.error("Failed to list files: " .. stderr)
else
    log.info("Files found:\n" .. stdout)
end
```

---

## `fs` Module

The `fs` module provides functions for interacting with the file system.

### `fs.read(path)`

Reads the entire content of a file.

-   **Parameters:**
    -   `path` (string): The path to the file.
-   **Returns:**
    -   `content` (string or nil): The content of the file, or `nil` if an error occurs.
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

### `fs.write(path, content)`

Writes a string to a file, overwriting it if it exists.

-   **Parameters:**
    -   `path` (string): The path to the file.
    -   `content` (string): The content to write.
-   **Returns:**
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

### `fs.append(path, content)`

Appends a string to the end of a file, creating it if it doesn't exist.

-   **Parameters:**
    -   `path` (string): The path to the file.
    -   `content` (string): The content to append.
-   **Returns:**
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

### `fs.exists(path)`

Checks if a file or directory exists at the given path.

-   **Parameters:**
    -   `path` (string): The path to check.
-   **Returns:**
    -   `exists` (boolean): `true` if the path exists, `false` otherwise.

### `fs.mkdir(path)`

Creates a directory, including any necessary parent directories.

-   **Parameters:**
    -   `path` (string): The directory path to create.
-   **Returns:**
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

### `fs.rm(path)`

Removes a file or an empty directory.

-   **Parameters:**
    -   `path` (string): The path to remove.
-   **Returns:**
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

### `fs.rm_r(path)`

Recursively removes a directory and all its contents.

-   **Parameters:**
    -   `path` (string): The path to the directory to remove.
-   **Returns:**
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

### `fs.ls(path)`

Lists the names of files and directories inside a given path.

-   **Parameters:**
    -   `path` (string): The path to the directory.
-   **Returns:**
    -   `files` (table or nil): A Lua table (array) of file and directory names, or `nil` on error.
    -   `err` (string or nil): An error message on failure, otherwise `nil`.

**Example:**

```lua
local dir = "/tmp/sloth-test"
fs.mkdir(dir)
fs.write(dir .. "/hello.txt", "Hello from Sloth! 🦥")
local files, err = fs.ls(dir)
if err then
    log.error("Could not list files: " .. err)
else
    log.info("Files in " .. dir .. ": " .. data.to_json(files))
end
fs.rm_r(dir)
```

---

## `net` Module

The `net` module provides networking utilities.

### `net.http_get(url)`

Performs an HTTP GET request.

-   **Parameters:**
    -   `url` (string): The URL to request.
-   **Returns:**
    -   `body` (string or nil): The response body.
    -   `status_code` (number): The HTTP status code (e.g., `200`).
    -   `headers` (table or nil): A Lua table of response headers.
    -   `err` (string or nil): An error message on failure.

### `net.http_post(url, body, [headers])`

Performs an HTTP POST request.

-   **Parameters:**
    -   `url` (string): The URL to post to.
    -   `body` (string): The request body.
    -   `headers` (table, optional): A Lua table of request headers.
-   **Returns:**
    -   `body` (string or nil): The response body.
    -   `status_code` (number): The HTTP status code.
    -   `headers` (table or nil): A Lua table of response headers.
    -   `err` (string or nil): An error message on failure.

### `net.download(url, destination_path)`

Downloads a file from a URL to a local path.

-   **Parameters:**
    -   `url` (string): The URL of the file to download.
    -   `destination_path` (string): The local path to save the file.
-   **Returns:**
    -   `err` (string or nil): An error message on failure.

**Example:**

```lua
log.info("Fetching a random cat fact...")
local body, status, _, err = net.http_get("https://catfact.ninja/fact")
if err or status ~= 200 then
    log.error("Failed to fetch cat fact: " .. (err or "status " .. status))
else
    local fact_data, json_err = data.parse_json(body)
    if json_err then
        log.error("Could not parse cat fact JSON: " .. json_err)
    else
        log.info("🐱 Cat Fact: " .. fact_data.fact)
    end
end
```

---

## `data` Module

The `data` module provides functions for data serialization and deserialization.

### `data.to_json(table)`

Converts a Lua table to a JSON string.

-   **Parameters:**
    -   `table` (table): The Lua table to convert.
-   **Returns:**
    -   `json_string` (string or nil): The resulting JSON string.
    -   `err` (string or nil): An error message on failure.

### `data.parse_json(json_string)`

Parses a JSON string into a Lua table.

-   **Parameters:**
    -   `json_string` (string): The JSON string to parse.
-   **Returns:**
    -   `table` (table or nil): The resulting Lua table.
    -   `err` (string or nil): An error message on failure.

### `data.to_yaml(table)`

Converts a Lua table to a YAML string.

-   **Parameters:**
    -   `table` (table): The Lua table to convert.
-   **Returns:**
    -   `yaml_string` (string or nil): The resulting YAML string.
    -   `err` (string or nil): An error message on failure.

### `data.parse_yaml(yaml_string)`

Parses a YAML string into a Lua table.

-   **Parameters:**
    -   `yaml_string` (string): The YAML string to parse.
-   **Returns:**
    -   `table` (table or nil): The resulting Lua table.
    -   `err` (string or nil): An error message on failure.

---

## `log` Module

The `log` module provides simple logging functions.

### `log.info(message)`
### `log.warn(message)`
### `log.error(message)`
### `log.debug(message)`

-   **Parameters:**
    -   `message` (string): The message to log.

**Example:**

```lua
log.info("Starting the task.")
log.warn("This is a warning.")
log.error("Something went wrong!")
log.debug("Here is some debug info.")
```

---

## `salt` Module

The `salt` module allows for direct execution of SaltStack commands.

### `salt.cmd(command_type, [arg1, arg2, ...])`

Executes a SaltStack command.

-   **Parameters:**
    -   `command_type` (string): The type of command, either `"salt"` or `"salt-call"`.
    -   `arg...` (string, optional): A variable number of string arguments for the command.
-   **Returns:**
    -   `stdout` (string): The standard output of the command.
    -   `stderr` (string): The standard error output of the command.
    -   `err` (string or nil): An error message if the command fails, otherwise `nil`.

**Example:**

```lua
-- Ping all minions
local stdout, stderr, err = salt.cmd("salt", "*", "test.ping")
if err then
    log.error("Salt command failed: " .. stderr)
else
    log.info("Salt ping result:\n" .. stdout)
end
```

