# Azure Module

The `azure` module provides an interface for interacting with Microsoft Azure using the `az` command-line tool.

## Configuration

This module requires the `az` CLI to be installed and authenticated. Before running pipelines that use this module, you must log in to your Azure account:

```bash
az login
```

The module will use your logged-in credentials for all commands.

## Generic Executor

### `azure.exec(args)`

Executes any `az` command. This function automatically adds the `--output json` flag (if not already present) to ensure that the output is machine-parsable.

**Parameters:**

- `args` (table): **Required.** A table of strings representing the command and arguments to pass to `az` (e.g., `{"group", "list", "--location", "eastus"}`).

**Returns:**

A table containing the following fields:
- `stdout` (string): The standard output from the command (as a JSON string).
- `stderr` (string): The standard error from the command.
- `exit_code` (number): The exit code of the command. `0` typically indicates success.

**Example:**

```lua
local result = azure.exec({"account", "show"})
if result.exit_code == 0 then
  local account_info, err = data.parse_json(result.stdout)
  if account_info then
    log.info("Logged in as: " .. account_info.user.name)
  end
end
```

## Resource Group (RG) Helpers

### `azure.rg.delete(params)`

Deletes a resource group.

**Parameters:**

- `params` (table): A table containing the following fields:
    - `name` (string): **Required.** The name of the resource group to delete.
    - `yes` (boolean): **Optional.** If `true`, adds the `--yes` flag to bypass the confirmation prompt.

**Returns:**

- `true` on success.
- `false, error_message` on failure.

**Example:**

```lua
local ok, err = azure.rg.delete({
  name = "my-test-rg",
  yes = true
})
if not ok then
  log.error("Failed to delete resource group: " .. err)
end
```

## Virtual Machine (VM) Helpers

### `azure.vm.list(params)`

Lists virtual machines.

**Parameters:**

- `params` (table): **Optional.** A table containing the following fields:
    - `resource_group` (string): The name of a resource group to scope the list to. If omitted, lists VMs in the entire subscription.

**Returns:**

- `vms` (table) on success, where the table is a parsed JSON array of your VM objects.
- `nil, error_message` on failure.

**Example:**

```lua
-- List all VMs in the subscription
local all_vms, err1 = azure.vm.list()

-- List VMs in a specific resource group
local specific_vms, err2 = azure.vm.list({resource_group = "my-production-rg"})
if specific_vms then
  for _, vm in ipairs(specific_vms) do
    print("Found VM: " .. vm.name)
  end
end
```
