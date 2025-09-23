-- examples/unified_fluent_workflow.lua
--
-- This pipeline demonstrates a complex, real-world workflow using the new
-- fluent, object-oriented API.
--
-- Workflow:
-- 1. Clones two Git repositories: one for infrastructure (hub) and one for an application.
-- 2. Sets up a Python virtual environment for the application.
-- 3. Deploys a foundational GCP network (Hub) using Pulumi.
-- 4. Passes the network details (outputs) to a second Pulumi stack.
-- 5. Deploys a GCP host/application (Spoke) into the network.
-- 6. Uses the AWS module to sync application assets to an S3 bucket.

-- Clone Git Repositories
-- ==============================================================================
log.info("Cloning infrastructure and application repositories...")

local hub_repo = git.clone(values.repos.hub.url, values.repos.hub.path)
log.info("Hub repo cloned at: " .. hub_repo.path)

local spoke_repo = git.clone(values.repos.spoke.url, values.repos.spoke.path)
log.info("Spoke repo cloned at: " .. spoke_repo.path)

-- Setup Python Virtual Environment for the Spoke Application
-- ==============================================================================
log.info("Setting up Python virtual environment for the Spoke application...")

local spoke_venv = python.venv(values.paths.spoke_venv)
  :create()
  :pip("install -r " .. spoke_repo.path .. "/requirements.txt")

log.info("Python venv created at: " .. spoke_venv.path)

-- Deploy GCP Hub Network (Pulumi Stack 1)
-- ==============================================================================
log.info("Deploying GCP Hub Network...")

local hub_stack = pulumi.stack("hub-network", {
  workdir = hub_repo.path,
  login = values.pulumi.login_url
})

-- Configure the Hub stack using a map
hub_stack:config_map({
  ["gcp:project"] = values.gcp.project,
  ["gcp:region"] = values.gcp.region,
  hub_network_name = "sloth-hub-network"
})

-- Run `pulumi up` and capture the result
local hub_result = hub_stack:up({ yes = true })
if not hub_result.success then
  log.error("Hub stack deployment failed: " .. hub_result.stderr)
  -- You might want to exit or handle the failure
end

log.info("Hub stack deployment successful.")
local hub_outputs = hub_stack:outputs()

-- Deploy GCP Spoke Host (Pulumi Stack 2)
-- ==============================================================================
log.info("Deploying GCP Spoke Host...")

local spoke_stack = pulumi.stack("spoke-host", {
  workdir = spoke_repo.path,
  login = values.pulumi.login_url,
  venv = spoke_venv -- Associate the Python venv with this stack
})

-- Configure the Spoke stack, passing outputs from the Hub stack
spoke_stack:config_map({
  ["gcp:project"] = values.gcp.project,
  ["gcp:region"] = values.gcp.region,
  hub_network_self_link = hub_outputs.network_self_link
})

local spoke_result = spoke_stack:up({ yes = true })
if not spoke_result.success then
  log.error("Spoke stack deployment failed: " .. spoke_result.stderr)
end

log.info("Spoke stack deployment successful.")
local spoke_outputs = spoke_stack:outputs()

-- Sync Application Assets to AWS S3
-- ==============================================================================
-- log.info("Syncing application assets to S3...")
-- 
-- local aws = aws.client({ profile = values.aws.profile })
-- 
-- aws:s3():sync({
--   from = spoke_repo.path .. "/app/static",
--   to = "s3://" .. values.aws.s3_bucket_name .. "/static",
--   delete = true
-- })
-- 
-- log.info("Sync to S3 complete.")

-- Final Output
-- ==============================================================================
log.info("Unified fluent workflow completed successfully!")

return {
  hub_network = hub_outputs,
  spoke_host = spoke_outputs
}