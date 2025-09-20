-- examples/cicd_gcp_hub_spoke.lua
--
-- This example demonstrates a full CI/CD pipeline for a Pulumi project.
-- It performs the following steps:
-- 1. Clones a Git repository containing a GCP Hub-Spoke infrastructure definition.
-- 2. Creates a Python virtual environment.
-- 3. Installs the required Python dependencies using pip.
-- 4. Logs into the Pulumi service (using a local file backend for this example).
-- 5. Runs `pulumi up` to deploy the infrastructure.

TaskDefinitions = {
  ["gcp_hub_spoke_deploy"] = {
    description = "CI/CD Pipeline: Clones, sets up the environment, and deploys the gcp-hub-spoke Pulumi project.",
    
    -- Create a temporary, isolated directory for this run.
    create_workdir_before_run = true,
    
    -- Clean up the directory only if the entire pipeline succeeds.
    -- If it fails, the directory is kept for debugging.
    clean_workdir_after_run = function(last_result)
      if not last_result.success then
        log.error("Pipeline failed. The workdir will be kept for debugging at: " .. last_result.output.workdir)
      end
      return last_result.success
    end,

    tasks = {
      {
        name = "clone_repository",
        description = "Clones the gcp-hub-spoke project repository.",
        command = function(params)
          local workdir = params.workdir
          log.info("Cloning repository into: " .. workdir)
          
          local repo_url = "https://github.com/chalkan3/gcp-hub-spoke.git"
          local result = require("git").repo(workdir):clone(repo_url)
          
          if not result.success then
            log.error("Failed to clone repository: " .. result.stderr)
            return false, "Git clone failed", { workdir = workdir }
          end
          
          log.info("Repository cloned successfully.")
          return true, "Repository cloned.", { workdir = workdir }
        end
      },
      {
        name = "setup_and_deploy",
        description = "Installs dependencies and deploys the Pulumi stack.",
        depends_on = "clone_repository",
        command = function(params, inputs)
          -- The workdir is the output from the previous task.
          local workdir = inputs.clone_repository.workdir
          log.info("Running deployment from: " .. workdir)

          -- 1. Set up Python virtual environment
          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          
          log.info("Creating Python virtual environment...")
          local venv_result = venv:create()
          if not venv_result.success then
            log.error("Failed to create venv: " .. venv_result.stderr)
            return false, "Venv creation failed", { workdir = workdir }
          end

          log.info("Installing Python dependencies from requirements.txt...")
          local pip_result = venv:pip("install -r " .. workdir .. "/requirements.txt")
          if not pip_result.success then
            log.error("Failed to install dependencies: " .. pip_result.stderr)
            return false, "Pip install failed", { workdir = workdir }
          end

          -- 2. Set up and run Pulumi
          local pulumi = require("pulumi")
          
          -- For this example, we'll use a local file backend to avoid needing a real login.
          -- The login path is relative to the workdir.
          pulumi.login("file://./pulumi_state")

          -- Define the stack. The workdir is crucial here.
          local stack = pulumi.stack("dev", { workdir = workdir })
          
          log.info("Initializing Pulumi stack...")
          stack:init()

          log.info("Running 'pulumi up'...")
          -- We pass '-y' to auto-approve the update.
          local up_result = stack:up({ yes = true })

          if not up_result.success then
            log.error("'pulumi up' failed. See output below.")
            log.error("STDOUT: " .. up_result.stdout)
            log.error("STDERR: " .. up_result.stderr)
            return false, "Pulumi up failed", { workdir = workdir }
          end

          log.info("Pulumi up completed successfully.")
          
          -- Return the workdir path for the clean_workdir_after_run function
          local final_output = stack:outputs()
          final_output.workdir = workdir
          
          return true, "Deployment finished successfully.", final_output
        end
      }
    }
  }
}
