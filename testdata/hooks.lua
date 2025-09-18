
-- testdata/hooks.lua
-- Used to track execution order.
EXECUTION_ORDER = {}

TaskDefinitions = {
    test_hooks = {
        description = "Tests pre_exec and post_exec hooks",
        tasks = {
            {
                name = "task_with_hooks",
                pre_exec = function()
                    table.insert(EXECUTION_ORDER, "PRE_EXEC")
                    return true, "Pre-hook OK"
                end,
                command = function()
                    table.insert(EXECUTION_ORDER, "COMMAND")
                    return true, "Command OK", {output_data = "some_result"}
                end,
                post_exec = function(params, output)
                    table.insert(EXECUTION_ORDER, "POST_EXEC")
                    if output.output_data == "some_result" then
                        return true, "Post-hook OK"
                    else
                        return false, "Post-hook failed to receive output"
                    end
                end
            }
        }
    }
}
