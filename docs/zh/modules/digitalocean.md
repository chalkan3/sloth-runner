# DigitalOcean 模块

`digitalocean` 模块提供了使用 `doctl` 命令行工具与您的 DigitalOcean 资源进行交互的界面。

## 配置

此模块需要安装并验证 `doctl` CLI。标准方法是在您的 DigitalOcean 控制面板中生成个人访问令牌，并将其设置为 `DIGITALOCEAN_ACCESS_TOKEN` 环境变量。

```bash
export DIGITALOCEAN_ACCESS_TOKEN="your_do_api_token_here"
```

该模块将自动将此令牌用于所有命令。

## 通用执行器

### `digitalocean.exec(args)`

执行任何 `doctl` 命令。此函数会自动添加 `--output json` 标志，以确保输出是机器可解析的。

**参数:**

- `args` (table): **必需。** 一个字符串表，表示要传递给 `doctl` 的命令和参数（例如 `{"compute", "droplet", "list"}`）。

**返回:**

一个包含以下字段的表：
- `stdout` (string): 命令的标准输出（作为 JSON 字符串）。
- `stderr` (string): 命令的标准错误。
- `exit_code` (number): 命令的退出代码。`0` 通常表示成功。

**示例:**

```lua
local result = digitalocean.exec({"account", "get"})
if result.exit_code == 0 then
  local account_info, err = data.parse_json(result.stdout)
  if account_info then
    log.info("帐户状态: " .. account_info.status)
  end
end
```

## Droplets 辅助函数

### `digitalocean.droplets.list()`

一个高级包装器，用于列出您帐户中的所有 Droplet。

**返回:**

- 成功时返回 `droplets` (table)，该表是您的 Droplet 对象的已解析 JSON 数组。
- 失败时返回 `nil, error_message`。

**示例:**

```lua
local droplets, err = digitalocean.droplets.list()
if droplets then
  for _, droplet in ipairs(droplets) do
    print("找到 Droplet: " .. droplet.name)
  end
end
```

### `digitalocean.droplets.delete(params)`

按 ID 删除特定的 Droplet。

**参数:**

- `params` (table): 一个包含以下字段的表：
    - `id` (string): **必需。** 要删除的 Droplet 的 ID。
    - `force` (boolean): **可选。** 如果为 `true`，则添加 `--force` 标志以绕过确认提示。默认为 `false`。

**返回:**

- 成功时返回 `true`。
- 失败时返回 `false, error_message`。

**示例:**

```lua
local ok, err = digitalocean.droplets.delete({
  id = "123456789",
  force = true
})
if not ok then
  log.error("删除 droplet 失败: " .. err)
end
```
