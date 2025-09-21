-- examples/gcp_list_instances.lua

-- This example demonstrates how to list GCP compute instances using the gcp module.
-- It defines a task that calls 'gcloud compute instances list' for a specific project.

TaskDefinitions = {
  main = {
    description = "A task group to list GCP compute instances.",
    tasks = {
      {
        name = "list-gcp-instances",
        description = "Lists GCP compute instances for project 'chalkan3'.",
        command = function(params, inputs)
          log.info("Attempting to list GCP compute instances for project 'chalkan3'...")

          local gcp = require("gcp")

          -- Define the arguments for the gcloud command
          local gcloud_args = {"compute", "instances", "list", "--project", "chalkan3"}

          -- Execute the gcloud command using the gcp module
          local result = gcp.exec(gcloud_args)

          -- Check the result of the execution
          if result and result.exit_code == 0 then
            log.info("Successfully listed GCP compute instances:")
            -- Print the standard output of the command, which contains the instance list
            print(result.stdout)
            return true, "Successfully listed GCP compute instances."
          else
            log.error("Failed to list GCP compute instances.")
            if result then
              log.error("Exit Code: " .. result.exit_code)
              log.error("Stderr: " .. result.stderr)
            end
            return false, "Failed to list GCP compute instances."
          end
        end
      }
    }
  }
}
