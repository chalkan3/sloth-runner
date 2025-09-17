TaskDefinitions = {
    basic_pipeline = {
        description = "A simple data processing pipeline",
        tasks = {
            {
                name = "fetch_data",
                description = "Simulates fetching raw data",
                command = function(params)
                    print("Lua: Executing fetch_data...")
                    -- Simulate success
                    return true, "echo 'Fetched raw data'", { raw_data = "some_data_from_api", source = "external_api" }
                end,
                post_exec = function(params, output)
                    print("Lua Hook: fetch_data completed. Raw data: " .. (output.raw_data or "N/A"))
                    return true, "fetch_data post_exec successful"
                end,
            },
            {
                name = "process_data",
                description = "Processes the raw data",
                depends_on = "fetch_data",
                command = function(params, input_from_dependency)
                    local raw_data = input_from_dependency.fetch_data.raw_data
                    print("Lua: Executing process_data with input: " .. raw_data)
                    -- Simulate a potential failure
                    if raw_data == "invalid_data" then
                        return false, "Invalid data received for processing"
                    end
                    return true, "echo 'Processed data'", { processed_data = "processed_" .. raw_data, status = "success" }
                end,
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: process_data preparing. Input source: " .. (input_from_dependency.fetch_data.source or "unknown"))
                    return true, "process_data pre_exec successful"
                end,
            },
            {
                name = "store_result",
                description = "Stores the final processed data",
                depends_on = "process_data",
                command = function(params, input_from_dependency)
                    local final_data = input_from_dependency.process_data.processed_data
                    print("Lua: Executing store_result with final data: " .. final_data)
                    return true, "echo 'Result stored'", { final_result = final_data, timestamp = os.time() }
                end,
            }
        }
    }
}
