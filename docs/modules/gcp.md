# GCP Module

The `gcp` module provides a simple interface for executing Google Cloud CLI (`gcloud`) commands from within a `sloth-runner` task.

## `gcp.exec(args)`

Executes a `gcloud` command with the specified arguments.

### Parameters

*   `args` (table): A Lua table (array) of strings representing the arguments to pass to the `gcloud` command. For example, `{"compute", "instances", "list"}`.

### Returns

A table containing the result of the command execution with the following keys:

*   `stdout` (string): The standard output from the command.
*   `stderr` (string): The standard error output from the command.
*   `exit_code` (number): The exit code of the command. An exit code of `0` typically indicates success.

### Example

This example defines a task that lists all Compute Engine instances in the `us-central1` region for a specific project.

```lua
-- examples/gcp_cli_example.lua

TaskDefinitions = {
  main = {
    description = "A task to list GCP compute instances.",
    tasks = {
      {
        name = "list-instances",
        description = "Lists GCE instances in us-central1.",
        command = function()
          log.info("Listing GCP instances...")
          
          -- require the gcp module to make it available
          local gcp = require("gcp")

          -- Execute the gcloud command
          local result = gcp.exec({
            "compute", 
            "instances", 
            "list", 
            "--project", "my-gcp-project-id",
            "--zones", "us-central1-a,us-central1-b"
          })

          -- Check the result
          if result and result.exit_code == 0 then
            log.info("Successfully listed instances.")
            print("--- INSTANCE LIST ---")
            print(result.stdout)
            print("---------------------")
            return true, "GCP command successful."
          else
            log.error("Failed to list GCP instances.")
            if result then
              log.error("Stderr: " .. result.stderr)
            end
            return false, "GCP command failed."
          end
        end
      }
    }
  }
}
```
