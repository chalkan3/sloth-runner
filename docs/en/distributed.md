# Distributed Task Execution

`sloth-runner` supports distributed task execution, allowing you to run tasks on remote agents. This enables scalable and distributed workflows, where different parts of your pipeline can be executed on different machines.

## How it Works

The distributed execution model in `sloth-runner` follows a master-agent architecture:

1.  **Master:** The main `sloth-runner` instance acts as the master. It parses the workflow definition, identifies tasks configured to run on remote agents, and dispatches them.
2.  **Agent:** A `sloth-runner` instance running in `agent` mode on a remote machine. It listens for incoming task execution requests from the master, executes the tasks, and sends back the results.

## Configuring Remote Tasks

To run a task on a remote agent, you need to define the agent in your task group and then specify the agent for the task.

### 1. Define Agents in Task Group

In your Lua task definition file, you can define a table of agents within your `TaskDefinitions` group. Each agent needs a unique name and an `address` (e.g., `host:port`) where the agent is listening.

```lua
TaskDefinitions = {
  my_distributed_group = {
    description = "A task group with distributed tasks.",
    agents = {
      my_remote_agent = { address = "localhost:50051" },
      another_agent = { address = "192.168.1.100:50051" }
    },
    tasks = {
      -- ... tasks defined here ...
    }
  }
}
```

### 2. Assign Task to an Agent

Once agents are defined in the task group, you can assign a task to a specific agent using the `agent` field in the task definition:

```lua
TaskDefinitions = {
  my_distributed_group = {
    -- ... agent definitions ...
    tasks = {
      {
        name = "remote_hello",
        description = "Runs a hello world task on a remote agent.",
        agent = "my_remote_agent", -- Specify the agent name here
        command = function(params)
          log.info("Hello from remote agent!")
          return true, "Remote task executed."
        end
      },
      {
        name = "local_task",
        description = "This task runs locally.",
        command = "echo 'Hello from local machine!'"
      }
    }
  }
}
```

## Running an Agent

To start a `sloth-runner` instance in agent mode, use the `agent` command:

```bash
sloth-runner agent -p 50051
```

*   `-p, --port`: Specifies the port the agent should listen on. Defaults to `50051`.

When an agent starts, it will listen for incoming gRPC requests from the master `sloth-runner` instance. Upon receiving a task, it will execute it in its local environment and return the result, along with any updated workspace files, back to the master.

## Workspace Synchronization

When a task is dispatched to a remote agent, `sloth-runner` automatically handles the synchronization of the task's workspace:

1.  **Master to Agent:** The master creates a tarball of the current task's working directory and sends it to the agent.
2.  **Agent Execution:** The agent extracts the tarball into a temporary directory, executes the task within that directory, and any changes made to the files in the temporary directory are captured.
3.  **Agent to Master:** After task completion, the agent creates a tarball of the modified temporary directory and sends it back to the master. The master then extracts this tarball, updating its local workspace with any changes made by the remote task.

This ensures that remote tasks have access to all necessary files and that any modifications they make are reflected back in the main workflow.