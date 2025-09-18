
-- testdata/params.lua
TaskDefinitions = {
    test_params = {
        description = "Tests parameter substitution from a values file",
        tasks = {
            {
                name = "task_with_params",
                command = function(params, inputs)
                    if params.message == "Hello from values!" and params.nested.key == "value" then
                        return true, "Params correctly substituted"
                    else
                        return false, "Param substitution failed. Got: " .. data.to_json(params)
                    end
                end
            }
        }
    }
}
