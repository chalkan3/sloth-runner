# Advanced Examples

This section presents more complex examples and use cases that combine multiple Sloth-Runner modules for end-to-end automation.

## Full Example: End-to-End CI/CD Pipeline

This tutorial demonstrates how to build a complete CI/CD pipeline using the `git`, `pulumi`, and `salt` modules to version code, provision infrastructure, and deploy an application.

### Scenario

Imagine you have a Pulumi infrastructure project and an application project. You want to automate the following flow:

1.  Clone the infrastructure repository.
2.  Update a version file within the repository.
3.  Commit and push this change to Git.
4.  Execute `pulumi up` to provision or update the infrastructure (e.g., a staging environment).
5.  Use Salt to configure the provisioned servers and deploy the application.

### Lua Script (`examples/pulumi_git_combined_example.lua`)

```lua
-- examples/pulumi_git_combined_example.lua

command = function(params)
    log.info("Starting combined Pulumi and Git example...")

    local pulumi_repo_url = "https://github.com/my-org/my-pulumi-infra.git" -- Example Pulumi repo
    local pulumi_repo_path = "./pulumi-infra-checkout"
    local new_infra_version = params.infra_version or "v1.0.0-infra"
    local pulumi_project_workdir = pulumi_repo_path .. "/my-vpc-project" -- Subdirectory within the cloned repo
    local repo

    -- 1. Clone or open the Pulumi repository
    log.info("Step 1: Cloning or opening Pulumi repository...")
    if not fs.exists(pulumi_repo_path) then
        log.info("Cloning Pulumi repository: " .. pulumi_repo_url)
        local cloned_repo, clone_err = git.clone(pulumi_repo_url, pulumi_repo_path)
        if clone_err then
            log.error("Failed to clone Pulumi repository: " .. clone_err)
            return false, "Git clone failed."
        end
        repo = cloned_repo
    else
        log.info("Pulumi repository already exists, opening local reference.")
        local opened_repo, open_err = git.repo(pulumi_repo_path)
        if open_err then
            log.error("Failed to open Pulumi repository: " .. open_err)
            return false, "Git repo open failed."
        end
        repo = opened_repo
    end

    if not repo then
        return false, "Failed to get Pulumi repository reference."
    end

    -- 2. Update the repository (pull)
    log.info("Step 2: Pulling latest changes from Pulumi repository...")
    repo:checkout("main"):pull("origin", "main")
    local pull_result = repo:result()
    if not pull_result.success then
        log.error("Failed to pull Pulumi repository: " .. pull_result.stderr)
        return false, "Git pull failed."
    end
    log.info("Pulumi repository updated. Stdout: " .. pull_result.stdout)

    -- 3. Simulate a change in the Pulumi code (e.g., update a version file)
    log.info("Step 3: Simulating a change in Pulumi code (updating version file)...")
    local infra_version_file = pulumi_repo_path .. "/INFRA_VERSION"
    fs.write(infra_version_file, new_infra_version)
    log.info("Updated INFRA_VERSION file to: " .. new_infra_version)

    -- 4. Commit and push the changes
    log.info("Step 4: Committing and pushing infrastructure version change...")
    local commit_message = "ci: Bump infrastructure version to " .. new_infra_version
    repo:add(infra_version_file)
        :commit(commit_message)
        :push("origin", "main") -- No follow_tags here, just the commit

    local push_result = repo:result()
    if not push_result.success then
        log.error("Failed to push infrastructure changes: " .. push_result.stderr)
        return false, "Git push failed for infra changes."
    end
    log.info("Infrastructure version change pushed. Stdout: " .. push_result.stdout)

    -- 5. Execute 'pulumi up' for the project
    log.info("Step 5: Running pulumi up for the infrastructure project...")
    local infra_stack = pulumi.stack("my-org/my-infra/dev", {
        workdir = pulumi_project_workdir -- Use the subdirectory of the Pulumi project
    })

    local pulumi_up_result = infra_stack:up({ non_interactive = true })

    if not pulumi_up_result.success then
        log.error("Pulumi up failed: " .. pulumi_up_result.stderr)
        return false, "Pulumi up failed."
    end
    log.info("Pulumi up completed successfully. Stdout: " .. pulumi_up_result.stdout)

    -- 6. Configure and deploy application using Salt (Example)
    log.info("Step 6: Configuring and deploying application using Salt...")
    -- Assuming Pulumi up provided the server IP or hostname
    -- For this example, we'll use a fictitious IP
    local server_ip = "192.168.1.100" -- Replace with actual output from Pulumi, if any
    local salt_target = salt.target(server_ip)

    log.info("Running Salt test.ping on " .. server_ip .. "...")
    salt_target:ping()
    local ping_result = salt_target:result()
    if not ping_result.success then
        log.error("Salt ping failed for " .. server_ip .. ": " .. ping_result.stderr)
        return false, "Salt ping failed."
    end
    log.info("Salt ping successful. Stdout: " .. data.to_json(ping_result.stdout)) -- Assuming ping returns JSON

    log.info("Applying Salt state 'app.install' on " .. server_ip .. "...")
    salt_target:cmd('state.apply', 'app.install')
    local salt_apply_result = salt_target:result()
    if not salt_apply_result.success then
        log.error("Salt state.apply failed for " .. server_ip .. ": " .. salt_apply_result.stderr)
        return false, "Salt state.apply failed."
    end
    log.info("Salt state.apply successful. Stdout: " .. data.to_json(salt_apply_result.stdout))

    log.info("Combined Pulumi and Git example finished successfully.")
    return true, "Combined Pulumi and Git example finished."
end

TaskDefinitions = {
    pulumi_git_combined_example = {
        description = "Demonstrates combined usage of 'pulumi' and 'git' modules for CI/CD pipeline.",
        tasks = {
            {
                name = "run_combined_example",
                command = command,
                params = {
                    infra_version = "v1.0.0-test-combined"
                }
            }
        }
    }
}
```

---
[English](./advanced-examples.md) | [Português](../pt/advanced-examples.md) | [中文](../zh/advanced-examples.md)