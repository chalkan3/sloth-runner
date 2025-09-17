-- examples/conditional_execution.lua

TaskDefinitions = {
    conditional_workflow = {
        description = "A workflow to demonstrate conditional execution with run_if and abort_if.",
        tasks = {
            {
                name = "check_condition_for_run",
                description = "This task creates a file that the next task checks for.",
                command = "touch /tmp/sloth_runner_run_condition"
            },
            {
                name = "conditional_task",
                description = "This task only runs if the condition file exists.",
                depends_on = "check_condition_for_run",
                run_if = "test -f /tmp/sloth_runner_run_condition",
                command = "echo 'Conditional task is running because the condition was met.'"
            },
            {
                name = "cleanup_run_condition",
                description = "Cleans up the run condition file.",
                depends_on = "conditional_task",
                command = "rm /tmp/sloth_runner_run_condition"
            },
            {
                name = "check_abort_condition",
                description = "This task will abort if a specific file exists.",
                abort_if = "test -f /tmp/sloth_runner_abort_condition",
                command = "echo 'This will not run if the abort condition is met.'"
            },
            {
                name = "final_task",
                description = "This task will not be reached if the abort condition is met.",
                depends_on = "check_abort_condition",
                command = "echo 'This is the final task.'"
            }
        }
    }
}
