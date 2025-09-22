-- examples/gcp_list_buckets.lua
--
-- This pipeline demonstrates how to list GCP Storage buckets.

TaskDefinitions = {
  gcp_bucket_lister = {
    description = "Lists all GCP Storage buckets in a given project.",
    tasks = {
      {
        name = "list_gcp_buckets",
        command = function()
          log.info("Listing GCP Storage buckets...")

          local result = gcp.client({ project = "chalkan3" })
            :storage()
            :buckets()
            :list()

          if not result.success then
            log.error("Failed to list buckets: " .. result.stderr)
            return false, "Failed to list buckets."
          end

          log.info("Successfully listed buckets.")

          local buckets, err = data.parse_json(result.stdout)
          if err then
            log.error("Failed to decode JSON response: " .. err)
            log.info("Raw output: " .. result.stdout)
            return false, "Failed to parse bucket list."
          end

          if #buckets == 0 then
            log.info("No buckets found in project.")
          else
            log.info("Found " .. #buckets .. " bucket(s):")
            for i, bucket in ipairs(buckets) do
              log.info("  - " .. bucket.name .. " (Location: " .. bucket.location .. ")")
            end
          end

          return true, "Buckets listed successfully."
        end
      }
    }
  }
}
