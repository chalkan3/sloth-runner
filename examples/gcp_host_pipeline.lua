TaskDefinitions = {
  ["gcp-host"] = {
    description = "CI/CD Pipeline: Clones, prepares environment, configures, and deploys the gcp-hosts project.",
    create_workdir_before_run = true,
    workdir = '/tmp/gcp-host-deployment',
    execution_mode = "shared_session", 
    clean_workdir_after_run = function(r) return r.success end,

    tasks = {
      {
        name = "clone_repository",
        command = function(params, inputs, session)
          log.info("Cloning repository...")
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
          log.info("Setting up Python environment...")
          local py = require("python")
          local venv_path = session.workdir .. "/.venv"
          local venv = py.venv(venv_path)
          
          log.info("Creating venv...")
          local create_result = venv:create()
          if not create_result.success then
            log.error("Failed to create python venv: " .. create_result.stderr)
            return false, "Failed to create python venv: " .. create_result.stderr
          end
          
          log.info("Installing dependencies...")
          local pip_result = venv:pip("install -r " .. session.workdir .. "/requirements.txt")
          if not pip_result.success then
            log.error("Failed to install python dependencies: " .. pip_result.stderr)
            return false, "Failed to install python dependencies: " .. pip_result.stderr
          end

          log.info("Python environment created.")
          return true, "Python environment created.", { venv_path = venv_path }
        end
      },
      {
        name = "init_stack",
        description = "Initializes the Pulumi stack and passes it to the next task.",
        depends_on = "setup_python_env",
        command = function(params, inputs, session)
          log.info("Initializing Pulumi stack...")
          local pulumi = require("pulumi")
          local stack_name = "organization/gcp-host/prod"
          local stack = pulumi.stack(stack_name, { 
            workdir = session.workdir, 
            venv_path = inputs.setup_python_env.venv_path,
            login_url = 'gs://pulumi-state-backend-chalkan3' 
          })
          
          local result = stack:select({ create = true })
          if not result.success then
            log.error("Failed to select pulumi stack: " .. result.stderr)
            return false, "Failed to select pulumi stack: " .. result.stderr
          end
          
          log.info("Pulumi stack selected.")
          return true, "Pulumi stack selected.", { 
            stack_name = stack_name,
            workdir = session.workdir,
            venv_path = inputs.setup_python_env.venv_path,
            login_url = 'gs://pulumi-state-backend-chalkan3'
          }
        end
      },
      {
        name = "configure_pulumi",
        description = "Sets all required Pulumi config values using the passed stack object.",
        depends_on = "init_stack",
        command = function(params, inputs)
          log.info("Configuring Pulumi...")
          local pulumi = require("pulumi")
          local stack_info = inputs.init_stack
          local stack = pulumi.stack(stack_info.stack_name, {
            workdir = stack_info.workdir,
            venv_path = stack_info.venv_path,
            login_url = stack_info.login_url
          })
          
          local configs = {
            { "gcp:project", "chalkan3" },
            { "gcp:zone", "us-central1-a" },
            { "gcp-host:machineType", "e2-medium" },
            { "gcp-host:bootDiskImage", "ubuntu-os-cloud/ubuntu-2204-lts" },
            { "gcp-host:instanceName", "vm-gcp-squid-proxy-02" },
            { "gcp-host:saltMasterIp", "34.57.154.158" },
            { "gcp-host:firewallPorts", '["3128"]' },
            { "gcp-host:saltGrains", '{ "roles": ["squid"], "environment": "production" }' }
          }

          for _, conf in ipairs(configs) do
            log.info("Setting config: " .. conf[1])
            local result = stack:config(conf[1], conf[2])
            if not result.success then
              log.error("Failed to set config '" .. conf[1] .. "': " .. result.stderr)
              return false, "Failed to set config '" .. conf[1] .. "': " .. result.stderr
            end
          end

          log.info("Pulumi configured.")
          return true, "Stack Configurated", inputs.init_stack
        end
      },
      {
        name = "deploy_stack",
        description = "Deploys the Pulumi stack using the passed stack object.",
        depends_on = "configure_pulumi",
        command = function(params, inputs)
          log.info("Deploying Pulumi stack...")
          local pulumi = require("pulumi")
          local stack_info = inputs.configure_pulumi
          local stack = pulumi.stack(stack_info.stack_name, {
            workdir = stack_info.workdir,
            venv_path = stack_info.venv_path,
            login_url = stack_info.login_url
          })
          
          local result = stack:up({ yes = true, skip_preview = true })
          if not result.success then
            local error_message = "Failed to deploy pulumi stack: " .. result.stderr
            log.error(error_message)
            return false, error_message
          end
          log.info("Pulumi stack deployed.")
          return true, "Pulumi stack deployed."
        end
      }
    }
  }
}
