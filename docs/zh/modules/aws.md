# AWS 模块

`aws` 模块提供了一个全面的界面，用于使用 AWS CLI 与亚马逊网络服务进行交互。它旨在与标准的 AWS 凭证链无缝协作，并为 `aws-vault` 提供一流的支持以增强安全性。

## 配置

`values.yaml` 中无需特定配置。该模块依赖于您的环境配置为可与 AWS 交互。这可以通过以下方式实现：
- EC2 实例或 ECS/EKS 任务的 IAM 角色。
- 标准环境变量（`AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY` 等）。
- 已配置的 `~/.aws/credentials` 文件。
- 使用带有命名配置文件的 `aws-vault`。

## 通用执行器

### `aws.exec(args, opts)`

这是该模块的核心功能。它执行任何 AWS CLI 命令并返回结果。

**参数:**

- `args` (table): **必需。** 一个字符串表，表示要传递给 AWS CLI 的命令和参数（例如 `{"s3", "ls", "--recursive"}`）。
- `opts` (table): **可选。** 一个执行选项表。
    - `profile` (string): 如果提供，将使用 `aws-vault exec <profile> -- aws ...` 执行命令。如果省略，将直接运行 `aws ...`。

**返回:**

一个包含以下字段的表：
- `stdout` (string): 命令的标准输出。
- `stderr` (string): 命令的标准错误。
- `exit_code` (number): 命令的退出代码。`0` 通常表示成功。

**示例:**

```lua
-- 使用默认凭证
local result = aws.exec({"sts", "get-caller-identity"})
if result.exit_code == 0 then
  print(result.stdout)
end

-- 使用 aws-vault 配置文件
local result_with_profile = aws.exec({"ec2", "describe-instances"}, {profile = "my-prod-profile"})
```

## S3 辅助函数

### `aws.s3.sync(params)`

`aws s3 sync` 命令的高级包装器，用于将目录与 S3 同步。

**参数:**

- `params` (table): 一个包含以下字段的表：
    - `source` (string): **必需。** 源目录或 S3 路径。
    - `destination` (string): **必需。** 目标目录或 S3 路径。
    - `profile` (string): **可选。** 要使用的 `aws-vault` 配置文件。
    - `delete` (boolean): **可选。** 如果为 `true`，则向同步命令添加 `--delete` 标志。

**返回:**

- 成功时返回 `true`。
- 失败时返回 `false, error_message`。

**示例:**

```lua
local ok, err = aws.s3.sync({
  source = "./build",
  destination = "s3://my-app-bucket/static",
  profile = "deployment-profile",
  delete = true
})
if not ok then
  log.error("S3 同步失败: " .. err)
end
```

## Secrets Manager 辅助函数

### `aws.secretsmanager.get_secret(params)`

从 AWS Secrets Manager 检索密钥的值。此函数通过直接返回 `SecretString` 来简化该过程。

**参数:**

- `params` (table): 一个包含以下字段的表：
    - `secret_id` (string): **必需。** 要检索的密钥的名称或 ARN。
    - `profile` (string): **可选。** 要使用的 `aws-vault` 配置文件。

**返回:**

- 成功时返回 `secret_string` (string)。
- 失败时返回 `nil, error_message`。

**示例:**

```lua
local db_password, err = aws.secretsmanager.get_secret({
  secret_id = "production/database/password",
  profile = "my-app-profile"
})

if not db_password then
  log.error("获取密钥失败: " .. err)
  return false, "配置失败。"
end

-- 现在您可以使用 db_password 变量
```
