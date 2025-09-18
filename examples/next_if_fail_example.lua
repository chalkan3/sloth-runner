-- examples/next_if_fail_example.lua

TaskDefinitions = {
    next_if_fail_demo = {
        description = "A workflow to demonstrate the next_if_fail functionality.",
        tasks = {
            {
                name = "task_that_fails",
                description = "This task is designed to fail.",
                command = function()
                    log.error("This task is intentionally failing.")
                    return false, "Intentional failure"
                end
            },
            {
                name = "task_after_failure",
                description = "This task runs only if task_that_fails fails.",
                next_if_fail = "task_that_fails",
                command = "echo 'This task ran because the previous one failed.'"
            },
            {
                name = "task_that_should_be_skipped",
                description = "This task depends on the failing task and should be skipped.",
                depends_on = "task_that_fails",
                command = "echo 'This should not be printed.'"
            },
            {
                name = "final_task",
                description = "This task depends on the task that runs after failure.",
                depends_on = "task_after_failure",
                command = "echo 'This is the final task.'"
            }
        }
    }
}
