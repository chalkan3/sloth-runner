-- examples/gcp_host_destroy_pipeline.lua
--
-- This pipeline destroys the GCP Hub and Spoke infrastructure in a modular way.

TaskDefinitions = {
  gcp_deployment_destroy = {
    description = "Destroys the GCP Hub and Spoke architecture.",
    tasks = {
      {
        name = "destroy_spoke_stack",
        command = function()
          log.info("Destroying GCP Spoke Host...")
          local spoke_stack = pulumi.stack(values.pulumi.spoke.stack_name, {
            workdir = values.repos.spoke.path,
            login = values.pulumi.login_url,
            venv_path = values.paths.spoke_venv
          })
          local spoke_result = spoke_stack:destroy({ yes = true })
          if not spoke_result.success then
            log.error("Spoke stack destruction failed: " .. spoke_result.stdout)
            -- Return false to halt the pipeline if spoke destruction fails
            return false, "Spoke stack destruction failed."
          end
          log.info("Spoke stack destroyed successfully.")
          return true, "Spoke stack destroyed."
        end
      },
      {
        name = "destroy_hub_stack",
        depends_on = "destroy_spoke_stack",
        command = function()
          log.info("Destroying GCP Hub Network...")
          local hub_stack = pulumi.stack(values.pulumi.hub.stack_name, {
            workdir = values.repos.hub.path,
            login = values.pulumi.login_url
          })
          local hub_result = hub_stack:destroy({ yes = true })
          if not hub_result.success then
            log.error("Hub stack destruction failed: " .. hub_result.stdout)
            return false, "Hub stack destruction failed."
          end
          log.info("Hub stack destroyed successfully.")
          return true, "Hub stack destroyed."
        end
      },
      {
          name = "final_summary",
          depends_on = "destroy_hub_stack",
          command = function()
              log.info("GCP Hub and Spoke destruction completed successfully!")
              return true, "Destruction successful."
          end
      }
    }
  }
}
