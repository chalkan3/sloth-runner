--
-- mock_test.lua
--
-- This test file demonstrates how to use the mocking framework to test
-- the logic of a pipeline without executing real commands.
--

-- We need a dummy task definition to test against.
-- In a real scenario, you would load your actual pipeline file.
TaskDefinitions = {
  ["aws-pipeline"] = {
    tasks = {
      {
        name = "check-identity-and-decide",
        description = "A task whose logic depends on the output of aws.exec",
        command = function()
          local result = aws.exec({"sts", "get-caller-identity"})
          if not result.success then
            return false, "Failed to call AWS."
          end
          
          -- Logic: check if the AWS account is a specific one
          local data, err = data.parse_json(result.stdout)
          if err or not data.Account then
             return false, "Could not parse AWS response."
          end

          if data.Account == "123456789012" then
            return true, "Running in the correct account."
          else
            return false, "Running in the WRONG account."
          end
        end
      }
    }
  }
}


test.describe("AWS Pipeline Logic", function()
  test.it("should succeed if aws.exec returns the correct account ID", function()
    -- Mock the aws.exec function to return a successful result with the correct account
    test.mock("aws.exec", {
      returns = {
        { 
          success = true, 
          exit_code = 0,
          stdout = '{"Account": "123456789012", "UserId": "test", "Arn": "arn:test"}'
        }
      }
    })

    -- Run the task that calls the mocked function
    local result = test.run_task("check-identity-and-decide")

    -- Assert that the task's logic interpreted the mocked result correctly
    assert.is_true(result.success, "Task should succeed")
    assert.equals(result.message, "Running in the correct account.", "Success message should be correct")
  end)

  test.it("should fail if aws.exec returns a different account ID", function()
    -- Mock the aws.exec function to return a successful result but with the WRONG account
    test.mock("aws.exec", {
      returns = {
        { 
          success = true, 
          exit_code = 0,
          stdout = '{"Account": "999999999999", "UserId": "test", "Arn": "arn:test"}'
        }
      }
    })

    local result = test.run_task("check-identity-and-decide")

    assert.is_true(not result.success, "Task should fail")
    assert.equals(result.message, "Running in the WRONG account.", "Failure message should be correct")
  end)

  test.it("should fail if the aws.exec call itself fails", function()
    -- Mock the aws.exec function to return a failure
    test.mock("aws.exec", {
      returns = {
        { 
          success = false, 
          exit_code = 1,
          stderr = "AWS connection failed"
        }
      }
    })

    local result = test.run_task("check-identity-and-decide")

    assert.is_true(not result.success, "Task should fail")
    assert.equals(result.message, "Failed to call AWS.", "Failure message should indicate the AWS call failed")
  end)
end)
