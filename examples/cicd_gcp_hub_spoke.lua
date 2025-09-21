-- examples/cicd_gcp_hub_spoke.lua
--
-- This example demonstrates a full CI/CD pipeline for a Pulumi project.
-- It performs the following steps:
-- 1. Clones a Git repository containing a GCP Hub-Spoke infrastructure definition.
-- 2. Creates a Python virtual environment.
-- 3. Installs the required Python dependencies using pip.
-- 4. Sets all required Pulumi config values programmatically.
-- 5. Runs `pulumi up` to deploy the infrastructure.

TaskDefinitions = {
  ["gcp_hub_spoke_deploy"] = {
    description = "CI/CD Pipeline: Clones, sets up, configures, and deploys the gcp-hub-spoke Pulumi project.",

    tasks = {
      {
        name = "clone_repository",
        description = "Clones the gcp-hub-spoke project repository into a new temp dir.",
        command = function()
          local temp_dir, err = fs.tmpname()
          if err then
            return false, "Failed to get temporary directory name: " .. err
          end
          fs.mkdir(temp_dir)
          log.info("Created temporary directory for clone: " .. temp_dir)

          local repo_url = "https://github.com/chalkan3/gcp-hub-spoke.git"
          local result = require("git").clone(repo_url, temp_dir)

          if not result.success then
            local err_msg = "Failed to clone repository"
            if result.stderr then
              err_msg = err_msg .. ": " .. result.stderr
            end
            log.error(err_msg)
            return false, "Git clone failed", { workdir = temp_dir }
          end

          log.info("Repository cloned successfully.")
          return true, "Repository cloned.", { workdir = temp_dir }
        end
      },
      {
        name = "setup_and_deploy",
        description = "Installs dependencies, sets config, and deploys the Pulumi stack.",
        depends_on = "clone_repository",
        command = function(params, inputs)
          local workdir = inputs.clone_repository.workdir
          if not workdir then
            return false, "Workdir not received from clone_repository task."
          end
          log.info("Running deployment from: " .. workdir)

          -- 1. Set up Python virtual environment
          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          venv:create()
          venv:pip("install -r " .. workdir .. "/requirements.txt")

          -- 2. Set up Pulumi
          local pulumi = require("pulumi")
          local stack = pulumi.stack("dev", { 
            workdir = workdir, 
            venv = venv,
            login_url = 'gs://pulumi-state-backend-chalkan3'
          })

          -- Helper to convert Lua table to JSON string for Pulumi config
          local function to_json(tbl)
            local json_str, err = require("data").to_json(tbl)
            if err then
              error("Failed to serialize to JSON: " .. err)
            end
            return json_str
          end

          -- 3. Set Pulumi configuration values
          log.info("Setting Pulumi configuration...")
          stack:config("gcp-hub-spoke:project", "chalkan3")
          stack:config("gcp-hub-spoke:hub_vpc", to_json({
            name = "hub-vpc",
            cidr = "10.0.0.0/16",
            subnets = {
              { name = "subnet-a", cidr = "10.0.1.0/24", region = "us-central1" }
            }
          }))
          stack:config("gcp-hub-spoke:spoke_vpcs", to_json({
            {
              name = "spoke-vpc-1",
              cidr = "192.168.1.0/24",
              subnets = {
                { name = "subnet-1a", cidr = "192.168.1.0/28", region = "us-central1" }
              }
            },
            {
              name = "spoke-vpc-2",
              cidr = "192.168.2.0/24",
              subnets = {
                { name = "subnet-2a", cidr = "192.168.2.0/28", region = "us-east1" }
              }
            }
          }))
          stack:config("gcp-hub-spoke:vms", to_json({
            {
              name = "salt-minion-spoke1",
              machine_type = "e2-medium",
              boot_disk_image = "ubuntu-os-cloud/ubuntu-2204-lts",
              target_spoke_name = "spoke-vpc-1",
              zone = "us-central1-a",
              salt_master_ip = "34.57.154.158",
              salt_grains = {
                roles = { "web-server" },
                environment = "production"
              },
              open_ports = { 22 }
            }
          }))
          
          -- 4. Run Pulumi Up
          log.info("Running 'pulumi up'...")
          local up_result = stack:up({ yes = true })

          if not up_result.success then
            log.error("'pulumi up' failed. See output below.")
            log.error("STDOUT: " .. up_result.stdout)
            log.error("STDERR: " .. up_result.stderr)
            return false, "Pulumi up failed"
          end

          log.info("Pulumi up completed successfully.")
          return true, "Deployment finished successfully.", stack:outputs()
        end
      }
    }
  }
}