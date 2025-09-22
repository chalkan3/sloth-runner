# DigitalOcean Module

The `digitalocean` module provides an interface for interacting with your DigitalOcean resources using the `doctl` command-line tool.

## Configuration

This module requires the `doctl` CLI to be installed and authenticated. The standard way to do this is to generate a personal access token in your DigitalOcean control panel and set it as the `DIGITALOCEAN_ACCESS_TOKEN` environment variable.

```bash
export DIGITALOCEAN_ACCESS_TOKEN="your_do_api_token_here"
```

The module will automatically use this token for all commands.

## Generic Executor

### `digitalocean.exec(args)`

Executes any `doctl` command. This function automatically adds the `--output json` flag to ensure that the output is machine-parsable.

**Parameters:**

- `args` (table): **Required.** A table of strings representing the command and arguments to pass to `doctl` (e.g., `{"compute", "droplet", "list"}`).

**Returns:**

A table containing the following fields:
- `stdout` (string): The standard output from the command (as a JSON string).
- `stderr` (string): The standard error from the command.
- `exit_code` (number): The exit code of the command. `0` typically indicates success.

**Example:**

```lua
local result = digitalocean.exec({"account", "get"})
if result.exit_code == 0 then
  local account_info, err = data.parse_json(result.stdout)
  if account_info then
    log.info("Account status: " .. account_info.status)
  end
end
```

## Droplets Helpers

### `digitalocean.droplets.list()`

A high-level wrapper to list all Droplets in your account.

**Returns:**

- `droplets` (table) on success, where the table is a parsed JSON array of your Droplet objects.
- `nil, error_message` on failure.

**Example:**

```lua
local droplets, err = digitalocean.droplets.list()
if droplets then
  for _, droplet in ipairs(droplets) do
    print("Found Droplet: " .. droplet.name)
  end
end
```

### `digitalocean.droplets.delete(params)`

Deletes a specific Droplet by its ID.

**Parameters:**

- `params` (table): A table containing the following fields:
    - `id` (string): **Required.** The ID of the Droplet to delete.
    - `force` (boolean): **Optional.** If `true`, adds the `--force` flag to bypass the confirmation prompt. Defaults to `false`.

**Returns:**

- `true` on success.
- `false, error_message` on failure.

**Example:**

```lua
local ok, err = digitalocean.droplets.delete({
  id = "123456789",
  force = true
})
if not ok then
  log.error("Failed to delete droplet: " .. err)
end
```
