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
        command = function(params)
          local git = require("git")
          local result = git.clone("https://github.com/chalkan3/gcp-hosts.git", params.workdir)
          if not result.success then
            return false, "Failed to clone repository: " .. result.stderr
          end
          return true, "Repository cloned successfully."
        end
      },
      {
        name = "setup_python_env",
        description = "Creates a virtual environment and installs all dependencies.",
        depends_on = "clone_repository",
        command = function(params)
          local py = require("python")
          local venv_path = params.workdir .. "/.venv"
          local venv = py.venv(venv_path)
          
          local create_result = venv:create()
          if not create_result.success then
            return false, "Failed to create python venv: " .. create_result.stderr
          end
          
          local pip_result = venv:pip("install -r " .. params.workdir .. "/requirements.txt")
          if not pip_result.success then
            return false, "Failed to install python dependencies: " .. pip_result.stderr
          end

          return true, "Python environment created.", { venv_path = venv_path }
        end
      },
      {
        name = "init_stack",
        description = "Initializes the Pulumi stack if it does not exist.",
        depends_on = "setup_python_env",
        command = function(params, inputs)
          local pulumi = require("pulumi")
          local stack = pulumi.stack("organization/gcp-host/prod", { workdir = params.workdir, venv_path = inputs.setup_python_env.venv_path })
          local result = stack:select()
          if not result.success then
            return false, "Failed to select pulumi stack: " .. result.stderr
          end
          return true, "Pulumi stack selected.", { stack = stack, venv_path = inputs.setup_python_env.venv_path }
        end
      },
      {
        name = "configure_pulumi",
        description = "Sets all required Pulumi config values for the gcp-hosts stack.",
        depends_on = "init_stack",
        command = function(params, inputs)
          local stack = inputs.init_stack.stack
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
            local result = stack:config(conf[1], conf[2])
            if not result.success then
              return false, "Failed to set config '" .. conf[1] .. "': " .. result.stderr
            end
          end

          return true, "Stack Configurated", { stack = stack }
        end
      },
      {
        name = "deploy_stack",
        description = "Deploys the Pulumi stack using the prepared environment.",
        depends_on = "configure_pulumi",
        command = function(params, inputs)
          local stack = inputs.configure_pulumi.stack
          local result = stack:up({ yes = true, skip_preview = true })
          if not result.success then
            return false, "Failed to deploy pulumi stack: " .. result.stdout
          end
          return true, "Pulumi stack deployed."
        end
      }
    }
  }
}
