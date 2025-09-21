# Exec Module

The `exec` module is one of the most fundamental modules in `sloth-runner`. It provides a powerful function to execute arbitrary shell commands, giving you full control over the execution environment.

## `exec.run(command, [options])`

Executes a shell command using `bash -c`.

### Parameters

*   `command` (string): The shell command to execute.
*   `options` (table, optional): A table of options to control the execution.
    *   `workdir` (string): The working directory where the command should be executed. If not provided, it runs in the task group's temporary directory (if available) or the current directory.
    *   `env` (table): A dictionary of environment variables (key-value pairs) to set for the command's execution. These are added to the existing environment.

### Returns

A table containing the result of the command execution:

*   `success` (boolean): `true` if the command exited with a code of `0`, otherwise `false`.
*   `stdout` (string): The standard output from the command.
*   `stderr` (string): The standard error output from the command.

### Example

This example demonstrates how to use `exec.run` with a custom working directory and environment variables.

```lua
-- examples/exec_module_example.lua

TaskDefinitions = {
  main = {
    description = "A task to demonstrate the exec module.",
    tasks = {
      {
        name = "run-with-options",
        description = "Executes a command with a custom workdir and environment.",
        command = function()
          log.info("Preparing to run a custom command...")
          
          local exec = require("exec")
          
          -- Create a temporary directory for the example
          local temp_dir = "/tmp/sloth-exec-test"
          fs.mkdir(temp_dir)
          fs.write(temp_dir .. "/test.txt", "hello from test file")

          -- Define options
          local options = {
            workdir = temp_dir,
            env = {
              MY_VAR = "SlothRunner",
              ANOTHER_VAR = "is_awesome"
            }
          }

          -- Execute the command
          local result = exec.run("echo 'MY_VAR is $MY_VAR' && ls -l && cat test.txt", options)

          -- Clean up the temporary directory
          fs.rm_r(temp_dir)

          if result.success then
            log.info("Command executed successfully!")
            print("--- STDOUT ---")
            print(result.stdout)
            print("--------------")
            return true, "Exec command successful."
          else
            log.error("Exec command failed.")
            log.error("Stderr: " .. result.stderr)
            return false, "Exec command failed."
          end
        end
      }
    }
  }
}
```
