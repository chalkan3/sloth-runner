TaskDefinitions = {
  ["gcp-host-destroy"] = {
    description = "Destroys the gcp-hosts project infrastructure.",
    workdir = '/tmp/gcp-host-deployment', 
    execution_mode = "shared_session",
    clean_workdir_after_run = function(r) 
      log.info("Checking if workdir should be cleaned. Success: " .. tostring(r.success))
      return r.success 
    end,

    tasks = {
      {
        name = "clone_repository",
        command = function(params, inputs, session)
          log.info("Ensuring repository files are present...")
          local fs = require("fs")
          if fs.exists(session.workdir .. "/.git") then
            log.info("Repository already exists, skipping clone.")
            return true, "Repository already exists."
          end

          log.info("Cloning repository to get Pulumi project files...")
          local git = require("git")
          local result = git.clone("https://github.com/chalkan3/gcp-hosts.git", session.workdir)
          if not result.success then
            log.error("Failed to clone repository: " .. result.stderr)
            return false, "Failed to clone repository: " .. result.stderr
          end
          log.info("Repository cloned successfully.")
          return true, "Repository cloned successfully."
        end
      },
      {
        name = "setup_python_env",
        description = "Creates a virtual environment and installs all dependencies.",
        depends_on = "clone_repository",
        command = function(params, inputs, session)
          log.info("Setting up Python environment for destroy...")
          local py = require("python")
          local venv_path = session.workdir .. "/.venv"
          local venv = py.venv(venv_path)
          
          venv:create()
          
          log.info("Installing dependencies...")
          local pip_result = venv:pip("install -r " .. session.workdir .. "/requirements.txt")
          if not pip_result.success then
            log.error("Failed to install python dependencies: " .. pip_result.stderr)
            return false, "Failed to install python dependencies: " .. pip_result.stderr
          end

          log.info("Python environment is ready.")
          return true, "Python environment created.", { venv_path = venv_path }
        end
      },
      {
        name = "destroy_stack",
        description = "Destroys the Pulumi stack.",
        depends_on = "setup_python_env",
        command = function(params, inputs, session)
          log.info("Destroying Pulumi stack...")
          local pulumi = require("pulumi")
          local stack = pulumi.stack("organization/gcp-host/prod", { 
            workdir = session.workdir, 
            venv_path = inputs.setup_python_env.venv_path,
            login_url = 'gs://pulumi-state-backend-chalkan3' 
          })
          
          local result = stack:destroy({ yes = true })
          if not result.success then
            local error_message = "Failed to destroy pulumi stack: " .. result.stderr
            log.error(error_message)
            return false, error_message
          end
          log.info("Pulumi stack destroyed successfully.")
          return true, "Pulumi stack destroyed."
        end
      }
    }
  }
}
