# Pulumi 模块

`pulumi` 模块提供了一个流畅的 API 来编排 Pulumi 堆栈，使您能够直接从 `sloth-runner` 管理您的基础设施即代码 (IaC) 工作流。

---

## `pulumi.stack(name, options)`

创建一个 Pulumi 堆栈对象。

*   **参数:**
    *   `name` (string): 堆栈的全名 (例如, `"my-org/my-project/dev"`)。
    *   `options` (table): 一个选项表。
        *   `workdir` (string): **(必需)** Pulumi 项目目录的路径。
*   **返回:**
    *   `stack` (object): 一个 `PulumiStack` 对象。
    *   `error`: 如果无法初始化堆栈，则返回一个错误对象。

---

## `PulumiStack` 对象

此对象表示一个特定的 Pulumi 堆栈，并提供用于交互的方法。

### `stack:up([options])`

通过运行 `pulumi up` 创建或更新堆栈的资源。

*   **参数:**
    *   `options` (table, 可选):
        *   `yes` (boolean): 如果为 `true`，则传递 `--yes` 以自动批准更新。
        *   `config` (table): 要传递给堆栈的配置值字典。
        *   `args` (table): 要传递给命令的附加字符串参数列表。
*   **返回:**
    *   `result` (table): 一个包含 `success` (boolean)、`stdout` (string) 和 `stderr` (string) 的表。

### `stack:preview([options])`

通过运行 `pulumi preview` 预览更新将进行的更改。

*   **参数:** 与 `stack:up` 相同。
*   **返回:** 与 `stack:up` 相同。

### `stack:refresh([options])`

通过运行 `pulumi refresh` 刷新堆栈的状态。

*   **参数:** 与 `stack:up` 相同。
*   **返回:** 与 `stack:up` 相同。

### `stack:destroy([options])`

通过运行 `pulumi destroy` 销毁堆栈中的所有资源。

*   **参数:** 与 `stack:up` 相同。
*   **返回:** 与 `stack:up` 相同。

### `stack:outputs()`

检索已部署堆栈的输出。

*   **返回:**
    *   `outputs` (table): 堆栈输出的 Lua 表。
    *   `error`: 如果获取输出失败，则返回一个错误对象。

### 示例

此示例显示了一个常见模式：部署一个网络堆栈 (VPC)，然后使用其输出 (`vpcId`) 来配置和部署一个应用程序堆栈。

```lua
command = function()
  local pulumi = require("pulumi")

  -- 1. 定义 VPC 堆栈
  local vpc_stack = pulumi.stack("my-org/vpc/prod", { workdir = "./pulumi/vpc" })
  
  -- 2. 部署 VPC
  log.info("正在部署 VPC 堆栈...")
  local vpc_result = vpc_stack:up({ yes = true })
  if not vpc_result.success then
    return false, "VPC 部署失败: " .. vpc_result.stderr
  end

  -- 3. 从其输出中获取 VPC ID
  log.info("正在获取 VPC 输出...")
  local vpc_outputs, err = vpc_stack:outputs()
  if err then
    return false, "获取 VPC 输出失败: " .. err
  end
  local vpc_id = vpc_outputs.vpcId

  -- 4. 定义应用程序堆栈
  local app_stack = pulumi.stack("my-org/app/prod", { workdir = "./pulumi/app" })

  -- 5. 部署应用程序，将 vpcId 作为配置传递
  log.info("正在将应用程序堆栈部署到 VPC: " .. vpc_id)
  local app_result = app_stack:up({
    yes = true,
    config = { ["my-app:vpcId"] = vpc_id }
  })
  if not app_result.success then
    return false, "应用程序部署失败: " .. app_result.stderr
  end

  log.info("所有堆栈均已成功部署。")
  return true, "Pulumi 编排完成。"
end
```
