# 交互式 REPL

`sloth-runner repl` 命令将您带入一个交互式的 Read-Eval-Print Loop (REPL) 会话。这是一个强大的工具，用于调试、探索和快速实验 sloth-runner 模块。

## 启动 REPL

要启动会话，只需运行：
```bash
sloth-runner repl
```

您还可以预加载一个工作流文件，以使其 `TaskDefinitions` 和任何辅助函数在会话中可用。这对于调试现有的管道非常有用。

```bash
sloth-runner repl -f /path/to/your/pipeline.lua
```

## 功能

### 实时环境
REPL 提供了一个实时的 Lua 环境，您可以在其中执行任何 Lua 代码。所有内置的 sloth-runner 模块（`aws`、`docker`、`fs`、`log` 等）都已预加载并可供使用。

```
sloth> log.info("来自 REPL 的你好！")
sloth> result = fs.read("README.md")
sloth> print(string.sub(result, 1, 50))
```

### 自动补全
REPL 有一个复杂的自动补全系统。
- 开始输入全局变量或模块的名称（例如 `aws`）并按 `Tab` 查看建议。
- 输入模块名称后跟一个点（例如 `docker.`）并按 `Tab` 查看该模块中所有可用的函数。

### 历史记录
REPL 会保留您的命令历史记录。使用向上和向下箭头键浏览以前的命令。

## 会话示例

以下是使用 REPL 调试 Docker 命令的示例。

```bash
$ sloth-runner repl
Sloth-Runner Interactive REPL
输入 'exit' 或 'quit' 离开。
sloth> result = docker.exec({"ps", "-a"})
sloth> print(result.stdout)
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
sloth> -- 现在让我们尝试构建一个镜像
sloth> build_result = docker.build({tag="my-test", path="./examples/docker"})
sloth> print(build_result.success)
true
sloth> exit
再见！
```
