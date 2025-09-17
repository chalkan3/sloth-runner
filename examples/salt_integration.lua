-- Define a task group for SaltStack integration examples
TaskDefinitions = {
    salt_integration_group = {
        description = "Examples for integrating with SaltStack using the 'salt' module",
        tasks = {
            {
                name = "salt_ping_minion",
                description = "Pings a specific Salt minion",
                command = function(params, input)
                    log.info("Pinging Salt minion 'minion1'...")
                    -- Example: salt 'minion1' test.ping
                    local stdout, stderr, err = salt.cmd("salt", "minion1", "test.ping")

                    if err then
                        log.error("Salt command failed: " .. err .. " Stderr: " .. stderr)
                        return false, "Salt ping failed", nil
                    else
                        log.info("Salt ping successful. Stdout:\n" .. stdout)
                        if stderr ~= "" then
                            log.warn("Salt command produced stderr:\n" .. stderr)
                        end
                        return true, "Salt minion pinged successfully", {stdout = stdout, stderr = stderr}
                    end
                end,
                async = false,
                depends_on = {},
            },
            {
                name = "salt_call_local_state",
                description = "Runs a local state on the minion where task-runner is executed",
                command = function(params, input)
                    log.info("Running local state.highstate using salt-call...")
                    -- Example: salt-call state.highstate
                    local stdout, stderr, err = salt.cmd("salt-call", "state.highstate")

                    if err then
                        log.error("Salt-call command failed: " .. err .. " Stderr: " .. stderr)
                        return false, "Salt-call state.highstate failed", nil
                    else
                        log.info("Salt-call state.highstate successful. Stdout:\n" .. stdout)
                        if stderr ~= "" then
                            log.warn("Salt-call command produced stderr:\n" .. stderr)
                        end
                        return true, "Salt-call state.highstate executed successfully", {stdout = stdout, stderr = stderr}
                    end
                end,
                async = false,
                depends_on = {"salt_ping_minion"},
            },
            {
                name = "salt_run_command_on_all_minions",
                description = "Runs a shell command on all Salt minions",
                command = function(params, input)
                    log.info("Running 'ls -l /tmp' on all Salt minions...")
                    -- Example: salt '*' cmd.run 'ls -l /tmp'
                    local stdout, stderr, err = salt.cmd("salt", "*", "cmd.run", "ls -l /tmp")

                    if err then
                        log.error("Salt command failed: " .. err .. " Stderr: " .. stderr)
                        return false, "Salt cmd.run failed", nil
                    else
                        log.info("Salt cmd.run successful. Stdout:\n" .. stdout)
                        if stderr ~= "" then
                            log.warn("Salt command produced stderr:\n" .. stderr)
                        end
                        return true, "Salt cmd.run executed successfully", {stdout = stdout, stderr = stderr}
                    end
                end,
                async = false,
                depends_on = {},
            },
            {
                name = "test_specific_minions",
                description = "Pings specific Salt minions 'keiteguica' and 'ladyguica'",
                command = function(params, input)
                    log.info("Pinging Salt minion 'keiteguica'...")
                    local stdout_keiteguica, stderr_keiteguica, err_keiteguica = salt.cmd("salt", "keiteguica", "test.ping")

                    if err_keiteguica then
                        log.error("Salt command for keiteguica failed: " .. err_keiteguica .. " Stderr: " .. stderr_keiteguica)
                        return false, "Salt ping for keiteguica failed", nil
                    else
                        log.info("Salt ping for keiteguica successful. Stdout:\n" .. stdout_keiteguica)
                        if stderr_keiteguica ~= "" then
                            log.warn("Salt command for keiteguica produced stderr:\n" .. stderr_keiteguica)
                        end
                    end

                    log.info("Pinging Salt minion 'ladyguica'...")
                    local stdout_ladyguica, stderr_ladyguica, err_ladyguica = salt.cmd("salt", "ladyguica", "test.ping")

                    if err_ladyguica then
                        log.error("Salt command for ladyguica failed: " .. err_ladyguica .. " Stderr: " .. stderr_ladyguica)
                        return false, "Salt ping for ladyguica failed", nil
                    else
                        log.info("Salt ping for ladyguica successful. Stdout:\n" .. stdout_ladyguica)
                        if stderr_ladyguica ~= "" then
                            log.warn("Salt command for ladyguica produced stderr:\n" .. stderr_ladyguica)
                        end
                    end

                    if err_keiteguica or err_ladyguica then
                        return false, "One or more Salt pings failed", nil
                    else
                        return true, "Both Salt minions pinged successfully", {
                            keiteguica_stdout = stdout_keiteguica,
                            keiteguica_stderr = stderr_keiteguica,
                            ladyguica_stdout = stdout_ladyguica,
                            ladyguica_stderr = stderr_ladyguica,
                        }
                    end
                end,
                async = false,
                depends_on = {},
            },
        },
    },
}
