-- examples/retries_and_timeout.lua

TaskDefinitions = {
    robust_workflow = {
        description = "A workflow to demonstrate retries and timeouts",
        tasks = {
            {
                name = "flaky_task",
                description = "This task fails 50% of the time",
                retries = 3,
                command = function()
                    if math.random() < 0.5 then
                        log.error("Simulating a random failure!")
                        return false, "Random failure occurred"
                    end
                    return true, "echo 'Flaky task succeeded!'", { result = "success" }
                end
            },
            {
                name = "long_running_task",
                description = "This task simulates a long process that will time out",
                timeout = "2s",
                command = "sleep 5 && echo 'This should not be printed'"
            },
            {
                name = "final_task",
                description = "This task runs only if the flaky task eventually succeeds",
                depends_on = "flaky_task",
                command = "echo 'The flaky task was successful!'"
            }
        }
    }
}
