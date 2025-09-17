# 🦥 Getting Started Tutorial

Welcome to Sloth Runner! This guide will walk you through creating and running your first set of tasks.

## Prerequisites

Before you begin, make sure you have:
1.  Go (version 1.21+) installed on your system.
2.  The `sloth-runner` executable installed. If not, follow the installation instructions in the main [README.md](../README.md).

## Step 1: Create Your First Task File

Let's create a simple Lua file named `my_tasks.lua`. This file will define our tasks.

```lua
-- my_tasks.lua

TaskDefinitions = {
    hello_world_group = {
        description = "A simple group to say hello",
        tasks = {
            {
                name = "say_hello",
                description = "Prints a friendly greeting",
                command = "echo 'Hello from Sloth Runner! 🦥'"
            }
        }
    }
}
```

This defines a group named `hello_world_group` which contains a single task, `say_hello`. This task simply executes a shell command to print a message.

## Step 2: Run Your Task

Now, let's run the task using the `sloth-runner` CLI. Open your terminal in the same directory where you saved `my_tasks.lua` and run:

```bash
sloth-runner run -f my_tasks.lua
```

You should see the spinner animation and then a success message, indicating your task ran correctly!

## Step 3: Add a Dependent Task

Let's make it more interesting by adding a second task that depends on the first one. Modify `my_tasks.lua`:

```lua
-- my_tasks.lua

TaskDefinitions = {
    hello_world_group = {
        description = "A simple group to say hello",
        tasks = {
            {
                name = "say_hello",
                description = "Prints a friendly greeting",
                command = function()
                    -- We return a table as output
                    return true, "echo 'Hello from Sloth Runner! 🦥'", { message = "Hello World" }
                end
            },
            {
                name = "show_message",
                description = "Shows the message from the first task",
                depends_on = "say_hello", -- This creates the dependency
                command = function(params, inputs)
                    -- The output from 'say_hello' is available here!
                    local received_message = inputs.say_hello.message
                    local command_string = "echo 'The first task said: " .. received_message .. "'"
                    return true, command_string, { confirmation = "Message received!" }
                end
            }
        }
    }
}
```

**Changes:**
-   The `say_hello` task now has a `command` function that returns an output table: `{ message = "Hello World" }`.
-   The new `show_message` task `depends_on` `say_hello`.
-   The `command` function for `show_message` receives the output from its dependency in the `inputs` argument and uses it to build its own command.

## Step 4: Run the Dependent Task

Now, let's run only the final task, `show_message`. Sloth Runner will automatically figure out that it needs to run `say_hello` first.

```bash
sloth-runner run -f my_tasks.lua -t show_message
```

You will see both tasks execute in the correct order.

## What's Next?

Congratulations! You've successfully created and run a task pipeline.

-   Explore the other files in the `/examples` directory to see more complex workflows.
-   Check out the detailed **[Lua API Reference](LUA_API.md)** to see all the powerful modules (`fs`, `net`, `data`, etc.) you can use in your tasks.
