# Pulumi Module

The `pulumi` module in Sloth-Runner allows you to orchestrate your Pulumi stacks directly from your Lua scripts. This is ideal for Infrastructure as Code (IaC) workflows where you need to provision, update, or destroy cloud resources as part of a larger automation pipeline.

## Common Use Cases

*   **Dynamic Provisioning:** Create staging or test environments on demand.
*   **Infrastructure Updates:** Automate the deployment of new versions of your infrastructure.
*   **Environment Management:** Destroy environments after use to save costs.
*   **CI/CD Integration:** Execute `pulumi up` or `preview` as part of a CI/CD pipeline.

## API Reference

### `pulumi.stack(name, options_table)`

Creates a new instance of a Pulumi stack, allowing you to interact with it.

*   `name` (string): The full name of the Pulumi stack (e.g., "my-org/my-project/dev").
*   `options_table` (Lua table): A table of options to configure the stack:
    *   `workdir` (string): **Required.** The path to the root directory of the Pulumi project associated with this stack.

**Returns:**
*   `PulumiStack` (userdata): An instance of the `PulumiStack` object for the specified stack.

### `PulumiStack` Object Methods

All methods below are called on the `PulumiStack` instance (e.g., `my_stack:up(...)`).

#### `stack:up(options)`

Executes the `pulumi up` command to create or update the stack's resources.

*   `options` (Lua table, optional): A table of options for the `up` command:
    *   `non_interactive` (boolean): If `true`, adds the `--non-interactive` and `--yes` flags to the `pulumi up` command.
    *   `config` (Lua table): A table of key-value pairs to pass configurations to the stack (e.g., `["my-app:vpcId"] = vpc_id`).
    *   `args` (Lua table of strings): A list of additional arguments to be passed directly to the `pulumi up` command.

**Returns:**
*   `result` (Lua table): A table containing:
    *   `success` (boolean): `true` if the operation was successful, `false` otherwise.
    *   `stdout` (string): The standard output of the Pulumi command.
    *   `stderr` (string): The standard error output of the Pulumi command.
    *   `error` (string or `nil`): A Go error message if the command execution failed.

#### `stack:preview(options)`

Executes the `pulumi preview` command to show a preview of the changes that would be applied.

*   `options` (Lua table, optional): The same options as for `stack:up()`.

**Returns:**
*   `result` (Lua table): The same return format as `stack:up()`.

#### `stack:refresh(options)`

Executes the `pulumi refresh` command to update the stack's state with the real resources in the cloud.

*   `options` (Lua table, optional): The same options as for `stack:up()`.

**Returns:**
*   `result` (Lua table): The same return format as `stack:up()`.

#### `stack:destroy(options)`

Executes the `pulumi destroy` command to destroy all resources in the stack.

*   `options` (Lua table, optional): The same options as for `stack:up()`.

**Retorna:**
*   `result` (Lua table): The same return format as `stack:up()`.

#### `stack:outputs()`

Gets the outputs of the Pulumi stack.

**Returns:**
*   `outputs` (Lua table): A Lua table where keys are output names and values are the respective stack outputs.
*   `error` (string or `nil`): An error message if the operation fails.

## Usage Examples

### Basic Pulumi Orchestration Example

This example demonstrates how to deploy two Pulumi stacks, passing an output from the first as an input to the second.

```lua
-- examples/pulumi_example.lua

command = function()
    log.info("Starting Pulumi orchestration example...")

    -- Example 1: Deploy a base stack (e.g., VPC)
    log.info("Deploying the base infrastructure stack (VPC)...")
    local vpc_stack = pulumi.stack("my-org/vpc-network/prod", {
        workdir = "./pulumi/vpc" -- Assuming the Pulumi project directory is here
    })

    -- Execute 'pulumi up' non-interactively
    local vpc_result = vpc_stack:up({ non_interactive = true })

    -- Check the VPC deployment result
    if not vpc_result.success then
        log.error("VPC stack deployment failed: " .. vpc_result.stderr)
        return false, "VPC deployment failed."
    end
    log.info("VPC stack deployed successfully. Stdout: " .. vpc_result.stdout)

    -- Get outputs from the VPC stack
    local vpc_outputs, outputs_err = vpc_stack:outputs()
    if outputs_err then
        log.error("Failed to get VPC stack outputs: " .. outputs_err)
        return false, "Failed to get VPC outputs."
    end

    local vpc_id = vpc_outputs.vpcId -- Assuming the stack exports 'vpcId'
    if not vpc_id then
        log.warn("VPC stack did not export 'vpcId'. Continuing without it.")
        vpc_id = "unknown-vpc-id"
    end
    log.info("Obtained VPC ID from outputs: " .. vpc_id)

    -- Example 2: Deploy an application stack, using outputs from the previous stack as config
    log.info("Deploying the application stack into VPC: " .. vpc_id)
    local app_stack = pulumi.stack("my-org/app-server/prod", {
        workdir = "./pulumi/app" -- Assuming the app's Pulumi project directory is here
    })

    -- Execute 'pulumi up' passing outputs from the previous stack as configuration
    local app_result = app_stack:up({
        non_interactive = true,
        config = {
            ["my-app:vpcId"] = vpc_id,
            ["aws:region"] = "us-east-1"
        }
    })

    -- Check the application deployment result
    if not app_result.success then
        log.error("Application stack deployment failed: " .. app_result.stderr)
        return false, "Application deployment failed."
    end
    log.info("Application stack deployed successfully. Stdout: " .. app_result.stdout)

    log.info("Pulumi orchestration example finished successfully.")
    return true, "Pulumi orchestration example finished."
end

TaskDefinitions = {
    pulumi_orchestration_example = {
        description = "Demonstrates using the 'pulumi' module to orchestrate infrastructure stacks.",
        tasks = {
            {
                name = "run_pulumi_orchestration",
                command = command
            }
        }
    }
}
```

---
**Available Languages:**
[English](./pulumi.md) | [Português](../../pt/modules/pulumi.md) | [中文](../../zh/modules/pulumi.md)