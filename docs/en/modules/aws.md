# AWS Module

The `aws` module provides a comprehensive interface for interacting with Amazon Web Services using the AWS CLI. It is designed to work seamlessly with standard AWS credential chains and also has first-class support for `aws-vault` for enhanced security.

## Configuration

No specific configuration in `values.yaml` is required. The module relies on your environment being configured to interact with AWS. This can be achieved through:
- IAM roles for EC2 instances or ECS/EKS tasks.
- Standard environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.).
- A configured `~/.aws/credentials` file.
- Using `aws-vault` with a named profile.

## Generic Executor

### `aws.exec(args, opts)`

This is the core function of the module. It executes any AWS CLI command and returns the result.

**Parameters:**

- `args` (table): **Required.** A table of strings representing the command and arguments to pass to the AWS CLI (e.g., `{"s3", "ls", "--recursive"}`).
- `opts` (table): **Optional.** A table of options for the execution.
    - `profile` (string): If provided, the command will be executed using `aws-vault exec <profile> -- aws ...`. If omitted, it will run `aws ...` directly.

**Returns:**

A table containing the following fields:
- `stdout` (string): The standard output from the command.
- `stderr` (string): The standard error from the command.
- `exit_code` (number): The exit code of the command. `0` typically indicates success.

**Example:**

```lua
-- Using default credentials
local result = aws.exec({"sts", "get-caller-identity"})
if result.exit_code == 0 then
  print(result.stdout)
end

-- Using an aws-vault profile
local result_with_profile = aws.exec({"ec2", "describe-instances"}, {profile = "my-prod-profile"})
```

## S3 Helpers

### `aws.s3.sync(params)`

A high-level wrapper for the `aws s3 sync` command, useful for synchronizing directories with S3.

**Parameters:**

- `params` (table): A table containing the following fields:
    - `source` (string): **Required.** The source directory or S3 path.
    - `destination` (string): **Required.** The destination directory or S3 path.
    - `profile` (string): **Optional.** The `aws-vault` profile to use.
    - `delete` (boolean): **Optional.** If `true`, adds the `--delete` flag to the sync command.

**Returns:**

- `true` on success.
- `false, error_message` on failure.

**Example:**

```lua
local ok, err = aws.s3.sync({
  source = "./build",
  destination = "s3://my-app-bucket/static",
  profile = "deployment-profile",
  delete = true
})
if not ok then
  log.error("S3 sync failed: " .. err)
end
```

## Secrets Manager Helpers

### `aws.secretsmanager.get_secret(params)`

Retrieves a secret's value from AWS Secrets Manager. This function simplifies the process by directly returning the `SecretString`.

**Parameters:**

- `params` (table): A table containing the following fields:
    - `secret_id` (string): **Required.** The name or ARN of the secret to retrieve.
    - `profile` (string): **Optional.** The `aws-vault` profile to use.

**Returns:**

- `secret_string` (string) on success.
- `nil, error_message` on failure.

**Example:**

```lua
local db_password, err = aws.secretsmanager.get_secret({
  secret_id = "production/database/password",
  profile = "my-app-profile"
})

if not db_password then
  log.error("Failed to get secret: " .. err)
  return false, "Config failed."
end

-- Now you can use the db_password variable
```
