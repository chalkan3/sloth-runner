-- examples/gcp_pulumi_orchestration.lua
--
-- This pipeline demonstrates a complete, modular orchestration for deploying a GCP Hub and Spoke network.

TaskDefinitions = {
  gcp_deployment = {
    description = "Orchestrates the deployment of a GCP Hub and Spoke architecture.",
    tasks = {
      {
        name = "setup_workspace",
        command = function()
          log.info("Cleaning up previous run artifacts...")
          fs.rm_r(values.paths.base_workdir)
          fs.mkdir(values.paths.base_workdir)
          return true, "Workspace cleaned and created."
        end
      },
      {
        name = "clone_hub_repo",
        depends_on = "setup_workspace",
        command = function()
          log.info("Cloning Hub repository...")
          local hub_repo = git.clone(values.repos.hub.url, values.repos.hub.path)
          log.info("Hub repo cloned to: " .. hub_repo.path)
          return true, "Hub repo cloned.", { repo = hub_repo }
        end
      },
      {
        name = "clone_spoke_repo",
        depends_on = "setup_workspace",
        command = function()
          log.info("Cloning Spoke repository...")
          local spoke_repo = git.clone(values.repos.spoke.url, values.repos.spoke.path)
          log.info("Spoke repo cloned to: " .. spoke_repo.path)
          return true, "Spoke repo cloned.", { repo = spoke_repo }
        end
      },
      {
        name = "setup_spoke_venv",
        depends_on = "clone_spoke_repo",
        command = function(inputs)
          log.info("Setting up Python venv for the host manager...")
          local spoke_repo = inputs.clone_spoke_repo.repo
          local spoke_venv = python.venv(values.paths.spoke_venv)
            :create()
            :pip("install setuptools")
            :pip("install -r " .. spoke_repo.path .. "/requirements.txt")
          log.info("Python venv for spoke is ready at: " .. values.paths.spoke_venv)
          return true, "Spoke venv created.", { venv = spoke_venv }
        end
      },
      {
        name = "deploy_hub_stack",
        depends_on = "clone_hub_repo",
        command = function(inputs)
          log.info("Deploying GCP Hub Network...")
          local hub_repo = inputs.clone_hub_repo.repo
          local hub_stack = pulumi.stack(values.pulumi.hub.stack_name, {
            workdir = hub_repo.path,
            login = values.pulumi.login_url
          })
          hub_stack:select():config_map(values.pulumi.hub.config)
          local hub_result = hub_stack:up({ yes = true })
          if not hub_result.success then
            log.error("Hub stack deployment failed: " .. hub_result.stdout)
            return false, "Hub stack deployment failed."
          end
          log.info("Hub stack deployed successfully.")
          local hub_outputs = hub_stack:outputs()
          return true, "Hub stack deployed.", { outputs = hub_outputs }
        end
      },
      {
        name = "deploy_spoke_stack",
        depends_on = { "setup_spoke_venv", "deploy_hub_stack" },
        command = function(inputs)
          log.info("Deploying GCP Spoke Host...")
          local spoke_repo = inputs.setup_spoke_venv.repo
          local spoke_venv = inputs.setup_spoke_venv.venv
          local hub_outputs = inputs.deploy_hub_stack.outputs

          local spoke_stack = pulumi.stack(values.pulumi.spoke.stack_name, {
            workdir = spoke_repo.path,
            login = values.pulumi.login_url,
            venv = spoke_venv
          })

          local spoke_config = values.pulumi.spoke.config
          spoke_config.hub_network_self_link = hub_outputs.network_self_link

          spoke_stack:select():config_map(spoke_config)
          local spoke_result = spoke_stack:up({ yes = true })
          if not spoke_result.success then
            log.error("Spoke stack deployment failed: " .. spoke_result.stdout)
            return false, "Spoke stack deployment failed."
          end
          log.info("Spoke stack deployed successfully.")
          local spoke_outputs = spoke_stack:outputs()
          return true, "Spoke stack deployed.", { outputs = spoke_outputs }
        end
      },
      {
          name = "final_summary",
          depends_on = "deploy_spoke_stack",
          command = function(inputs)
              log.info("GCP Hub and Spoke orchestration completed successfully!")
              -- You can access outputs from dependencies like this:
              -- local hub_outputs = inputs.deploy_hub_stack.outputs
              -- local spoke_outputs = inputs.deploy_spoke_stack.outputs
              return true, "Orchestration successful."
          end
      }
    }
  }
}