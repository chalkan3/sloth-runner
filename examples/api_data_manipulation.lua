-- Define a task group for API and data manipulation examples
TaskDefinitions = {
    api_data_group = {
        description = "Examples for consuming APIs and manipulating JSON/YAML data",
        tasks = {
            {
                name = "fetch_and_transform_data",
                description = "Fetches data from an API, parses JSON, and converts to YAML",
                command = function(params, input)
                    log.info("Fetching data from JSONPlaceholder API...")
                    local body, status_code, headers, err = net.http_get("https://jsonplaceholder.typicode.com/posts/1")

                    if err then
                        log.error("HTTP GET failed: " .. err)
                        return false, "Failed to fetch data", nil
                    end

                    if status_code ~= 200 then
                        log.error("API returned status code: " .. status_code .. " Body: " .. body)
                        return false, "API returned non-200 status", nil
                    end

                    log.info("Successfully fetched data. Status Code: " .. status_code)
                    log.debug("Raw JSON Body: " .. body)

                    -- Parse JSON
                    local json_data, json_err = data.parse_json(body)
                    if json_err then
                        log.error("Failed to parse JSON: " .. json_err)
                        return false, "Failed to parse JSON", nil
                    end
                    log.info("Parsed JSON data.")
                    -- log.debug("JSON Data (Lua table): " .. data.to_json(json_data)) -- Can't print table directly, convert back to JSON for logging

                    -- Accessing data
                    log.info("Title from API: " .. json_data.title)

                    -- Convert to YAML
                    local yaml_string, yaml_err = data.to_yaml(json_data)
                    if yaml_err then
                        log.error("Failed to convert to YAML: " .. yaml_err)
                        return false, "Failed to convert to YAML", nil
                    end
                    log.info("Converted data to YAML.")
                    log.info("YAML Representation:\n" .. yaml_string)

                    -- Demonstrate YAML to JSON round-trip
                    local parsed_yaml_data, parse_yaml_err = data.parse_yaml(yaml_string)
                    if parse_yaml_err then
                        log.error("Failed to parse YAML string back: " .. parse_yaml_err)
                        return false, "Failed to parse YAML back", nil
                    end
                    log.info("Parsed YAML string back to Lua table.")
                    log.debug("Type of parsed_yaml_data: " .. type(parsed_yaml_data)) -- Added debug log

                    local json_from_yaml, json_from_yaml_err = data.to_json(parsed_yaml_data)
                    if json_from_yaml_err then
                        log.error("Failed to convert parsed YAML to JSON: " .. json_from_yaml_err)
                        return false, "Failed to convert parsed YAML to JSON", nil
                    end
                    log.info("Converted parsed YAML back to JSON.")
                    log.debug("JSON from YAML round-trip:\n" .. json_from_yaml)


                    return true, "API data fetched and transformed successfully", {original_json = json_data, yaml_output = yaml_string}
                end,
                async = false,
                depends_on = {},
            },
        },
    },
}