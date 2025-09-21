# GCP 模块

`gcp` 模块提供了一个简单的界面，用于从 `sloth-runner` 任务内部执行谷歌云命令行界面 (`gcloud`) 命令。

## `gcp.exec(args)`

使用指定的参数执行 `gcloud` 命令。

### 参数

*   `args` (table): 一个 Lua 表（数组），包含要传递给 `gcloud` 命令的字符串参数。例如，`{"compute", "instances", "list"}`。

### 返回

一个包含命令执行结果的表，其中包含以下键：

*   `stdout` (string): 命令的标准输出。
*   `stderr` (string): 命令的标准错误输出。
*   `exit_code` (number): 命令的退出代码。退出代码 `0` 通常表示成功。

### 示例

此示例定义了一个任务，用于列出特定项目在 `us-central1` 区域中的所有 Compute Engine 实例。

```lua
-- examples/gcp_cli_example.lua

TaskDefinitions = {
  main = {
    description = "一个列出 GCP 计算实例的任务。",
    tasks = {
      {
        name = "list-instances",
        description = "列出 us-central1 中的 GCE 实例。",
        command = function()
          log.info("正在列出 GCP 实例...")
          
          -- 需要 gcp 模块使其可用
          local gcp = require("gcp")

          -- 执行 gcloud 命令
          local result = gcp.exec({
            "compute", 
            "instances", 
            "list", 
            "--project", "my-gcp-project-id",
            "--zones", "us-central1-a,us-central1-b"
          })

          -- 检查结果
          if result and result.exit_code == 0 then
            log.info("成功列出实例。")
            print("--- 实例列表 ---")
            print(result.stdout)
            print("---------------------")
            return true, "GCP 命令成功。"
          else
            log.error("未能列出 GCP 实例。")
            if result then
              log.error("Stderr: " .. result.stderr)
            end
            return false, "GCP 命令失败。"
          end
        end
      }
    }
  }
}
```
