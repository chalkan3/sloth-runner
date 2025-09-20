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

### Example

Here is a complete example of a test file (`examples/basic_pipeline_test.lua`) that tests the `examples/basic_pipeline.lua` workflow.

```lua
-- examples/basic_pipeline_test.lua

test.describe("Basic Pipeline Tests", function()
  test.it(function()
    local result = test.run_task("fetch_data")
    assert.is_true(result.success, "fetch_data should run successfully")
  end)

  test.it(function()
    local result = test.run_task("process_data")
    assert.is_true(result.success, "process_data task should succeed")
    
    -- Note: This test is simplified. In a real scenario, you might want to
    -- mock the input from the 'fetch_data' dependency.
    local expected_output = "processed_some_data_from_api"
    local actual_output = result.output and result.output.final_data
    
    assert.equals(actual_output, expected_output, "process_data should produce the correct output")
  end)

  test.it(function()
    assert.equals("hello", "world", "this assertion is designed to fail")
  end)
end)
```

When you run this test, you will get a clear report in your terminal indicating which assertions passed and which failed, along with a final summary.
