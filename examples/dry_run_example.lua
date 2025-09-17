-- examples/dry_run_example.lua

TaskDefinitions = {
    dry_run_demo = {
        description = "A workflow to demonstrate the dry-run functionality.",
        tasks = {
            {
                name = "task_one",
                description = "This task would normally do something.",
                command = "echo 'Executing Task One'"
            },
            {
                name = "task_two",
                description = "This task depends on the first one.",
                depends_on = "task_one",
                command = "echo 'Executing Task Two'"
            },
            {
                name = "task_three",
                description = "This task would also do something.",
                command = "echo 'Executing Task Three'"
            }
        }
    }
}
