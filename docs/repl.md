# Interactive REPL

The `sloth-runner repl` command drops you into an interactive Read-Eval-Print Loop (REPL) session. This is a powerful tool for debugging, exploration, and quick experimentation with the sloth-runner modules.

## Starting the REPL

To start a session, simply run:
```bash
sloth-runner repl
```

You can also pre-load a workflow file to have its `TaskDefinitions` and any helper functions available in the session. This is incredibly useful for debugging an existing pipeline.

```bash
sloth-runner repl -f /path/to/your/pipeline.lua
```

## Features

### Live Environment
The REPL provides a live Lua environment where you can execute any Lua code. All the built-in sloth-runner modules (`aws`, `docker`, `fs`, `log`, etc.) are pre-loaded and ready to use.

```
sloth> log.info("Hello from the REPL!")
sloth> result = fs.read("README.md")
sloth> print(string.sub(result, 1, 50))
```

### Autocompletion
The REPL has a sophisticated autocompletion system.
- Start typing the name of a global variable or module (e.g., `aws`) and press `Tab` to see suggestions.
- Type a module name followed by a dot (e.g., `docker.`) and press `Tab` to see all the functions available in that module.

### History
The REPL keeps a history of your commands. Use the up and down arrow keys to navigate through previous commands.

## Example Session

Here is an example of using the REPL to debug a Docker command.

```bash
$ sloth-runner repl
Sloth-Runner Interactive REPL
Type 'exit' or 'quit' to leave.
sloth> result = docker.exec({"ps", "-a"})
sloth> print(result.stdout)
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
sloth> -- Now let's try to build an image
sloth> build_result = docker.build({tag="my-test", path="./examples/docker"})
sloth> print(build_result.success)
true
sloth> exit
Bye!
```
