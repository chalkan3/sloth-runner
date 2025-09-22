# Testing Workflows

The sloth-runner includes a built-in testing framework that allows you to write unit and integration tests for your task workflows. Writing tests for your automation is crucial for ensuring reliability, preventing regressions, and having confidence when making changes.

## The `test` Command

You can run a test file using the `sloth-runner test` command. It requires two main files: the workflow you want to test and the test script itself.

```bash
sloth-runner test -w <path_to_workflow.lua> -f <path_to_test_file.lua>
```

-   `-w, --workflow`: Specifies the path to the main `TaskDefinitions` file that you want to test.
-   `-f, --file`: Specifies the path to your test file.

## Writing Tests

Tests are written in Lua and use two new global modules provided by the test runner: `test` and `assert`.

### The `test` Module

The `test` module is used to structure your tests and to run specific tasks from your workflow.

-   `test.describe(suite_name, function)`: Groups related tests into a "suite". This is for organization.
-   `test.it(function)`: Defines an individual test case. The description of the test should be included in the assertion messages inside this function.
-   `test.run_task(task_name)`: This is the core function of the testing framework. It executes a single task by its name from the loaded workflow file. It returns a `result` table containing the execution details.

The `result` table returned by `run_task` has the following structure:

```lua
{
  success = true, -- boolean: true if the task succeeded, false otherwise
  message = "Task executed successfully", -- string: The message returned by the task
  duration = "1.23ms", -- string: The execution duration
  output = { ... }, -- table: The output table returned by the task
  error = nil -- string: The error message if the task failed
}
```

### The `assert` Module

The `assert` module provides functions to check the results of your task executions.

-   `assert.is_true(value, message)`: Checks if the `value` is true.
-   `assert.equals(actual, expected, message)`: Checks if the `actual` value is equal to the `expected` value.

### Mocking Modules

To test the logic of your pipelines without making real external calls (e.g., to AWS, Docker, or Terraform), the testing framework includes a powerful mocking feature.

#### Strict Mocking Policy

The test runner enforces a **strict mocking policy**. When running in test mode, any call to a module function (like `aws.exec` or `docker.build`) that has **not** been explicitly mocked will cause the test to fail immediately. This ensures that your tests are fully self-contained, deterministic, and do not have unintended side effects.

#### `test.mock(function_name, mock_definition)`

This function allows you to define a fake return value for any mockable module function.

-   `function_name` (string): The full name of the function to mock (e.g., `"aws.s3.sync"`, `"docker.build"`).
-   `mock_definition` (table): A table that defines what the mocked function should return. It **must** contain a `returns` key, which is a list of the values the function will return.

The `returns` list is crucial because Lua functions can return multiple values.

**Example:**

```lua
-- Mock a function that returns a single result table
test.mock("docker.build", {
  returns = {
    { success = true, stdout = "Successfully built image" }
  }
})

-- Mock a function that returns two values (e.g., a value and an error)
-- This simulates a successful call to terraform.output
test.mock("terraform.output", {
  returns = { "my_file.txt", nil }
})

-- This simulates a failed call
test.mock("terraform.output", {
  returns = { nil, "output not found" }
})
```

### Complete Mocking Example

Let's say you have a task that calls `aws.exec` and has logic that depends on the output.

**Task in `my_workflow.lua`:**
```lua
-- ...
{
  name = "check-account",
  command = function()
    local result = aws.exec({"sts", "get-caller-identity"})
    local data = data.parse_json(result.stdout)
    if data.Account == "123456789012" then
      return true, "Correct account."
    else
      return false, "Wrong account."
    end
  end
}
-- ...
```

**Test in `my_test.lua`:**
```lua
test.describe("Account Check Logic", function()
  test.it(function()
    -- Mock the return value of aws.exec
    test.mock("aws.exec", {
      returns = {
        {
          success = true,
          stdout = '{"Account": "123456789012"}'
        }
      }
    })

    -- Run the task that uses the mock
    local result = test.run_task("check-account")

    -- Assert that the task's logic worked correctly with the mocked data
    assert.is_true(result.success, "Task should succeed with the correct account ID")
    assert.equals(result.message, "Correct account.", "Message should be correct")
  end)
end)
```
