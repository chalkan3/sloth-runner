TaskDefinitions = {
    complex_data_workflow = {
        description = "A complex data processing workflow for {{.Env}}",
        tasks = {
            {
                name = "fetch_raw_data",
                description = "Simulates fetching raw data from a source",
                command = function(params)
                    print("Lua: Fetching raw data...")
                    return true, "echo 'Fetched raw data'", { raw_data = "user_transactions_2023", source = "external_api" }
                end,
                async = true,
                post_exec = function(params, output)
                    print("Lua Hook: fetch_raw_data completed. Raw data: " .. (output.raw_data or "N/A"))
                    return true, "fetch_raw_data post_exec successful"
                end,
            },
            {
                name = "validate_data_schema",
                description = "Validates the schema of the raw data",
                depends_on = "fetch_raw_data",
                command = function(params, input_from_dependency)
                    local raw_data = input_from_dependency.fetch_raw_data.raw_data
                    print("Lua: Validating schema for " .. raw_data .. "...")
                    -- Simulate failure in Production environment
                    if "{{.Env}}" == "Production" then
                        return false, "Schema validation failed in Production environment"
                    end
                    return true, "echo 'Schema validated'", { validated_data = raw_data .. "_validated", validation_status = "success" }
                end,
                async = true,
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: validate_data_schema preparing. Input: " .. (input_from_dependency.fetch_raw_data.raw_data or "N/A"))
                    return true, "validate_data_schema pre_exec successful"
                end,
            },
            {
                name = "transform_data_format",
                description = "Transforms data to a standardized format",
                depends_on = "validate_data_schema",
                command = function(params, input_from_dependency)
                    local validated_data = input_from_dependency.validate_data_schema.validated_data
                    print("Lua: Transforming format for " .. validated_data .. "...")
                    return true, "echo 'Format transformed'", { transformed_data = validated_data .. "_transformed", format_type = "avro" }
                end,
                async = true,
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: transform_data_format preparing. Input: " .. (input_from_dependency.validate_data_schema.validated_data or "N/A"))
                    return true, "transform_data_format pre_exec successful"
                end,
            },
            {
                name = "enrich_data",
                description = "Enriches raw data with external information",
                depends_on = "fetch_raw_data",
                command = function(params, input_from_dependency)
                    local raw_data = input_from_dependency.fetch_raw_data.raw_data
                    print("Lua: Enriching data for " .. raw_data .. "...")
                    return true, "echo 'Data enriched'", { enriched_info = "geo_location_added", original_data = raw_data }
                end,
                async = true,
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: enrich_data preparing. Input: " .. (input_from_dependency.fetch_raw_data.raw_data or "N/A"))
                    return true, "enrich_data pre_exec successful"
                end,
            },
            {
                name = "load_to_staging",
                description = "Loads transformed data to staging area",
                depends_on = "transform_data_format",
                command = function(params, input_from_dependency)
                    local transformed_data = input_from_dependency.transform_data_format.transformed_data
                    print("Lua: Loading " .. transformed_data .. " to staging...")
                    return true, "echo 'Loaded to staging'", { staging_id = "STG_" .. transformed_data .. "_" .. os.time() }
                end,
                async = true,
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: load_to_staging preparing. Input: " .. (input_from_dependency.transform_data_format.transformed_data or "N/A"))
                    return true, "load_to_staging pre_exec successful"
                end,
            },
            {
                name = "generate_report",
                description = "Generates final report based on staged and enriched data",
                depends_on = {"load_to_staging", "enrich_data"}, -- Multiple dependencies
                command = function(params, input_from_dependency)
                    local staging_id = input_from_dependency.load_to_staging.staging_id
                    local enriched_info = input_from_dependency.enrich_data.enriched_info
                    print("Lua: Generating report for staging_id: " .. staging_id .. " with enriched info: " .. enriched_info .. "...")
                    return true, "echo 'Report generated'", { report_url = "http://reports.example.com/" .. staging_id .. "_" .. os.time() .. ".pdf" }
                end,
                async = false, -- This task is synchronous
                pre_exec = function(params, input_from_dependency)
                    print("Lua Hook: generate_report preparing. Staging ID: " .. (input_from_dependency.load_to_staging.staging_id or "N/A") .. ", Enriched Info: " .. (input_from_dependency.enrich_data.enriched_info or "N/A"))
                end,
            }
        }
    }
}
