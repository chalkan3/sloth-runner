
-- testdata/retries_success.lua
-- Create a temporary file to track attempts
local attempt_file = "/tmp/sloth_runner_retry_test.txt"
fs.rm(attempt_file) -- Ensure it's clean before starting

function count_attempt()
    local count = 0
    if fs.exists(attempt_file) then
        count = tonumber(fs.read(attempt_file))
    end
    count = count + 1
    fs.write(attempt_file, tostring(count))
    return count
end

TaskDefinitions = {
    test_retries_success = {
        description = "Tests a task that succeeds after a retry",
        tasks = {
            {
                name = "flaky_task",
                retries = 2,
                command = function()
                    local attempt = count_attempt()
                    if attempt < 2 then
                        return false, "Simulating failure on attempt " .. attempt
                    end
                    return true, "Succeeded on attempt " .. attempt
                end
            }
        }
    }
}
