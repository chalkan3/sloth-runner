# Terraform 模块

`terraform` 模块提供了一个高级界面，用于编排 `terraform` CLI 命令，允许您直接在 Sloth-Runner 管道内管理您的基础架构生命周期。

## 配置

此模块需要安装 `terraform` CLI 并可在系统的 PATH 中使用。所有命令都必须在您的 `.tf` 文件所在的特定 `workdir` 中执行。

## 函数

### `terraform.init(params)`

初始化 Terraform 工作目录。

- `params` (table):
    - `workdir` (string): **必需。** 包含 Terraform 文件的目录的路径。
- **返回:** 包含 `success`、`stdout`、`stderr` 和 `exit_code` 的结果表。

### `terraform.plan(params)`

创建 Terraform 执行计划。

- `params` (table):
    - `workdir` (string): **必需。** 目录的路径。
    - `out` (string): **可选。** 用于保存生成的计划的文件名。
- **返回:** 结果表。

### `terraform.apply(params)`

应用 Terraform 计划。

- `params` (table):
    - `workdir` (string): **必需。** 目录的路径。
    - `plan` (string): **可选。** 要应用的计划文件的路径。
    - `auto_approve` (boolean): **可选。** 如果为 `true`，则无需交互式批准即可应用更改。
- **返回:** 结果表。

### `terraform.destroy(params)`

销毁 Terraform 管理的基础架构。

- `params` (table):
    - `workdir` (string): **必需。** 目录的路径。
    - `auto_approve` (boolean): **可选。** 如果为 `true`，则无需交互式批准即可销毁资源。
- **返回:** 结果表。

### `terraform.output(params)`

从 Terraform 状态文件读取输出变量。

- `params` (table):
    - `workdir` (string): **必需。** 目录的路径。
    - `name` (string): **可选。** 要读取的特定输出的名称。如果省略，则所有输出都作为表返回。
- **返回:**
    - 成功时: 输出的已解析 JSON 值（可以是字符串、表等）。
    - 失败时: `nil, error_message`。

## 完整生命周期示例

```lua
local tf_workdir = "./examples/terraform"

-- 任务 1: Init
local result_init = terraform.init({workdir = tf_workdir})
if not result_init.success then return false, "Init 失败" end

-- 任务 2: Plan
local result_plan = terraform.plan({workdir = tf_workdir})
if not result_plan.success then return false, "Plan 失败" end

-- 任务 3: Apply
local result_apply = terraform.apply({workdir = tf_workdir, auto_approve = true})
if not result_apply.success then return false, "Apply 失败" end

-- 任务 4: Get Output
local filename, err = terraform.output({workdir = tf_workdir, name = "report_filename"})
if not filename then return false, "Output 失败: " .. err end
log.info("Terraform 创建的文件: " .. filename)

-- 任务 5: Destroy
local result_destroy = terraform.destroy({workdir = tf_workdir, auto_approve = true})
if not result_destroy.success then return false, "Destroy 失败" end
```
