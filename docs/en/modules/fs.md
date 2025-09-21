# FS Module

The `fs` module provides essential functions for interacting with the file system directly from your Lua scripts.

---

## `fs.read(path)`

Reads the entire content of a file.

*   **Parameters:**
    *   `path` (string): The path to the file.
*   **Returns:**
    *   `string`: The content of the file.
    *   `error`: An error object if the read fails.

---

## `fs.write(path, content)`

Writes content to a file, overwriting it if it already exists.

*   **Parameters:**
    *   `path` (string): The path to the file.
    *   `content` (string): The content to write.
*   **Returns:**
    *   `error`: An error object if the write fails.

---

## `fs.append(path, content)`

Appends content to the end of a file. Creates the file if it doesn't exist.

*   **Parameters:**
    *   `path` (string): The path to the file.
    *   `content` (string): The content to append.
*   **Returns:**
    *   `error`: An error object if the append fails.

---

## `fs.exists(path)`

Checks if a file or directory exists at the given path.

*   **Parameters:**
    *   `path` (string): The path to check.
*   **Returns:**
    *   `boolean`: `true` if the path exists, `false` otherwise.

---

## `fs.mkdir(path)`

Creates a directory at the given path, including any necessary parent directories (like `mkdir -p`).

*   **Parameters:**
    *   `path` (string): The directory path to create.
*   **Returns:**
    *   `error`: An error object if the creation fails.

---

## `fs.rm(path)`

Removes a single file.

*   **Parameters:**
    *   `path` (string): The path to the file to remove.
*   **Returns:**
    *   `error`: An error object if the removal fails.

---

## `fs.rm_r(path)`

Removes a file or directory recursively (like `rm -rf`).

*   **Parameters:**
    *   `path` (string): The path to remove.
*   **Returns:**
    *   `error`: An error object if the removal fails.

---

## `fs.ls(path)`

Lists the contents of a directory.

*   **Parameters:**
    *   `path` (string): The path to the directory.
*   **Returns:**
    *   `table`: A table containing the names of files and subdirectories.
    *   `error`: An error object if the listing fails.

---

## `fs.tmpname()`

Generates a unique temporary directory path. Note: This function only returns the name; it does not create the directory.

*   **Returns:**
    *   `string`: A unique path suitable for a temporary directory.
    *   `error`: An error object if a name could not be generated.

### Example

```lua
command = function()
  local fs = require("fs")
  
  local tmp_dir = "/tmp/fs-example"
  log.info("Creating directory: " .. tmp_dir)
  fs.mkdir(tmp_dir)

  local file_path = tmp_dir .. "/my_file.txt"
  log.info("Writing to file: " .. file_path)
  fs.write(file_path, "Hello, Sloth Runner!\n")

  log.info("Appending to file...")
  fs.append(file_path, "This is a new line.")

  if fs.exists(file_path) then
    log.info("File content: " .. fs.read(file_path))
  end

  log.info("Listing contents of " .. tmp_dir)
  local contents = fs.ls(tmp_dir)
  for i, name in ipairs(contents) do
    print("- " .. name)
  end

  log.info("Cleaning up...")
  fs.rm_r(tmp_dir)
  
  return true, "FS module operations successful."
end
```
