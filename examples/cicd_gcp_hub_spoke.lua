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
    
    tasks = {
      {
        name = "clone_repository",
        description = "Clones the gcp-hub-spoke project repository into a new temp dir.",
        command = function()
          -- Create a unique temporary directory for this execution.
          local temp_dir, err = fs.tmpname()
          if err then
            return false, "Failed to get temporary directory name: " .. err
          end
          fs.mkdir(temp_dir)
          log.info("Created temporary directory for clone: " .. temp_dir)
          
          local repo_url = "https://github.com/chalkan3/gcp-hub-spoke.git"
          local result = require("git").clone(repo_url, temp_dir)
          
          if not result.success then
            local err_msg = "Failed to clone repository"
            if result.stderr then
              err_msg = err_msg .. ": " .. result.stderr
            end
            log.error(err_msg)
            return false, "Git clone failed", { workdir = temp_dir }
          end
          
          log.info("Repository cloned successfully.")
          -- Return the path to the directory as an output.
          return true, "Repository cloned.", { workdir = temp_dir }
        end
      },
      {
        name = "setup_and_deploy",
        description = "Installs dependencies and deploys the Pulumi stack.",
        depends_on = "clone_repository",
        command = function(params, inputs)
          -- The workdir now comes explicitly from the output of the previous task.
          local workdir = inputs.clone_repository.workdir
          if not workdir then
            return false, "Workdir not received from clone_repository task."
          end
          log.info("Running deployment from: " .. workdir)

          -- 1. Set up Python virtual environment
          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          
          log.info("Creating Python virtual environment...")
          local venv_result = venv:create()
          if not venv_result.success then
            log.error("Failed to create venv: " .. venv_result.stderr)
            return false, "Venv creation failed"
          end

          log.info("Installing Python dependencies from requirements.txt...")
          local pip_result = venv:pip("install -r " .. workdir .. "/requirements.txt")
          if not pip_result.success then
            log.error("Failed to install dependencies: " .. pip_result.stderr)
            return false, "Pip install failed"
          end

          -- 2. Set up and run Pulumi
          local pulumi = require("pulumi")
          
          pulumi.login("file://" .. workdir .. "/pulumi_state")

          -- Define the stack, passing the venv object to ensure it runs in the virtual environment.
          local stack = pulumi.stack("dev", { workdir = workdir, venv = venv })
          
          log.info("Running 'pulumi up'...")
          local up_result = stack:up({ yes = true })

          if not up_result.success then
            log.error("'pulumi up' failed. See output below.")
            log.error("STDOUT: " .. up_result.stdout)
            log.error("STDERR: " .. up_result.stderr)
            return false, "Pulumi up failed"
          end

          log.info("Pulumi up completed successfully.")
          
          return true, "Deployment finished successfully.", stack:outputs()
        end
      }
    }
  }
}