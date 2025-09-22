-- examples/gcp_list_sql_instances.lua
--
-- This pipeline demonstrates how to list GCP Cloud SQL instances.

TaskDefinitions = {
  gcp_sql_lister = {
    description = "Lists all GCP Cloud SQL instances in a given project.",
    tasks = {
      {
        name = "list_gcp_sql_instances",
        command = function()
          log.info("Listing GCP Cloud SQL instances...")

          local result = gcp.client({ project = "chalkan3" })
            :sql()
            :instances()
            :list()

          if not result.success then
            log.error("Failed to list SQL instances: " .. result.stderr)
            return false, "Failed to list SQL instances."
          end

          log.info("Successfully listed SQL instances.")

          local instances, err = data.parse_json(result.stdout)
          if err then
            log.error("Failed to decode JSON response: " .. err)
            log.info("Raw output: " .. result.stdout)
            return false, "Failed to parse SQL instance list."
          end

          if #instances == 0 then
            log.info("No SQL instances found in project.")
          else
            log.info("Found " .. #instances .. " SQL instance(s):")
            for i, instance in ipairs(instances) do
              log.info("  - " .. instance.name .. " (DB Version: " .. instance.databaseVersion .. ", Region: " .. instance.region .. ")")
            end
          end

          return true, "SQL instances listed successfully."
        end
      }
    }
  }
}
