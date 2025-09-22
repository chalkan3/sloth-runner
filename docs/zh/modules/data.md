# Data 模块

`data` 模块提供了在 Lua 表和常见数据格式（如 JSON 和 YAML）之间解析和序列化数据的功能。

---

## `data.parse_json(json_string)`

将 JSON 字符串解析为 Lua 表。

*   **参数:**
    *   `json_string` (string): 要解析的 JSON 格式字符串。
*   **返回:**
    *   `table`: 生成的 Lua 表。
    *   `error`: 如果解析失败，则返回一个错误对象。

---

## `data.to_json(lua_table)`

将 Lua 表序列化为 JSON 字符串。

*   **参数:**
    *   `lua_table` (table): 要序列化的 Lua 表。
*   **返回:**
    *   `string`: 生成的 JSON 字符串。
    *   `error`: 如果序列化失败，则返回一个错误对象。

---

## `data.parse_yaml(yaml_string)`

将 YAML 字符串解析为 Lua 表。

*   **参数:**
    *   `yaml_string` (string): 要解析的 YAML 格式字符串。
*   **返回:**
    *   `table`: 生成的 Lua 表。
    *   `error`: 如果解析失败，则返回一个错误对象。

---

## `data.to_yaml(lua_table)`

将 Lua 表序列化为 YAML 字符串。

*   **参数:**
    *   `lua_table` (table): 要序列化的 Lua 表。
*   **返回:**
    *   `string`: 生成的 YAML 字符串。
    *   `error`: 如果序列化失败，则返回一个错误对象。

### 示例

```lua
command = function()
  local data = require("data")

  -- JSON 示例
  log.info("测试 JSON 序列化...")
  local my_table = { name = "sloth-runner", version = 1.0, features = { "tasks", "lua" } }
  local json_str, err = data.to_json(my_table)
  if err then
    return false, "序列化到 JSON 失败: " .. err
  end
  print("序列化的 JSON: " .. json_str)

  log.info("测试 JSON 解析...")
  local parsed_table, err = data.parse_json(json_str)
  if err then
    return false, "解析 JSON 失败: " .. err
  end
  log.info("从 JSON 解析的名称: " .. parsed_table.name)

  -- YAML 示例
  log.info("测试 YAML 序列化...")
  local yaml_str, err = data.to_yaml(my_table)
  if err then
    return false, "序列化到 YAML 失败: " .. err
  end
  print("序列化的 YAML:\n" .. yaml_str)
  
  log.info("测试 YAML 解析...")
  parsed_table, err = data.parse_yaml(yaml_str)
  if err then
    return false, "解析 YAML 失败: " .. err
  end
  log.info("从 YAML 解析的版本: " .. parsed_table.version)

  return true, "Data 模块操作成功。"
end
```

```