TaskDefinitions = {
    data_operations_test = {
        description = "Tests various data serialization/deserialization operations using the 'data' module",
        tasks = {
            {
                name = "test_json_parse_and_to_json",
                description = "Parses JSON string and converts Lua table back to JSON",
                command = function(params)
                    print("Lua: Testing JSON parse and to_json...")
                    local json_str = '{"name": "TaskRunner", "version": 1.0, "active": true, "features": ["fs", "net", "data"]}'
                    local parsed_data, err = data.parse_json(json_str)
                    if err then
                        return false, "Failed to parse JSON: " .. err
                    end

                    print("Parsed JSON name: " .. parsed_data.name)
                    print("Parsed JSON version: " .. parsed_data.version)
                    print("Parsed JSON active: " .. tostring(parsed_data.active))
                    print("Parsed JSON features[1]: " .. parsed_data.features[1])

                    if parsed_data.name ~= "TaskRunner" or parsed_data.version ~= 1.0 or parsed_data.active ~= true or parsed_data.features[1] ~= "fs" then
                        return false, "Parsed JSON data mismatch"
                    end

                    local new_json_str, err = data.to_json(parsed_data)
                    if err then
                        return false, "Failed to convert to JSON: " .. err
                    end
                    print("Converted back to JSON:\n" .. new_json_str)
                    -- Basic check, full comparison is complex for Lua
                    if not string.find(new_json_str, "TaskRunner") or not string.find(new_json_str, "1") then
                        return false, "Converted JSON data mismatch"
                    end

                    return true, "JSON operations successful"
                end,
            },
            {
                name = "test_yaml_parse_and_to_yaml",
                description = "Parses YAML string and converts Lua table back to YAML",
                depends_on = "test_json_parse_and_to_json",
                command = function(params)
                    print("Lua: Testing YAML parse and to_yaml...")
                    local yaml_str = [[ 
name: YAMLTest
version: 2.0
enabled: false
items:
  - item1
  - item2
config:
  key: value
]]
                    local parsed_data, err = data.parse_yaml(yaml_str)
                    if err then
                        return false, "Failed to parse YAML: " .. err
                    end

                    print("Parsed YAML name: " .. parsed_data.name)
                    print("Parsed YAML version: " .. parsed_data.version)
                    print("Parsed YAML enabled: " .. tostring(parsed_data.enabled))
                    print("Parsed YAML items[1]: " .. parsed_data.items[1])
                    print("Parsed YAML config.key: " .. parsed_data.config.key)

                    if parsed_data.name ~= "YAMLTest" or parsed_data.version ~= 2.0 or parsed_data.enabled ~= false or parsed_data.items[1] ~= "item1" or parsed_data.config.key ~= "value" then
                        return false, "Parsed YAML data mismatch"
                    end

                    local new_yaml_str, err = data.to_yaml(parsed_data)
                    if err then
                        return false, "Failed to convert to YAML: " .. err
                    end
                    print("Converted back to YAML:\n" .. new_yaml_str)
                    -- Basic check
                    if not string.find(new_yaml_str, "YAMLTest") or not string.find(new_yaml_str, "item1") then
                        return false, "Converted YAML data mismatch"
                    end

                    return true, "YAML operations successful"
                end,
            },
        }
    }
}
