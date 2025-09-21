# Python Module

The `python` module provides a convenient way to manage Python virtual environments (`venv`) and execute scripts from within your `sloth-runner` tasks. This is particularly useful for workflows that involve Python-based tools or scripts.

---

## `python.venv(path)`

Creates a Python virtual environment object. Note that this only creates the object in Lua; the environment itself is not created on the file system until you call `:create()`.

*   **Parameters:**
    *   `path` (string): The file system path where the virtual environment should be created (e.g., `./.venv`).
*   **Returns:**
    *   `venv` (object): A virtual environment object with methods to interact with it.

---

### `venv:create()`

Creates the virtual environment on the file system at the specified path.

*   **Returns:**
    *   `error`: An error object if the creation fails.

---

### `venv:pip(command)`

Executes a `pip` command within the context of the virtual environment.

*   **Parameters:**
    *   `command` (string): The arguments to pass to `pip` (e.g., `install -r requirements.txt`).
*   **Returns:**
    *   `result` (table): A table containing the `stdout`, `stderr`, and `exit_code` of the `pip` command.

---

### `venv:exec(script_path)`

Executes a Python script using the Python interpreter from the virtual environment.

*   **Parameters:**
    *   `script_path` (string): The path to the Python script to execute.
*   **Returns:**
    *   `result` (table): A table containing the `stdout`, `stderr`, and `exit_code` of the script execution.

### Example

This example demonstrates a complete lifecycle: creating a virtual environment, installing dependencies from a `requirements.txt` file, and running a Python script.

```lua
-- examples/python_venv_lifecycle_example.lua

TaskDefinitions = {
  main = {
    description = "A task to demonstrate the Python venv lifecycle.",
    create_workdir_before_run = true, -- Use a temporary workdir
    tasks = {
      {
        name = "run-python-script",
        description = "Creates a venv, installs dependencies, and runs a script.",
        command = function(params)
          local python = require("python")
          local workdir = params.workdir -- Get the temp workdir from the group
          
          -- 1. Write our Python script and dependencies to the workdir
          fs.write(workdir .. "/requirements.txt", "requests==2.28.1")
          fs.write(workdir .. "/main.py", "import requests\nprint(f'Hello from Python! Using requests version: {requests.__version__}')")

          -- 2. Create a venv object
          local venv_path = workdir .. "/.venv"
          log.info("Setting up virtual environment at: " .. venv_path)
          local venv = python.venv(venv_path)

          -- 3. Create the venv on the filesystem
          venv:create()

          -- 4. Install dependencies using pip
          log.info("Installing dependencies from requirements.txt...")
          local pip_result = venv:pip("install -r " .. workdir .. "/requirements.txt")
          if pip_result.exit_code ~= 0 then
            log.error("Pip install failed: " .. pip_result.stderr)
            return false, "Failed to install Python dependencies."
          end

          -- 5. Execute the script
          log.info("Running the Python script...")
          local exec_result = venv:exec(workdir .. "/main.py")
          if exec_result.exit_code ~= 0 then
            log.error("Python script failed: " .. exec_result.stderr)
            return false, "Python script execution failed."
          end

          log.info("Python script executed successfully.")
          print("---\n--- Python Script Output ---")
          print(exec_result.stdout)
          print("----------------------------")

          return true, "Python venv lifecycle complete."
        end
      }
    }
  }
}
```

```