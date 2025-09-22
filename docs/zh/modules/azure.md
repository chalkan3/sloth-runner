# Azure 模块

`azure` 模块提供了使用 `az` 命令行工具与 Microsoft Azure 进行交互的界面。

## 配置

此模块需要安装并验证 `az` CLI。在使用此模块的管道运行之前，您必须登录到您的 Azure 帐户：

```bash
az login
```

该模块将使用您登录的凭据执行所有命令。

## 通用执行器

### `azure.exec(args)`

执行任何 `az` 命令。此函数会自动添加 `--output json` 标志（如果尚不存在），以确保输出是机器可解析的。

**参数:**

- `args` (table): **必需。** 一个字符串表，表示要传递给 `az` 的命令和参数（例如 `{"group", "list", "--location", "eastus"}`）。

**返回:**

一个包含以下字段的表：
- `stdout` (string): 命令的标准输出（作为 JSON 字符串）。
- `stderr` (string): 命令的标准错误。
- `exit_code` (number): 命令的退出代码。`0` 通常表示成功。

**示例:**

```lua
local result = azure.exec({"account", "show"})
if result.exit_code == 0 then
  local account_info, err = data.parse_json(result.stdout)
  if account_info then
    log.info("登录为: " .. account_info.user.name)
  end
end
```

## 资源组 (RG) 辅助函数

### `azure.rg.delete(params)`

删除资源组。

**参数:**

- `params` (table): 一个包含以下字段的表：
    - `name` (string): **必需。** 要删除的资源组的名称。
    - `yes` (boolean): **可选。** 如果为 `true`，则添加 `--yes` 标志以绕过确认提示。

**返回:**

- 成功时返回 `true`。
- 失败时返回 `false, error_message`。

**示例:**

```lua
local ok, err = azure.rg.delete({
  name = "my-test-rg",
  yes = true
})
if not ok then
  log.error("删除资源组失败: " .. err)
end
```

## 虚拟机 (VM) 辅助函数

### `azure.vm.list(params)`

列出虚拟机。

**参数:**

- `params` (table): **可选。** 一个包含以下字段的表：
    - `resource_group` (string): 用于将列表范围限定为的资源组的名称。如果省略，则列出整个订阅中的 VM。

**返回:**

- 成功时返回 `vms` (table)，该表是您的 VM 对象的已解析 JSON 数组。
- 失败时返回 `nil, error_message`。

**示例:**

```lua
-- 列出订阅中的所有 VM
local all_vms, err1 = azure.vm.list()

-- 列出特定资源组中的 VM
local specific_vms, err2 = azure.vm.list({resource_group = "my-production-rg"})
if specific_vms then
  for _, vm in ipairs(specific_vms) do
    print("找到 VM: " .. vm.name)
  end
end
```
