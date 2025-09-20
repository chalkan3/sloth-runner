-- examples/basic_pipeline_test.lua

test.describe("Basic Pipeline Tests", function()
  test.it(function()
    local result = test.run_task("fetch_data")
    assert.is_true(result.success, "fetch_data should run successfully")
  end)

  test.it(function()
    local result = test.run_task("process_data")
    assert.is_true(result.success, "process_data task should succeed")
    
    local expected_output = "processed_some_data_from_api"
    local actual_output = result.output and result.output.final_data
    
    assert.equals(actual_output, expected_output, "process_data should produce the correct output")
  end)

  test.it(function()
    assert.equals("hello", "world", "this assertion is designed to fail")
  end)
end)
