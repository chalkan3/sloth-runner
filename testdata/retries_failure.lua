
-- testdata/retries_failure.lua
TaskDefinitions = {
    test_retries_failure = {
        description = "Tests a task that fails after all retries",
        tasks = {
            {
                name = "persistent_failure_task",
                retries = 2, -- Will attempt 1 initial + 2 retries = 3 total
                command = function(params, inputs)
                    log.info("Executing persistent_failure_task")
                    return false, "This task always fails"
                end
            }
        }
    }
}
