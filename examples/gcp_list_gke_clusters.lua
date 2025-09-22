-- examples/gcp_list_gke_clusters.lua
--
-- This pipeline demonstrates how to list GKE clusters.

TaskDefinitions = {
  gcp_gke_lister = {
    description = "Lists all GKE clusters in a given project.",
    tasks = {
      {
        name = "list_gke_clusters",
        command = function()
          log.info("Listing GKE clusters...")

          local result = gcp.client({ project = "chalkan3" })
            :gke()
            :clusters()
            :list()

          if not result.success then
            log.error("Failed to list GKE clusters: " .. result.stderr)
            return false, "Failed to list GKE clusters."
          end

          log.info("Successfully listed GKE clusters.")

          local clusters, err = data.parse_json(result.stdout)
          if err then
            log.error("Failed to decode JSON response: " .. err)
            log.info("Raw output: " .. result.stdout)
            return false, "Failed to parse GKE cluster list."
          end

          if #clusters == 0 then
            log.info("No GKE clusters found in project.")
          else
            log.info("Found " .. #clusters .. " GKE cluster(s):")
            for i, cluster in ipairs(clusters) do
              log.info("  - " .. cluster.name .. " (Status: " .. cluster.status .. ", Location: " .. cluster.location .. ")")
            end
          end

          return true, "GKE clusters listed successfully."
        end
      }
    }
  }
}
