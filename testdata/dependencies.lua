
-- testdata/dependencies.lua
-- Used to track execution order.
EXECUTION_ORDER = {}

TaskDefinitions = {
    test_dependencies = {
        description = "Tests task dependencies execution order",
        tasks = {
            {
                name = "task_C",
                depends_on = {"task_B"},
                command = function()
                    table.insert(EXECUTION_ORDER, "C")
                    return true, "Task C done"
                end
            },
            {
                name = "task_A",
                command = function()
                    table.insert(EXECUTION_ORDER, "A")
                    return true, "Task A done"
                end
            },
            {
                name = "task_B",
                depends_on = {"task_A"},
                command = function()
                    table.insert(EXECUTION_ORDER, "B")
                    return true, "Task B done"
                end
            }
        }
    }
}
