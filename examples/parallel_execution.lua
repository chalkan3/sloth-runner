-- examples/parallel_execution.lua

TaskDefinitions = {
  parallel_demo = {
    description = "A demo of the parallel execution feature.",
    tasks = {
      run_in_parallel = {
        name = "run_in_parallel",
        description = "This task will run several sub-tasks in parallel.",
        command = function()
          log.info("--- Starting parallel execution demo ---")

          -- Define some sub-tasks to run
          local sub_task_1 = {
            name = "Short sleep",
            command = "echo 'Sub-task 1 starting...'; sleep 2; echo 'Sub-task 1 finished.'"
          }

          local sub_task_2 = {
            name = "Medium sleep",
            command = "echo 'Sub-task 2 starting...'; sleep 4; echo 'Sub-task 2 finished.'"
          }

          local sub_task_3 = {
            name = "Long sleep",
            command = "echo 'Sub-task 3 starting...'; sleep 6; echo 'Sub-task 3 finished.'"
          }
          
          local start_time = os.time()
          log.info("Calling parallel() with 3 sub-tasks (2s, 4s, 6s)...")

          -- Use the new parallel function
          local results, err = parallel({sub_task_1, sub_task_2, sub_task_3})

          local end_time = os.time()
          log.info("Parallel execution finished in " .. (end_time - start_time) .. " seconds.")
          log.info("Expected time is ~6 seconds (the duration of the longest task).")

          if err then
            log.error("Parallel execution failed: " .. err)
            return false, "Parallel block failed"
          end

          log.info("--- Results from parallel execution ---")
          for i, result in ipairs(results) do
            log.info("Task: " .. result.name .. ", Status: " .. result.status .. ", Duration: " .. result.duration)
			if result.error then
				log.error("  Error: " .. result.error)
			end
          end

          return true, "Parallel execution demo completed successfully."
        end
      }
    }
  }
}