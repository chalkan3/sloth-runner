# Sloth Runner Examples

This document provides a comprehensive, real-world example of how to use Sloth Runner to orchestrate a complex cloud deployment on Google Cloud Platform (GCP) using Git, Python, and Pulumi.

## Example 1: Creating a GCP Hub-and-Spoke Infrastructure with Pulumi

This is the main example. It demonstrates a complete orchestration pipeline that performs the following actions:
1.  Clones two separate Git repositories: one for a "hub" network and one for a "spoke" network containing a VM.
2.  Creates a dedicated Python virtual environment for the spoke project.
3.  Deploys the hub network using its Pulumi stack. This includes the main VPC and two spoke VPCs with peering.
4.  Retrieves the network outputs from the hub deployment.
5.  Deploys the spoke stack, which creates a Compute Engine VM within one of the hub's subnets.

**To run this pipeline:**
```bash
go run ./cmd/sloth-runner run --file examples/gcp_pulumi_orchestration.lua --values configs/gcp_deployment_values.yaml --yes
```

---

### **Pipeline: `examples/gcp_pulumi_orchestration.lua`**
```lua
-- examples/gcp_pulumi_orchestration.lua
--
-- This pipeline demonstrates a complete orchestration for deploying a GCP Hub and Spoke network.

TaskDefinitions = {
  gcp_deployment = {
    description = "Orchestrates the deployment of a GCP Hub and Spoke architecture.",
    tasks = {
      {
        name = "orchestrate_gcp",
        command = function()
          -- Cleanup and Setup
          -- ==============================================================================
          log.info("Cleaning up previous run artifacts...")
          fs.rm_r(values.paths.base_workdir)
          fs.mkdir(values.paths.base_workdir)

          -- Clone Git Repositories
          -- ==============================================================================
          log.info("Cloning Hub and Spoke repositories...")

          local hub_repo = git.clone(values.repos.hub.url, values.repos.hub.path)
          log.info("Hub repo cloned to: " .. hub_repo.path)

          local spoke_repo = git.clone(values.repos.spoke.url, values.repos.spoke.path)
          log.info("Spoke repo cloned to: " .. spoke_repo.path)

          -- Setup Python Virtual Environment for the Host Manager (Spoke)
          -- ==============================================================================
          log.info("Setting up Python venv for the host manager...")

          local spoke_venv = python.venv(values.paths.spoke_venv)
            :create()
            :pip("install setuptools")
            :pip("install -r " .. spoke_repo.path .. "/requirements.txt")

          log.info("Python venv for spoke is ready at: " .. values.paths.spoke_venv)

          -- Deploy GCP Hub Network (Pulumi Stack 1)
          -- ==============================================================================
          log.info("Deploying GCP Hub Network...")

          local hub_stack = pulumi.stack(values.pulumi.hub.stack_name, {
            workdir = hub_repo.path,
            login = values.pulumi.login_url
          })

          -- Configure the Hub stack from the values file
          hub_stack:select()
            :config_map(values.pulumi.hub.config)

          local hub_result = hub_stack:up({ yes = true })
          if not hub_result.success then
            log.error("Hub stack deployment failed: " .. hub_result.stdout)
            return false, "Hub stack deployment failed."
          end

          log.info("Hub stack deployed successfully.")
          local hub_outputs = hub_stack:outputs()

          -- Deploy GCP Spoke Host (Pulumi Stack 2)
          -- ==============================================================================
          log.info("Deploying GCP Spoke Host...")

          local spoke_stack = pulumi.stack(values.pulumi.spoke.stack_name, {
            workdir = spoke_repo.path,
            login = values.pulumi.login_url,
            venv = spoke_venv
          })

          -- Configure the Spoke stack, combining static values and Hub outputs
          local spoke_config = values.pulumi.spoke.config
          spoke_config.hub_network_self_link = hub_outputs.network_self_link -- Pass output from Hub

          spoke_stack:select()
            :config_map(spoke_config)

          local spoke_result = spoke_stack:up({ yes = true })
          if not spoke_result.success then
            log.error("Spoke stack deployment failed: " .. spoke_result.stdout)
            return false, "Spoke stack deployment failed."
          end

          log.info("Spoke stack deployed successfully.")
          local spoke_outputs = spoke_stack:outputs()

          -- Final Output
          -- ==============================================================================
          log.info("GCP Hub and Spoke orchestration completed successfully!")

          local final_outputs = {
            hub_outputs = hub_outputs,
            spoke_outputs = spoke_outputs
          }
          return true, "Orchestration successful", final_outputs
        end
      }
    }
  }
}
```

---

### **Configuration: `configs/gcp_deployment_values.yaml`**
```yaml
# configs/gcp_deployment_values.yaml
#
# Configuration values for the gcp_pulumi_orchestration.lua pipeline.

# Base paths for cloning the repositories and creating the Python venv.
paths:
  base_workdir: "/tmp/gcp_pulumi_orchestration"
  spoke_venv: "/tmp/gcp_pulumi_orchestration/spoke_venv"

# Git repository configurations.
repos:
  hub:
    url: "https://github.com/chalkan3/gcp-hub-spoke"
    path: "/tmp/gcp_pulumi_orchestration/hub"
  spoke:
    url: "https://github.com/chalkan3/gcp-host-manager"
    path: "/tmp/gcp_pulumi_orchestration/spoke"

# Pulumi configuration.
pulumi:
  login_url: "gs://pulumi-state-backend-chalkan3/"
  hub:
    stack_name: "hub-dev"
    config:
      gcp-hub-spoke:project: "chalkan3"
      gcp-hub-spoke:region: "us-central1"
      gcp-hub-spoke:hub-vpc:
        name: "sloth-runner-hub-network"
        cidr: "10.0.0.0/16"
        subnets:
          - name: "subnet-a"
            cidr: "10.0.1.0/24"
            region: "us-central1"
          - name: "subnet-b"
            cidr: "10.0.2.0/24"
            region: "us-central1"
      gcp-hub-spoke:spoke-vpcs:
        - name: "spoke-dev"
          cidr: "192.168.10.0/24"
          subnets:
            - name: "dev-subnet-apps"
              cidr: "192.168.10.0/26"
              region: "us-central1"
        - name: "spoke-prod"
          cidr: "192.168.20.0/24"
          subnets:
            - name: "prod-subnet-apps"
              cidr: "192.168.20.0/25"
              region: "us-east1"
  spoke:
    stack_name: "spoke-dev"
    config:
      gcp-host-manager:project: "chalkan3"
      gcp-host-manager:region: "us-central1"
      gcp-host-manager:instance_name: "sloth-runner-spoke-instance"
      gcp-host-manager:instance_type: "e2-medium"
      gcp-host-manager:subnet_name: "dev-subnet-apps"
      gcp-host-manager:hub-spoke-stack-reference: "organization/gcp-hub-spoke/hub-dev"
      # The 'hub_network_self_link' will be injected dynamically by the pipeline.
```

---

## Example 2: Destroying the GCP Infrastructure

This pipeline complements the first example by cleanly destroying all the infrastructure that was created. It runs the `pulumi destroy` command on both stacks in the correct order (spoke first, then hub) to ensure a clean teardown.

**To run this pipeline:**
```bash
go run ./cmd/sloth-runner run --file examples/gcp_host_destroy_pipeline.lua --values configs/gcp_deployment_values.yaml --yes
```

---

### **Pipeline: `examples/gcp_host_destroy_pipeline.lua`**
```lua
-- examples/gcp_host_destroy_pipeline.lua
--
-- This pipeline destroys the GCP Hub and Spoke infrastructure.

TaskDefinitions = {
  gcp_deployment_destroy = {
    description = "Destroys the GCP Hub and Spoke architecture.",
    tasks = {
      {
        name = "destroy_gcp_stacks",
        command = function()
          log.info("Destroying GCP Spoke Host...")

          local spoke_stack = pulumi.stack(values.pulumi.spoke.stack_name, {
            workdir = values.repos.spoke.path,
            login = values.pulumi.login_url,
            venv_path = values.paths.spoke_venv
          })

          local spoke_result = spoke_stack:destroy({ yes = true })
          if not spoke_result.success then
            log.error("Spoke stack destruction failed: " .. spoke_result.stdout)
            -- We continue to the hub destruction even if the spoke fails
          else
            log.info("Spoke stack destroyed successfully.")
          end

          log.info("Destroying GCP Hub Network...")

          local hub_stack = pulumi.stack(values.pulumi.hub.stack_name, {
            workdir = values.repos.hub.path,
            login = values.pulumi.login_url
          })

          local hub_result = hub_stack:destroy({ yes = true })
          if not hub_result.success then
            log.error("Hub stack destruction failed: " .. hub_result.stdout)
            return false, "Hub stack destruction failed."
          end

          log.info("Hub stack destroyed successfully.")
          log.info("GCP Hub and Spoke destruction completed successfully!")

          return true, "Destruction successful"
        end
      }
    }
  }
}
```
*Note: This pipeline uses the same `configs/gcp_deployment_values.yaml` file as the creation pipeline.*