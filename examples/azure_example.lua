--
-- azure_example.lua
--
-- This example demonstrates how to use the azure module to manage
-- resources in your Azure subscription.
--
-- This pipeline assumes you have the `az` CLI installed and are logged in
-- via `az login`.
--
-- To run this example:
-- 1. Create a resource group in Azure for testing.
-- 2. (Optional) Create a VM inside that resource group.
-- 3. Set the `resource_group_to_delete` variable to the name of your test RG.
-- 4. Run the pipeline:
--    go run ./cmd/sloth-runner -f examples/azure_example.lua
--

local log = require("log")

-- Configuration --
-- The name of the Resource Group to list VMs from and then delete.
local resource_group_to_delete = "my-sloth-runner-test-rg"
-------------------


TaskDefinitions = {
  ["azure-management"] = {
    description = "A pipeline to list VMs and manage Azure Resource Groups.",

    tasks = {
      {
        name = "list_vms_in_group",
        description = "Lists all Virtual Machines in a specific resource group.",
        command = function()
          log.info("Listing all VMs in resource group: " .. resource_group_to_delete)
          local vms, err = azure.vm.list({resource_group = resource_group_to_delete})

          if not vms then
            log.error("Failed to list VMs: " .. err)
            return false, "az vm list failed."
          end

          log.info("Successfully retrieved VM list.")
          if #vms == 0 then
            log.info("No VMs found in resource group '" .. resource_group_to_delete .. "'.")
          else
            print("--- VMs in " .. resource_group_to_delete .. " ---")
            for _, vm in ipairs(vms) do
              print(string.format("Name: %s, Location: %s, Power State: %s", vm.name, vm.location, vm.powerState))
            end
            print("--------------------")
          end
          
          return true, "VMs listed."
        end
      },
      {
        name = "delete_resource_group",
        description = "Deletes the specified resource group.",
        depends_on = "list_vms_in_group",
        command = function()
          log.warn("Proceeding to delete resource group: " .. resource_group_to_delete)
          
          local ok, err = azure.rg.delete({
            name = resource_group_to_delete,
            yes = true -- Bypasses the interactive confirmation
          })

          if not ok then
            log.error("Failed to delete resource group: " .. err)
            return false, "Resource group deletion failed."
          end

          log.info("Successfully initiated deletion of resource group: " .. resource_group_to_delete)
          return true, "Resource group deleted."
        end
      }
    }
  }
}
