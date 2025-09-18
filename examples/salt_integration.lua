-- Define a task group for SaltStack integration examples
TaskDefinitions = {
    salt_integration_group = {
        description = "Examples for integrating with SaltStack using the 'salt' module",
        tasks = {
            {
                name = "salt_ping_minion",
                description = "Pings a specific Salt minion using the fluent API",
                command = function(params, input)
                    log.info("Pinging Salt minion 'keiteguica'...")
                    local stdout, stderr, err = salt.target("keiteguica"):ping():result()

                    if err then
                        log.error("Salt ping failed: " .. err .. " Stderr: " .. stderr)
                        return false, "Salt ping failed"
                    else
                        log.info("Salt ping successful. Result: " .. tostring(stdout))
                        return true, "Salt minion pinged successfully", {result = stdout}
                    end
                end,
            },
            {
                name = "salt_run_command_on_all_minions",
                description = "Runs a shell command on all Salt minions using the fluent API",
                command = function(params, input)
                    log.info("Running 'ls -l /tmp' on all Salt minions...")
                    local stdout, stderr, err = salt.target("*"):cmd("cmd.run", "ls -l /tmp"):result()

                    if err then
                        log.error("Salt cmd.run failed: " .. err .. " Stderr: " .. stderr)
                        return false, "Salt cmd.run failed"
                    else
                        log.info("Salt cmd.run successful. Result: " .. tostring(stdout))
                        return true, "Salt cmd.run executed successfully", {result = stdout}
                    end
                end,
            },
        },
    },
}