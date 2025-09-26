# 分布式任务执行

`sloth-runner` 支持分布式任务执行，允许您在远程代理上运行任务。这使得可扩展的分布式工作流成为可能，其中管道的不同部分可以在不同的机器上执行。

## 工作原理

`sloth-runner` 中的分布式执行模型遵循主从架构：

1.  **主节点：** 主要的 `sloth-runner` 实例充当主节点。它解析工作流定义，识别配置为在远程代理上运行的任务，并分派它们。
2.  **代理：** 在远程机器上以 `agent` 模式运行的 `sloth-runner` 实例。它侦听来自主节点的传入任务执行请求，执行任务，并将结果发回。

## 配置远程任务

要在远程代理上运行任务，您需要在任务组中定义代理，然后为任务指定代理。

### 1. 在任务组中定义代理

在您的 Lua 任务定义文件中，您可以在 `TaskDefinitions` 组中定义一个代理表。每个代理都需要一个唯一的名称和一个 `address`（例如，`host:port`），代理在该地址上侦听。

```lua
TaskDefinitions = {
  my_distributed_group = {
    description = "一个包含分布式任务的任务组。",
    agents = {
      my_remote_agent = { address = "localhost:50051" },
      another_agent = { address = "192.168.1.100:50051" }
    },
    tasks = {
      -- ... 在此处定义任务 ...
    }
  }
}
```

### 2. 将任务分配给代理

在任务组中定义代理后，您可以使用任务定义中的 `agent` 字段将任务分配给特定的代理：

```lua
TaskDefinitions = {
  my_distributed_group = {
    -- ... 代理定义 ...
    tasks = {
      {
        name = "remote_hello",
        description = "在远程代理上运行 hello world 任务。",
        agent = "my_remote_agent", -- 在此处指定代理名称
        command = function(params)
          log.info("来自远程代理的问候！")
          return true, "远程任务已执行。"
        end
      },
      {
        name = "local_task",
        description = "此任务在本地运行。",
        command = "echo '来自本地机器的问候！'"
      }
    }
  }
}
```

## 运行代理

要以代理模式启动 `sloth-runner` 实例，请使用 `agent` 命令：

```bash
sloth-runner agent -p 50051
```

*   `-p, --port`：指定代理应侦听的端口。默认为 `50051`。

当代理启动时，它将侦听来自主 `sloth-runner` 实例的传入 gRPC 请求。收到任务后，它将在其本地环境中执行任务，并将结果以及任何更新的工作区文件返回给主节点。

## 工作区同步

当任务分派到远程代理时，`sloth-runner` 会自动处理任务工作区的同步：

1.  **主节点到代理：** 主节点创建当前任务工作目录的 tarball，并将其发送到代理。
2.  **代理执行：** 代理将 tarball 解压缩到临时目录中，在该目录中执行任务，并捕获对临时目录中文件所做的任何更改。
3.  **代理到主节点：** 任务完成后，代理创建修改后的临时目录的 tarball，并将其发回给主节点。然后，主节点解压缩此 tarball，用远程任务所做的任何更改更新其本地工作区。