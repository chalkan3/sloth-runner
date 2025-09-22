-- examples/gcp_pulumi_orchestration.lua
--
-- This pipeline demonstrates a complete orchestration for deploying a GCP Hub and Spoke network.

TaskDefinitions = {
  gcp_deployment = {
    description = "Orchestrates the deployment of a GCP Hub and Spoke architecture.",
    tasks = {
      {
        name = "orchestrate_gcp",
        command = function()
          -- Cleanup and Setup
          -- ==============================================================================
          log.info("Cleaning up previous run artifacts...")
          fs.rm_r(values.paths.base_workdir)
          fs.mkdir(values.paths.base_workdir)

          -- Clone Git Repositories
          -- ==============================================================================
          log.info("Cloning Hub and Spoke repositories...")

          local hub_repo = git.clone(values.repos.hub.url, values.repos.hub.path)
          log.info("Hub repo cloned to: " .. hub_repo.path)

          local spoke_repo = git.clone(values.repos.spoke.url, values.repos.spoke.path)
          log.info("Spoke repo cloned to: " .. spoke_repo.path)

          -- Setup Python Virtual Environment for the Host Manager (Spoke)
          -- ==============================================================================
          log.info("Setting up Python venv for the host manager...")

          local spoke_venv = python.venv(values.paths.spoke_venv)
            :create()
            :pip("install -r " .. spoke_repo.path .. "/requirements.txt")

          log.info("Python venv for spoke is ready at: " .. values.paths.spoke_venv)

          -- Deploy GCP Hub Network (Pulumi Stack 1)
          -- ==============================================================================
          log.info("Deploying GCP Hub Network...")

          local hub_stack = pulumi.stack(values.pulumi.hub.stack_name, {
            workdir = hub_repo.path,
            login = values.pulumi.login_url
          })

          -- Configure the Hub stack from the values file
          hub_stack:select()
            :config_map(values.pulumi.hub.config)

          local hub_result = hub_stack:up({ yes = true })
          if not hub_result.success then
            log.error("Hub stack deployment failed: " .. hub_result.stdout)
            return false, "Hub stack deployment failed."
          end

          log.info("Hub stack deployed successfully.")
          local hub_outputs = hub_stack:outputs()

          -- Deploy GCP Spoke Host (Pulumi Stack 2)
          -- ==============================================================================
          log.info("Deploying GCP Spoke Host...")

          local spoke_stack = pulumi.stack(values.pulumi.spoke.stack_name, {
            workdir = spoke_repo.path,
            login = values.pulumi.login_url,
            venv = spoke_venv
          })

          -- Configure the Spoke stack, combining static values and Hub outputs
          local spoke_config = values.pulumi.spoke.config
          spoke_config.hub_network_self_link = hub_outputs.network_self_link -- Pass output from Hub

          spoke_stack:select()
            :config_map(spoke_config)

          local spoke_result = spoke_stack:up({ yes = true })
          if not spoke_result.success then
            log.error("Spoke stack deployment failed: " .. spoke_result.stderr)
            return false, "Spoke stack deployment failed."
          end

          log.info("Spoke stack deployed successfully.")
          local spoke_outputs = spoke_stack:outputs()

          -- Final Output
          -- ==============================================================================
          log.info("GCP Hub and Spoke orchestration completed successfully!")

          local final_outputs = {
            hub_outputs = hub_outputs,
            spoke_outputs = spoke_outputs
          }
          return true, "Orchestration successful", final_outputs
        end
      }
    }
  }
}
