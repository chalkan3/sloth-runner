-- examples/export_example.lua

-- This example demonstrates how to use the new global 'export' function
-- to pass data from the script to the CLI when using the --return flag.

TaskDefinitions = {
  main = {
    description = "A task group to demonstrate the export function.",
    tasks = {
      {
        name = "export-data-task",
        description = "Exports a table and also returns a value.",
        command = function(params, inputs)
          log.info("Exporting some data...")

          -- Use the global export function to send a table to the runner
          export({
            exported_value = "this came from the export function",
            another_key = 12345,
            is_exported = true
          })

          log.info("Export complete. The task will now finish and return its own output.")

          -- The task's own return value will be merged with the exported data.
          -- If keys conflict, the exported value will win.
          return true, "Task finished successfully.", { task_return_value = "this came from the task's return" }
        end
      }
    }
  }
}
