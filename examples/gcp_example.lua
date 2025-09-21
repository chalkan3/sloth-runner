-- examples/gcp_example.lua

-- This example demonstrates how to use the new gcp module to execute gcloud commands.
-- It defines a simple task that lists the active gcloud configuration.

TaskDefinitions = {
  main = {
    description = "A task group to demonstrate GCP CLI integration.",
    tasks = {
      {
        name = "list-gcloud-config",
        description = "Lists the current gcloud configuration.",
        command = function(params, inputs)
          log.info("Attempting to list gcloud config...")

          -- Execute the gcloud command using the gcp module
          local result = gcp.exec({"config", "list"})

          -- Check the result of the execution
          if result and result.exit_code == 0 then
            log.info("Successfully listed gcloud config:")
            -- Print the standard output of the command
            print(result.stdout)
            return true, "Successfully executed gcloud config list."
          else
            log.error("Failed to execute gcloud config list.")
            if result then
              log.error("Stderr: " .. result.stderr)
            end
            return false, "Failed to execute gcloud config list."
          end
        end
      }
    }
  }
}
