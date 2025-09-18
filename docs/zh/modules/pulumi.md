# Pulumi 模块

Sloth-Runner 中的 `pulumi` 模块允许您直接从 Lua 脚本编排 Pulumi 堆栈。这非常适合基础设施即代码 (IaC) 工作流，您需要作为大型自动化管道的一部分来预置、更新或销毁云资源。

## 常见用例

*   **动态预置：** 按需创建暂存或测试环境。
*   **基础设施更新：** 自动化基础设施新版本的部署。
*   **环境管理：** 使用后销毁环境以节省成本。
*   **CI/CD 集成：** 作为 CI/CD 管道的一部分执行 `pulumi up` 或 `preview`。

## API 参考

### `pulumi.stack(name, options_table)`

创建 Pulumi 堆栈的新实例，允许您与其交互。

*   `name` (字符串)：Pulumi 堆栈的完整名称（例如，“my-org/my-project/dev”）。
*   `options_table` (Lua 表)：用于配置堆栈的选项表：
    *   `workdir` (字符串)：**必需。** 与此堆栈关联的 Pulumi 项目根目录的路径。

**返回：**
*   `PulumiStack` (用户数据)：指定堆栈的 `PulumiStack` 对象的实例。

### `PulumiStack` 对象方法

以下所有方法都在 `PulumiStack` 实例上调用（例如，`my_stack:up(...)`）。

#### `stack:up(options)`

执行 `pulumi up` 命令以创建或更新堆栈的资源。

*   `options` (Lua 表，可选)：`up` 命令的选项表：
    *   `non_interactive` (布尔值)：如果为 `true`，则将 `--non-interactive` 和 `--yes` 标志添加到 `pulumi up` 命令。
    *   `config` (Lua 表)：用于将配置传递给堆栈的键值对表（例如，`["my-app:vpcId"] = vpc_id`）。
    *   `args` (字符串的 Lua 表)：要直接传递给 `pulumi up` 命令的其他参数列表。

**返回：**
*   `result` (Lua 表)：包含以下内容的表：
    *   `success` (布尔值)：如果操作成功，则为 `true`；否则为 `false`。
    *   `stdout` (字符串)：Pulumi 命令的标准输出。
    *   `stderr` (字符串)：Pulumi 命令的标准错误输出。
    *   `error` (字符串或 `nil`)：如果命令执行失败，则为 Go 错误消息。

#### `stack:preview(options)`

执行 `pulumi preview` 命令以显示将要应用的更改的预览。

*   `options` (Lua 表，可选)：与 `stack:up()` 相同的选项。

**返回：**
*   `result` (Lua 表)：与 `stack:up()` 相同的返回格式。

#### `stack:refresh(options)`

执行 `pulumi refresh` 命令以使用云中的实际资源更新堆栈的状态。

*   `options` (Lua 表，可选)：与 `stack:up()` 相同的选项。

**返回：**
*   `result` (Lua 表)：与 `stack:up()` 相同的返回格式。

#### `stack:destroy(options)`

执行 `pulumi destroy` 命令以销毁堆栈中的所有资源。

*   `options` (Lua 表，可选)：与 `stack:up()` 相同的选项。

**返回：**
*   `result` (Lua 表)：与 `stack:up()` 相同的返回格式。

#### `stack:outputs()`

获取 Pulumi 堆栈的输出。

**返回：**
*   `outputs` (Lua 表)：一个 Lua 表，其中键是输出名称，值是相应的堆栈输出。
*   `error` (字符串或 `nil`)：如果操作失败，则为错误消息。

## 使用示例

### 基本 Pulumi 编排示例

此示例演示如何部署两个 Pulumi 堆栈，将第一个堆栈的输出作为输入传递给第二个堆栈。

```lua
-- examples/pulumi_example.lua

command = function()
    log.info("正在开始 Pulumi 编排示例...")

    -- 示例 1: 部署一个基础堆栈 (例如 VPC)
    log.info("正在部署基础架构堆栈 (VPC)...")
    local vpc_stack = pulumi.stack("my-org/vpc-network/prod", {
        workdir = "./pulumi/vpc" -- 假设 Pulumi 项目目录在这里
    })

    -- 以非交互方式执行 'pulumi up'
    local vpc_result = vpc_stack:up({ non_interactive = true })

    -- 检查 VPC 部署结果
    if not vpc_result.success then
        log.error("VPC 堆栈部署失败: " .. vpc_result.stderr)
        return false, "VPC 部署失败。"
    end
    log.info("VPC 堆栈部署成功。Stdout: " .. vpc_result.stdout)

    -- 从 VPC 堆栈获取输出
    local vpc_outputs, outputs_err = vpc_stack:outputs()
    if outputs_err then
        log.error("获取 VPC 堆栈输出失败: " .. outputs_err)
        return false, "获取 VPC 输出失败。"
    end

    local vpc_id = vpc_outputs.vpcId -- 假设堆栈导出 'vpcId'
    if not vpc_id then
        log.warn("VPC 堆栈未导出 'vpcId'。将在没有它的情况下继续。")
        vpc_id = "unknown-vpc-id"
    end
    log.info("从输出中获取的 VPC ID: " .. vpc_id)

    -- 示例 2: 部署一个应用程序堆栈，使用前一个堆栈的输出作为配置
    log.info("正在将应用程序堆栈部署到 VPC: " .. vpc_id)
    local app_stack = pulumi.stack("my-org/app-server/prod", {
        workdir = "./pulumi/app" -- 假设应用程序的 Pulumi 项目目录在这里
    })

    -- 执行 'pulumi up'，将前一个堆栈的输出作为配置传递
    local app_result = app_stack:up({
        non_interactive = true,
        config = {
            ["my-app:vpcId"] = vpc_id,
            ["aws:region"] = "us-east-1"
        }
    })

    -- 检查应用程序部署结果
    if not app_result.success then
        log.error("应用程序堆栈部署失败: " .. app_result.stderr)
        return false, "应用程序部署失败。"
    end
    log.info("应用程序堆栈部署成功。Stdout: " .. app_result.stdout)

    log.info("Pulumi 编排示例成功完成。")
    return true, "Pulumi 编排示例已完成。"
end

TaskDefinitions = {
    pulumi_orchestration_example = {
        description = "演示使用 'pulumi' 模块来编排基础设施堆栈。",
        tasks = {
            {
                name = "run_pulumi_orchestration",
                command = command
            }
        }
    }
}
```

---
[English](../../en/modules/pulumi.md) | [Português](../../pt/modules/pulumi.md) | [中文](./pulumi.md)