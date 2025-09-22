--
-- digitalocean_example.lua
--
-- This example demonstrates how to use the digitalocean module to manage
-- resources in your DigitalOcean account.
--
-- This pipeline assumes you have the `doctl` CLI installed and configured
-- with an access token via the DIGITALOCEAN_ACCESS_TOKEN environment variable.
--
-- To run this example:
-- 1. Make sure you have a Droplet you want to delete for the test.
-- 2. Set the `droplet_to_delete_name` variable to the name of that Droplet.
-- 3. Run the pipeline:
--    go run ./cmd/sloth-runner -f examples/digitalocean_example.lua
--

local log = require("log")

-- Configuration --
-- The name of the Droplet you want to delete at the end of the pipeline.
local droplet_to_delete_name = "my-test-droplet-to-delete"
-------------------


TaskDefinitions = {
  ["digitalocean-management"] = {
    description = "A pipeline to list and manage DigitalOcean resources.",

    tasks = {
      {
        name = "list_droplets",
        description = "Lists all Droplets in the account.",
        command = function()
          log.info("Listing all DigitalOcean Droplets...")
          local droplets, err = digitalocean.droplets.list()

          if not droplets then
            log.error("Failed to list Droplets: " .. err)
            return false, "doctl list failed."
          end

          log.info("Successfully retrieved Droplet list.")
          print("--- Droplets ---")
          for _, droplet in ipairs(droplets) do
            print(string.format("ID: %d, Name: %s, Status: %s, Region: %s", droplet.id, droplet.name, droplet.status, droplet.region.slug))
          end
          print("----------------")
          
          -- Pass the droplet list to the next task
          return true, "Droplets listed.", {droplets = droplets}
        end
      },
      {
        name = "delete_specific_droplet",
        description = "Finds a specific Droplet by name and deletes it.",
        depends_on = "list_droplets",
        command = function(params, deps)
          local droplets = deps.list_droplets.droplets
          local target_droplet_id = nil

          log.info("Searching for Droplet with name: " .. droplet_to_delete_name)
          for _, droplet in ipairs(droplets) do
            if droplet.name == droplet_to_delete_name then
              target_droplet_id = droplet.id
              break
            end
          end

          if not target_droplet_id then
            log.warn("Could not find a Droplet named '" .. droplet_to_delete_name .. "' to delete. Skipping.")
            -- We return true because not finding the droplet isn't a pipeline failure.
            return true, "Target Droplet not found."
          end

          log.info("Found Droplet with ID: " .. target_droplet_id .. ". Deleting now...")
          local ok, err = digitalocean.droplets.delete({
            id = tostring(target_droplet_id),
            force = true
          })

          if not ok then
            log.error("Failed to delete Droplet: " .. err)
            return false, "Droplet deletion failed."
          end

          log.info("Successfully initiated deletion of Droplet " .. droplet_to_delete_name)
          return true, "Droplet deleted."
        end
      }
    }
  }
}
