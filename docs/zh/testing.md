# 测试工作流

sloth-runner 包含一个内置的测试框架，允许您为任务工作流编写单元和集成测试。为您的自动化编写测试对于确保可靠性、防止回归以及在进行更改时充满信心至关重要。

## `test` 命令

您可以使用 `sloth-runner test` 命令运行测试文件。它需要两个主要文件：您要测试的工作流和测试脚本本身。

```bash
sloth-runner test -w <工作流路径.lua> -f <测试文件路径.lua>
```

-   `-w, --workflow`: 指定要测试的主 `TaskDefinitions` 文件的路径。
-   `-f, --file`: 指定您的测试文件的路径。

## 编写测试

测试是用 Lua 编写的，并使用测试运行器提供的两个新的全局模块：`test` 和 `assert`。

### `test` 模块

`test` 模块用于构建您的测试并运行工作流中的特定任务。

-   `test.describe(suite_name, function)`: 将相关测试分组到一个“套件”中。这用于组织。
-   `test.it(function)`: 定义单个测试用例。测试的描述应包含在此函数内的断言消息中。
-   `test.run_task(task_name)`: 这是测试框架的核心功能。它从加载的工作流文件中按名称执行单个任务。它返回一个包含执行详细信息的 `result` 表。

`run_task` 返回的 `result` 表具有以下结构：

```lua
{
  success = true, -- 布尔值：如果任务成功则为 true，否则为 false
  message = "任务执行成功", -- 字符串：任务返回的消息
  duration = "1.23ms", -- 字符串：执行持续时间
  output = { ... }, -- 表：任务返回的输出表
  error = nil -- 字符串：如果任务失败，则为错误消息
}
```

### `assert` 模块

`assert` 模块提供用于检查任务执行结果的函数。

-   `assert.is_true(value, message)`: 检查 `value` 是否为 true。
-   `assert.equals(actual, expected, message)`: 检查 `actual` 值是否等于 `expected` 值。

### 模块模拟 (Mocking)

为了测试您的管道逻辑而无需进行实际的外部调用（例如，对 AWS、Docker 或 Terraform），测试框架包含了一个强大的模拟功能。

#### 严格模拟策略

测试运行器强制执行 **严格的模拟策略**。在测试模式下运行时，任何对模块函数（如 `aws.exec` 或 `docker.build`）的调用如果 **没有** 被明确模拟，将导致测试立即失败。这可确保您的测试是完全自包含的、确定性的，并且没有意外的副作用。

#### `test.mock(function_name, mock_definition)`

此函数允许您为任何可模拟的模块函数定义一个伪造的返回值。

-   `function_name` (string): 要模拟的函数的全名（例如 `"aws.s3.sync"`, `"docker.build"`）。
-   `mock_definition` (table): 一个定义模拟函数应返回什么的表。它 **必须** 包含一个 `returns` 键，该键是一个函数将返回的值的列表。

`returns` 列表至关重要，因为 Lua 函数可以返回多个值。

**示例:**

```lua
-- 模拟一个返回单个结果表的函数
test.mock("docker.build", {
  returns = {
    { success = true, stdout = "成功构建镜像" }
  }
})

-- 模拟一个返回两个值的函数（例如，一个值和一个错误）
-- 这模拟了对 terraform.output 的成功调用
test.mock("terraform.output", {
  returns = { "my_file.txt", nil }
})

-- 这模拟了失败的调用
test.mock("terraform.output", {
  returns = { nil, "未找到输出" }
})
```

### 完整的模拟示例

假设您有一个调用 `aws.exec` 的任务，并且其逻辑取决于输出。

**`my_workflow.lua` 中的任务:**
```lua
-- ...
{
  name = "check-account",
  command = function()
    local result = aws.exec({"sts", "get-caller-identity"})
    local data = data.parse_json(result.stdout)
    if data.Account == "123456789012" then
      return true, "正确的帐户。"
    else
      return false, "错误的帐户。"
    end
  end
}
-- ...
```

**`my_test.lua` 中的测试:**
```lua
test.describe("帐户检查逻辑", function()
  test.it(function()
    -- 模拟 aws.exec 的返回值
    test.mock("aws.exec", {
      returns = {
        {
          success = true,
          stdout = '{"Account": "123456789012"}'
        }
      }
    })

    -- 运行使用模拟的任务
    local result = test.run_task("check-account")

    -- 断言任务的逻辑在模拟数据下是否正常工作
    assert.is_true(result.success, "使用正确的帐户 ID，任务应该成功")
    assert.equals(result.message, "正确的帐户。", "消息应该是正确的")
  end)
end)
```
