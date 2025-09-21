# Data Module

The `data` module provides functions for parsing and serializing data between Lua tables and common data formats like JSON and YAML.

---\n

## `data.parse_json(json_string)`

Parses a JSON string into a Lua table.

*   **Parameters:**
    *   `json_string` (string): The JSON formatted string to parse.
*   **Returns:**
    *   `table`: The resulting Lua table.
    *   `error`: An error object if parsing fails.

---\n

## `data.to_json(lua_table)`

Serializes a Lua table into a JSON string.

*   **Parameters:**
    *   `lua_table` (table): The Lua table to serialize.
*   **Returns:**
    *   `string`: The resulting JSON string.
    *   `error`: An error object if serialization fails.

---\n

## `data.parse_yaml(yaml_string)`

Parses a YAML string into a Lua table.

*   **Parameters:**
    *   `yaml_string` (string): The YAML formatted string to parse.
*   **Returns:**
    *   `table`: The resulting Lua table.
    *   `error`: An error object if parsing fails.

---\n

## `data.to_yaml(lua_table)`

Serializes a Lua table into a YAML string.

*   **Parameters:**
    *   `lua_table` (table): The Lua table to serialize.
*   **Returns:**
    *   `string`: The resulting YAML string.
    *   `error`: An error object if serialization fails.

### Example

```lua
command = function()
  local data = require("data")

  -- JSON Example
  log.info("Testing JSON serialization...")
  local my_table = { name = "sloth-runner", version = 1.0, features = { "tasks", "lua" } }
  local json_str, err = data.to_json(my_table)
  if err then
    return false, "Failed to serialize to JSON: " .. err
  end
  print("Serialized JSON: " .. json_str)

  log.info("Testing JSON parsing...")
  local parsed_table, err = data.parse_json(json_str)
  if err then
    return false, "Failed to parse JSON: " .. err
  end
  log.info("Parsed name from JSON: " .. parsed_table.name)

  -- YAML Example
  log.info("Testing YAML serialization...")
  local yaml_str, err = data.to_yaml(my_table)
  if err then
    return false, "Failed to serialize to YAML: " .. err
  end
  print("Serialized YAML:\n" .. yaml_str)
  
  log.info("Testing YAML parsing...")
  parsed_table, err = data.parse_yaml(yaml_str)
  if err then
    return false, "Failed to parse YAML: " .. err
  end
  log.info("Parsed version from YAML: " .. parsed_table.version)

  return true, "Data module operations successful."
end
```

