# Distributed Task Execution

`sloth-runner` supports distributed task execution, allowing you to run tasks on remote agents. This enables scalable and distributed workflows, where different parts of your pipeline can be executed on different machines.

## How it Works

The distributed execution model in `sloth-runner` follows a master-agent architecture:

1.  **Master:** The main `sloth-runner` instance acts as the master. It parses the workflow definition, identifies tasks configured to run on remote agents, and dispatches them.
2.  **Agent:** A `sloth-runner` instance running in `agent` mode on a remote machine. It listens for incoming task execution requests from the master, executes the tasks, and sends back the results.

## Configuring Remote Tasks

To run a task on a remote agent, you need to specify the `delegate_to` field in either the task group or the individual task definition.

### 1. Delegate to an Agent at the Task Group Level

You can define the agent directly within your `TaskDefinitions` group using the `delegate_to` field. All tasks within this group will then be delegated to this agent unless overridden by a task-specific `delegate_to`.

```lua
TaskDefinitions = {
  my_distributed_group = {
    description = "A task group with distributed tasks.",
    delegate_to = { address = "localhost:50051" }, -- Define the agent for the entire group
    tasks = {
      {
        name = "remote_hello",
        description = "Runs a hello world task on a remote agent.",
        -- No 'delegate_to' field needed here, it inherits from the group
        command = function(params)
          log.info("Hello from remote agent!")
          return true, "Remote task executed."
        end
      }
    }
  }
}
```

### 2. Delegate to an Agent at the Task Level

Alternatively, you can specify the `delegate_to` field directly on an individual task. This will override any group-level delegation or allow for ad-hoc remote execution.

```lua
TaskDefinitions = {
  my_group = {
    description = "A task group with a specific remote task.",
    tasks = {
      {
        name = "specific_remote_task",
        description = "Runs this task on a specific remote agent.",
        delegate_to = { address = "192.168.1.100:50051" }, -- Define agent for this task only
        command = function(params)
          log.info("Hello from a specific remote agent!")
          return true, "Specific remote task executed."
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