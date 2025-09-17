-- Define a task group for values.yaml integration examples
TaskDefinitions = {
    values_test_group = {
        description = "Examples for accessing values from values.yaml",
        tasks = {
            {
                name = "display_values",
                description = "Displays values loaded from values.yaml",
                command = function(params, input)
                    log.info("Accessing values from values.yaml:")

                    if values then
                        log.info("App Name: " .. values.app.name)
                        log.info("App Version: " .. values.app.version)
                        log.info("DB Host: " .. values.database.host)
                        log.info("Feature A enabled: " .. tostring(values.features.featureA))
                        log.info("First item in list: " .. values.list_of_items[1])

                        -- Iterate over list_of_items
                        log.info("List of items:")
                        for i, item in ipairs(values.list_of_items) do
                            log.info("  - " .. item)
                        end

                        return true, "Values displayed successfully", nil
                    else
                        log.error("Values table not found. Did you pass --values flag?")
                        return false, "Values not loaded", nil
                    end
                end,
                async = false,
                depends_on = {},
            },
        },
    },
}
