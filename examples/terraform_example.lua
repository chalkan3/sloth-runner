--
-- terraform_example.lua
--
-- This example demonstrates the full lifecycle of a Terraform project
-- orchestrated by the `terraform` module. It uses a simple local file
-- as a resource, so no cloud credentials are required.
--
-- The pipeline will:
-- 1. Initialize Terraform.
-- 2. Create an execution plan.
-- 3. Apply the plan to create a local file (`report.txt`).
-- 4. Read the output variable from Terraform to get the filename.
-- 5. Use the `fs` module to read the content of the created file.
-- 6. Destroy the Terraform-managed resources.
--
-- To run this example:
--    go run ./cmd/sloth-runner -f examples/terraform_example.lua
--

local log = require("log")

-- The working directory for all Terraform commands.
local tf_workdir = "./examples/terraform"

TaskDefinitions = {
  ["terraform-lifecycle"] = {
    description = "A pipeline to manage a Terraform project.",

    tasks = {
      {
        name = "init",
        description = "Initializes the Terraform project.",
        command = function()
          log.info("Running terraform init...")
          local result = terraform.init({workdir = tf_workdir})
          if not result.success then
            log.error("Terraform init failed: " .. result.stderr)
            return false, "Terraform init failed."
          end
          log.info("Terraform init successful.")
          return true, "Terraform initialized."
        end
      },
      {
        name = "plan",
        description = "Creates a Terraform execution plan.",
        depends_on = "init",
        command = function()
          log.info("Running terraform plan...")
          local result = terraform.plan({workdir = tf_workdir})
          if not result.success then
            log.error("Terraform plan failed: " .. result.stderr)
            return false, "Terraform plan failed."
          end
          log.info("Terraform plan successful.")
          print(result.stdout)
          return true, "Terraform plan created."
        end
      },
      {
        name = "apply",
        description = "Applies the Terraform plan.",
        depends_on = "plan",
        command = function()
          log.info("Running terraform apply...")
          local result = terraform.apply({workdir = tf_workdir, auto_approve = true})
          if not result.success then
            log.error("Terraform apply failed: " .. result.stderr)
            return false, "Terraform apply failed."
          end
          log.info("Terraform apply successful.")
          return true, "Terraform apply complete."
        end
      },
      {
        name = "get_output",
        description = "Reads the output variables from Terraform.",
        depends_on = "apply",
        command = function()
          log.info("Reading Terraform output...")
          local filename, err = terraform.output({workdir = tf_workdir, name = "report_filename"})
          if not filename then
            log.error("Failed to get Terraform output: " .. err)
            return false, "Terraform output failed."
          end
          
          log.info("Got filename from output: " .. filename)
          
          -- Read the content of the file created by Terraform
          local content, read_err = fs.read(filename)
          if read_err then
            log.error("Failed to read the report file: " .. read_err)
            return false, "Could not read artifact."
          end

          log.info("Successfully read content from Terraform-generated file:")
          print("--- Report Content ---")
          print(content)
          print("----------------------")

          return true, "Terraform output processed."
        end
      },
      {
        name = "destroy",
        description = "Destroys the Terraform-managed resources.",
        depends_on = "get_output",
        command = function()
          log.warn("Running terraform destroy...")
          local result = terraform.destroy({workdir = tf_workdir, auto_approve = true})
          if not result.success then
            log.error("Terraform destroy failed: " .. result.stderr)
            return false, "Terraform destroy failed."
          end
          log.info("Terraform destroy successful.")
          return true, "Terraform resources destroyed."
        end
      }
    }
  }
}
