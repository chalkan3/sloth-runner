# 分布式任务执行

`sloth-runner` 支持分布式任务执行，允许您在远程代理上运行任务。这使得可扩展的分布式工作流成为可能，其中管道的不同部分可以在不同的机器上执行。

## 工作原理

`sloth-runner` 中的分布式执行模型遵循主从架构：

1.  **主节点：** 主要的 `sloth-runner` 实例充当主节点。它解析工作流定义，识别配置为在远程代理上运行的任务，并分派它们。
2.  **代理：** 在远程机器上以 `agent` 模式运行的 `sloth-runner` 实例。它侦听来自主节点的传入任务执行请求，执行任务，并将结果发回。

## 配置远程任务

要在远程代理上运行任务，您需要在任务组或单个任务定义中指定 `delegate_to` 字段。

### 1. 在任务组级别委托给代理

您可以使用 `delegate_to` 字段直接在 `TaskDefinitions` 组中定义代理。此组中的所有任务都将委托给此代理，除非被任务特定的 `delegate_to` 覆盖。

```lua
TaskDefinitions = {
  my_distributed_group = {
    description = "一个包含分布式任务的任务组。",
    delegate_to = { address = "localhost:50051" }, -- 为整个组定义代理
    tasks = {
      {
        name = "remote_hello",
        description = "在远程代理上运行 hello world 任务。",
        -- 此处不需要 'delegate_to' 字段，它继承自组
        command = function(params)
          log.info("来自远程代理的问候！")
          return true, "远程任务已执行。"
        end
      }
    }
  }
}
```

### 2. 在任务级别委托给代理

或者，您可以直接在单个任务上指定 `delegate_to` 字段。这将覆盖任何组级别的委托或允许即席远程执行。

```lua
TaskDefinitions = {
  my_group = {
    description = "一个包含特定远程任务的任务组。",
    tasks = {
      {
        name = "specific_remote_task",
        description = "在特定远程代理上运行此任务。",
        delegate_to = { address = "192.168.1.100:50051" }, -- 仅为此任务定义代理
        command = function(params)
          log.info("来自特定远程代理的问候！")
          return true, "特定远程任务已执行。"
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