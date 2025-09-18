
-- testdata/timeout.lua
TaskDefinitions = {
    test_timeout = {
        description = "Tests task timeout",
        tasks = {
            {
                name = "long_running_task",
                timeout = "100ms",
                command = "sleep 1"
            }
        }
    }
}
