-- Define a task group
TaskDefinitions = {
    my_exec_group = {
        description = "A group for testing exec commands",
        tasks = {
            {
                name = "print_template_vars",
                description = "Prints template variables passed from Go",
                command = function(params, input)
                    local env = "{{.Env}}"
                    local is_prod = {{.IsProduction}}
                    local shards = {}
                    {{- range .Shards }}
                    table.insert(shards, {{.}})
                    {{- end }}

                    log.info("Environment: " .. env)
                    log.warn("Is Production: " .. tostring(is_prod))
                    log.debug("Shards: " .. table.concat(shards, ", "))
                    log.error("This is a test error message from Lua.")

                    return true, "Template variables printed", nil
                end,
                async = false,
                depends_on = {},
            },
            {
                name = "run_echo_command",
                description = "Runs a simple echo command using exec.run",
                command = function(params, input)
                    local result = exec.run("echo 'Hello from exec!'")
                    if not result.success then
                        return false, "Command failed: " .. result.stderr
                    else
                        return true, "Command executed successfully", {stdout = result.stdout, stderr = result.stderr}
                    end
                end,
                async = false,
                depends_on = {"print_template_vars"}, -- Added dependency
            },
            {
                name = "list_files",
                description = "Lists files in the current directory using exec.run",
                command = function(params, input)
                    local result = exec.run("ls -l")
                    if not result.success then
                        return false, "ls command failed: " .. result.stderr
                    else
                        return true, "ls command executed successfully", {stdout = result.stdout, stderr = result.stderr}
                    end
                end,
                async = false,
                depends_on = {"run_echo_command"},
            },
            {
                name = "another_task_in_exec_group",
                description = "Another task in the exec group",
                command = function(params, input)
                    log.info("Running another_task_in_exec_group")
                    return true, "another_task_in_exec_group completed", nil
                end,
                async = false,
                depends_on = {},
            },
        },
    },
    my_new_group = {
        description = "A new group for testing filtering",
        tasks = {
            {
                name = "new_task_1",
                description = "First task in the new group",
                command = function(params, input)
                    log.info("Running new_task_1")
                    return true, "new_task_1 completed", nil
                end,
                async = false,
                depends_on = {},
            },
            {
                name = "new_task_2",
                description = "Second task in the new group, depends on new_task_1",
                command = function(params, input)
                    log.info("Running new_task_2")
                    return true, "new_task_2 completed", nil
                end,
                async = false,
                depends_on = {"new_task_1"},
            },
        },
    },
}