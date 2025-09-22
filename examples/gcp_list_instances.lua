-- examples/gcp_list_instances.lua
--
-- This pipeline demonstrates how to list GCP compute instances.

TaskDefinitions = {
  gcp_instance_lister = {
    description = "Lists all GCP compute instances in a given project and zone.",
    tasks = {
      {
        name = "list_gcp_instances",
        command = function()
          log.info("Listing GCP instances...")

          -- You can get a list of zones by running: gcloud compute zones list
          local zone = "us-central1-a"

          -- Instantiate the GCP client and then chain calls to list instances
          local result = gcp.client({ project = "chalkan3" })
            :compute({ zone = zone })
            :instances()
            :list()

          if not result.success then
            log.error("Failed to list instances: " .. result.stderr)
            return false, "Failed to list instances."
          end

          log.info("Successfully listed instances in zone " .. zone .. ".")

          -- The result.stdout is a JSON string. We can decode it for better logging.
          local instances, err = data.parse_json(result.stdout)
          if err then
            log.error("Failed to decode JSON response: " .. err)
            -- Still, let's print the raw output
            log.info("Raw output: " .. result.stdout)
            return false, "Failed to parse instance list."
          end

          if #instances == 0 then
            log.info("No instances found in zone " .. zone .. ".")
          else
            log.info("Found " .. #instances .. " instance(s):")
            for i, instance in ipairs(instances) do
              log.info("  - " .. instance.name .. " (" .. instance.status .. ")")
            end
          end

          return true, "Instances listed successfully."
        end
      }
    }
  }
}