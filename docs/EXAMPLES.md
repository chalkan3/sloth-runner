
---

## Example 3: Generating Tasks with Dynamic Data using Templates

This example demonstrates how to use the `sloth-runner new` command with the `--set` flag to generate a task definition file where content is dynamically injected from the command line. This allows for highly reusable templates that can be customized without modification.

**To generate and run this example:**

1.  **Generate the task file:**
    ```bash
    sloth-runner new templated-task --template simple --set custom_message="This is a custom message from the CLI!" -o examples/templated_task.lua
    ```
    This command uses the `simple` template and injects `custom_message` into the generated Lua file.

2.  **Run the generated task:**
    ```bash
    sloth-runner run -f examples/templated_task.lua -g templated-task -t hello_task
    ```
    Observe the output, which should include the custom message you provided.

---

### **Pipeline: `examples/templated_task.lua`**

```lua
-- examples/templated_task.lua
--
-- This file is generated using 'sloth-runner new' with the --set flag.
-- It demonstrates how to inject dynamic data into templates.

TaskDefinitions = {
  ["templated-task"] = {
    description = "A task group generated with dynamic data.",
    tasks = {
      {
        name = "hello_task",
        description = "An example task with a custom message.",
        command = function(params)
          local workdir = params.workdir
          log.info("Running example task for group templated-task in: " .. workdir)
          log.info("Custom message: This is a custom message from the CLI!")
          local stdout, stderr, err = exec.command("echo 'Hello from sloth-runner!'")
          if err then
            log.error("Failed to run example task: " .. stderr)
            return false, "Dummy task failed."
          else
            log.info("Example task completed successfully.")
            print("Command output: " .. stdout)
            return true, "Dummy task executed successfully."
          end
        end
      }
    }
  }
}
```
