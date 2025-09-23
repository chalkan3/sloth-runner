-- examples/fluent_git_pulumi_workflow.lua
--
-- This pipeline demonstrates a fluent, object-oriented workflow for deploying
-- infrastructure from a Git repository using Pulumi.

TaskDefinitions = {
  ["fluent-git-pulumi-deploy"] = {
    description = "Clones a repo and deploys it using a fluent workflow.",
    workdir = "/tmp/fluent-deployment", -- Using a fixed workdir for predictability
    tasks = {
      {
        name = "clone_and_deploy",
        description = "Clones, configures, and deploys the infrastructure stack.",
        command = function(params)
          local git = require("git")
          local pulumi = require("pulumi")

          -- Step 1: Clone the repository.
          -- The 'git.clone' function now returns a rich 'repo' object.
          log.info("Cloning repository: " .. values.infra_stack.repo_url)
          local repo = git.clone(values.infra_stack.repo_url, params.workdir .. "/infra")
          if repo == nil then
            log.error("Failed to clone repository.")
            return false, "Git clone failed."
          end

          log.info("Repository cloned to: " .. repo.local.path)
          log.info("Current branch is: " .. repo.current.branch)

          -- Step 2: Setup Python Environment
          log.info("Setting up Python virtual environment...")
          local python = require("python")
          local venv = python.venv(repo.local.path .. "/.venv")
          venv:create()
          venv:pip("install -r " .. repo.local.path .. "/requirements.txt")

          -- Step 3: Create the Pulumi stack object.
          -- The login is handled implicitly by passing the 'login' parameter.
          log.info("Creating Pulumi stack object for '" .. values.infra_stack.name .. "'")
          local stack = pulumi.stack(values.infra_stack.name, {
            workdir = repo.local.path,
            login = values.gcp.pulumi_login_bucket
          })

          -- Step 4: Configure the stack using values from the YAML file.
          log.info("Applying configuration to the stack...")
          for key, value in pairs(values.infra_stack.config) do
            stack:config(key, value)
          end

          -- Step 5: Deploy the infrastructure.
          -- The 'up' method is called directly on the stack object.
          log.info("Deploying infrastructure with 'pulumi up'...")
          local up_result = stack:up({ yes = true })
          if not up_result.success then
            log.error("Pulumi 'up' failed: " .. up_result.stderr)
            return false, "Pulumi deployment failed."
          end

          log.info("Deployment successful. Fetching outputs...")
          -- Step 6: Get outputs from the stack object.
          local outputs = stack:outputs()

          log.info("--- Pulumi Outputs ---")
          for name, value in pairs(outputs) do
            log.info(name .. ": " .. tostring(value))
          end
          log.info("----------------------")

          return true, "Fluent Git+Pulumi workflow completed successfully."
        end
      }
    }
  }
}
