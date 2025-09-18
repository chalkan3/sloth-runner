
-- testdata/async.lua
TaskDefinitions = {
    test_async = {
        description = "Tests async task execution",
        tasks = {
            {
                name = "task_A",
                async = true,
                command = "sleep 0.2"
            },
            {
                name = "task_B",
                async = true,
                command = "sleep 0.2"
            }
        }
    }
}
