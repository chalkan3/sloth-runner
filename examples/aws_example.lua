--
-- aws_example.lua
--
-- This example demonstrates how to use the aws module to interact with
-- various AWS services. It shows how to use the generic `aws.exec` function
-- as well as the higher-level helpers for S3 and Secrets Manager.
--
-- This pipeline assumes you have AWS credentials configured in your environment
-- or are using aws-vault.
--
-- To run this example:
-- 1. (Optional) If using aws-vault, ensure it's configured with a profile.
-- 2. (Optional) Create an S3 bucket and a test secret in Secrets Manager.
-- 3. Edit the `aws_profile`, `s3_bucket`, and `secret_id` variables below.
-- 4. Run the pipeline:
--    go run ./cmd/sloth-runner -f examples/aws_example.lua
--

local log = require("log")

-- Configuration --
-- Set your AWS profile for aws-vault, or leave as "" to use default credentials.
local aws_profile = ""
-- Set your S3 bucket name for the sync example.
local s3_bucket = "your-s3-bucket-name"
-- Set the name of your secret in AWS Secrets Manager.
local secret_id = "your/secret/name"
-------------------


TaskDefinitions = {
  ["aws-examples"] = {
    description = "A pipeline demonstrating various AWS module functions.",
    create_workdir_before_run = true,
    clean_workdir_after_run = function(r) return r.success end,

    tasks = {
      {
        name = "check_aws_identity",
        description = "Verifies AWS credentials by calling sts get-caller-identity.",
        command = function()
          log.info("Checking AWS identity...")
          local result = aws.exec({"sts", "get-caller-identity"}, {profile = aws_profile})

          if result.exit_code ~= 0 then
            log.error("Failed to check AWS identity: " .. result.stderr)
            return false, "AWS identity check failed."
          end

          log.info("Successfully identified AWS user/role:")
          print(result.stdout)
          return true, "AWS identity verified."
        end
      },
      {
        name = "sync_files_to_s3",
        description = "Creates a local file and syncs it to an S3 bucket.",
        depends_on = "check_aws_identity",
        command = function(params)
          local workdir = params.workdir
          local file_path = workdir .. "/hello.txt"
          fs.write(file_path, "Hello from the Sloth-Runner AWS module!")
          log.info("Created local file: " .. file_path)

          log.info("Syncing local directory to s3://" .. s3_bucket .. "/test-sync/")
          local ok, err = aws.s3.sync({
            source = workdir,
            destination = "s3://" .. s3_bucket .. "/test-sync/",
            profile = aws_profile,
            delete = true
          })

          if not ok then
            log.error("Failed to sync to S3: " .. err)
            return false, "S3 sync failed."
          end

          log.info("S3 sync completed successfully.")
          return true, "Files synced to S3."
        end
      },
      {
        name = "get_secret_value",
        description = "Retrieves a secret from AWS Secrets Manager.",
        depends_on = "check_aws_identity",
        command = function()
          log.info("Attempting to retrieve secret: " .. secret_id)
          local secret_string, err = aws.secretsmanager.get_secret({
            secret_id = secret_id,
            profile = aws_profile
          })

          if not secret_string then
            log.error("Failed to retrieve secret: " .. err)
            return false, "Secret retrieval failed."
          end

          log.info("Successfully retrieved secret!")
          -- IMPORTANT: Be careful not to print the actual secret in production logs.
          -- This is just for demonstration.
          log.info("Secret Value (first 10 chars): " .. string.sub(secret_string, 1, 10) .. "...")
          
          -- You can now use this secret in subsequent steps.
          return true, "Secret retrieved."
        end
      }
    }
  }
}
