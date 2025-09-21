# Salt 模块

`salt` 模块提供了一个流畅的 API 来与 SaltStack 进行交互，允许您从 `sloth-runner` 工作流中运行远程执行命令和管理配置。

---

## `salt.client([options])`

创建一个 Salt 客户端对象。

*   **参数:**
    *   `options` (table, 可选): 一个选项表。
        *   `config_path` (string): Salt master 配置文件的路径。
*   **返回:**
    *   `client` (object): 一个 `SaltClient` 对象。

---

## `SaltClient` 对象

此对象表示 Salt master 的客户端，并提供用于定位 minions 的方法。

### `client:target(target_string, [expr_form])`

指定命令的目标 minion。

*   **参数:**
    *   `target_string` (string): 目标表达式 (例如, `"*"` 表示所有 minions, `"web-server-1"`, 或一个 grain 值)。
    *   `expr_form` (string, 可选): 要使用的定位类型 (例如, `"glob"`, `"grain"`, `"list"`)。默认为 glob。
*   **返回:**
    *   `target` (object): 一个 `SaltTarget` 对象。

---

## `SaltTarget` 对象

此对象表示一个特定的目标，并提供可链接的方法来执行 Salt 函数。

### `target:cmd(function, [arg1, arg2, ...])`

在目标上执行 Salt 执行模块函数。

*   **参数:**
    *   `function` (string): 要运行的函数的名称 (例如, `"test.ping"`, `"state.apply"`, `"cmd.run"`)。
    *   `arg1`, `arg2`, ... (any): 要传递给 Salt 函数的附加参数。
*   **返回:**
    *   `result` (table): 一个包含 `success` (boolean)、`stdout` (string 或 table) 和 `stderr` (string) 的表。如果 Salt 命令返回 JSON，`stdout` 将是一个解析后的 Lua 表。

### 示例

此示例演示了如何定位 minions 以 ping 它们并应用 Salt 状态。

```lua
command = function()
  local salt = require("salt")

  -- 1. 创建一个 Salt 客户端
  local client = salt.client()

  -- 2. 定位所有 minions 并 ping 它们
  log.info("正在 ping 所有 minions...")
  local ping_result = client:target("*"):cmd("test.ping")
  if not ping_result.success then
    return false, "Ping minions 失败: " .. ping_result.stderr
  end
  print("Ping 结果:")
  print(data.to_yaml(ping_result.stdout)) -- stdout 是一个表

  -- 3. 定位一个特定的 web 服务器并应用一个状态
  log.info("正在向 web-server-1 应用 'nginx' 状态...")
  local apply_result = client:target("web-server-1", "glob"):cmd("state.apply", "nginx")
  if not apply_result.success then
    return false, "应用状态失败: " .. apply_result.stderr
  end
  
  log.info("状态成功应用。")
  return true, "Salt 操作完成。"
end
```
