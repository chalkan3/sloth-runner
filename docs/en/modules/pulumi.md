# Pulumi Module

The `pulumi` module provides a fluent API to orchestrate Pulumi stacks, enabling you to manage your Infrastructure as Code (IaC) workflows directly from `sloth-runner`.

---

## `pulumi.stack(name, options)`

Creates a Pulumi stack object.

*   **Parameters:**
    *   `name` (string): The full name of the stack (e.g., `"my-org/my-project/dev"`).
    *   `options` (table): A table of options.
        *   `workdir` (string): **Required.** The path to the Pulumi project directory.
*   **Returns:**
    *   `stack` (object): A `PulumiStack` object.
    *   `error`: An error object if the stack cannot be initialized.

---

## The `PulumiStack` Object

This object represents a specific Pulumi stack and provides methods for interaction.

### `stack:up([options])`

Creates or updates the stack's resources by running `pulumi up`.

*   **Parameters:**
    *   `options` (table, optional):
        *   `yes` (boolean): If `true`, passes `--yes` to approve the update automatically.
        *   `config` (table): A dictionary of configuration values to pass to the stack.
        *   `args` (table): A list of additional string arguments to pass to the command.
*   **Returns:**
    *   `result` (table): A table containing `success` (boolean), `stdout` (string), and `stderr` (string).

### `stack:preview([options])`

Previews the changes that would be made by an update by running `pulumi preview`.

*   **Parameters:** Same as `stack:up`.
*   **Returns:** Same as `stack:up`.

### `stack:refresh([options])`

Refreshes the stack's state by running `pulumi refresh`.

*   **Parameters:** Same as `stack:up`.
*   **Returns:** Same as `stack:up`.

### `stack:destroy([options])`

Destroys all resources in the stack by running `pulumi destroy`.

*   **Parameters:** Same as `stack:up`.
*   **Returns:** Same as `stack:up`.

### `stack:outputs()`

Retrieves the outputs of a deployed stack.

*   **Returns:**
    *   `outputs` (table): A Lua table of the stack's outputs.
    *   `error`: An error object if fetching outputs fails.

### Example

This example shows a common pattern: deploying a networking stack (VPC) and then using its output (`vpcId`) to configure and deploy an application stack.

```lua
command = function()
  local pulumi = require("pulumi")

  -- 1. Define the VPC stack
  local vpc_stack = pulumi.stack("my-org/vpc/prod", { workdir = "./pulumi/vpc" })
  
  -- 2. Deploy the VPC
  log.info("Deploying VPC stack...")
  local vpc_result = vpc_stack:up({ yes = true })
  if not vpc_result.success then
    return false, "VPC deployment failed: " .. vpc_result.stderr
  end

  -- 3. Get the VPC ID from its outputs
  log.info("Fetching VPC outputs...")
  local vpc_outputs, err = vpc_stack:outputs()
  if err then
    return false, "Failed to get VPC outputs: " .. err
  end
  local vpc_id = vpc_outputs.vpcId

  -- 4. Define the App stack
  local app_stack = pulumi.stack("my-org/app/prod", { workdir = "./pulumi/app" })

  -- 5. Deploy the App, passing the vpcId as configuration
  log.info("Deploying App stack into VPC: " .. vpc_id)
  local app_result = app_stack:up({
    yes = true,
    config = { ["my-app:vpcId"] = vpc_id }
  })
  if not app_result.success then
    return false, "App deployment failed: " .. app_result.stderr
  end

  log.info("All stacks deployed successfully.")
  return true, "Pulumi orchestration complete."
end
```
