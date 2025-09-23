# Advanced Features

This document covers some of the more advanced features of `sloth-runner`, designed to enhance your development, debugging, and configuration workflows.

## Interactive Task Runner

For complex workflows, it can be useful to step through tasks one by one, inspect their outputs, and decide whether to proceed, skip, or retry a task. The interactive task runner provides a powerful way to debug and develop your task pipelines.

To use the interactive runner, add the `--interactive` flag to the `sloth-runner run` command:

```bash
sloth-runner run -f examples/basic_pipeline.lua --yes --interactive
```

When enabled, the runner will pause before executing each task and prompt you for an action:

```
? Task: fetch_data (Simulates fetching raw data)
> run
  skip
  abort
  continue
```

**Actions:**

*   **run:** (Default) Proceeds with executing the current task.
*   **skip:** Skips the current task and moves to the next one in the execution order.
*   **abort:** Aborts the entire task execution immediately.
*   **continue:** Executes the current task and all subsequent tasks without further prompts, effectively disabling interactive mode for the rest of the run.

## Enhanced `values.yaml` Templating

You can make your `values.yaml` files more dynamic by using Go template syntax to inject environment variables. This is particularly useful for providing sensitive information (like tokens or keys) or environment-specific configurations without hardcoding them.

`sloth-runner` processes `values.yaml` as a Go template, making any environment variables available under the `.Env` map.

**Example:**

1.  **Create a `values.yaml` file with a template placeholder:**

    ```yaml
    # values.yaml
    api_key: "{{ .Env.MY_API_KEY }}"
    region: "{{ .Env.AWS_REGION | default "us-east-1" }}"
    ```
    *Note: You can use `default` to provide a fallback value if the environment variable is not set.*

2.  **Create a Lua task that uses these values:**

    ```lua
    -- my_task.lua
    TaskDefinitions = {
      my_group = {
        tasks = {
          {
            name = "deploy",
            command = function()
              log.info("Deploying to region: " .. values.region)
              log.info("Using API key (first 5 chars): " .. string.sub(values.api_key, 1, 5) .. "...")
              return true, "Deployment successful."
            end
          }
        }
      }
    }
    ```

3.  **Run the task with the environment variables set:**

    ```bash
    export MY_API_KEY="supersecretkey12345"
    export AWS_REGION="us-west-2"

    sloth-runner run -f my_task.lua -v values.yaml --yes
    ```

**Output:**

The output will show that the values from the environment variables were correctly substituted:

```
INFO Deploying to region: us-west-2
INFO Using API key (first 5 chars): super...
```
