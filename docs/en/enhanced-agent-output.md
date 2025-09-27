# Enhanced `sloth-runner agent run` Output

## Purpose

This feature significantly improves the visual presentation and informational content of the `sloth-runner agent run` command's output. Previously, the output was a plain text dump, making it difficult to quickly ascertain the status and details of remote command executions. The enhancement aims to provide a more elegant, colorful, and robust user experience by leveraging the `pterm` library for terminal output.

The primary goals of this enhancement are:
*   **Clarity:** Clearly distinguish between successful and failed command executions.
*   **Readability:** Present information in a structured and easy-to-digest format.
*   **Expressiveness:** Utilize colors and visual elements to convey status and highlight important details.
*   **Completeness:** Ensure all relevant information (command, stdout, stderr, error messages) is presented comprehensively.

## Usage

The usage of the `sloth-runner agent run` command remains the same. You execute it from your local machine (where the master is running) to instruct a registered agent to execute a shell command.

```bash
go run ./cmd/sloth-runner agent run <agent_name> '<command_to_execute>'
```

*   `<agent_name>`: The name of the agent registered with the master (e.g., `agent1`, `agent2`).
*   `<command_to_execute>`: The shell command you want the agent to execute. Ensure proper quoting to prevent your local shell from interpreting the command before it reaches the agent.

## Output Style

The enhanced output now utilizes `pterm.DefaultBox` to encapsulate the command execution results, providing a clear visual boundary. Different colors and prefixes are used to indicate success or failure, and sections for the command, standard output, and standard error are clearly delineated.

### Successful Command Execution

Upon successful execution of a command on a remote agent, the output will be presented within a green-bordered box, with a `SUCCESS` title. It will clearly state that the command was successful, show the executed command, and display any `Stdout` content.

**Example Command:**
```bash
go run ./cmd/sloth-runner agent run agent1 'echo "Hello from agent1 on $(hostname)"'
```

**Example Output:**
```
┌─  SUCCESS  Command Execution Result on agent1 ──────────┐
|  SUCCESS  Command executed successfully!                |
|  INFO  Command: echo "Hello from agent1 on $(hostname)" |
| # Stdout:                                               |
| Hello from agent1 on ladyguica                          |
|                                                         |
|                                                         |
└─────────────────────────────────────────────────────────┘
```

### Failed Command Execution

In the event of a command failing on a remote agent, the output will be presented within a red-bordered box, with an `ERROR` title. It will clearly indicate that the command failed, show the executed command, and display any `Stdout`, `Stderr`, and the specific `Error` message returned by the agent.

**Example Command (Hypothetical Failure):**
```bash
go run ./cmd/sloth-runner agent run agent1 'non_existent_command'
```

**Example Output (Hypothetical):**
```
┌─  ERROR  Command Execution Result on agent1 ───────────┐
|  ERROR  Command failed on agent1!                     |
|  INFO  Command: non_existent_command                  |
| # Stderr:                                             |
| bash: non_existent_command: command not found         |
| # Error:                                              |
| exit status 127                                       |
|                                                       |
└───────────────────────────────────────────────────────┘
```

This enhanced output ensures that users receive immediate, clear, and visually distinct feedback on the status of their remote agent commands, significantly improving the debugging and monitoring experience.
