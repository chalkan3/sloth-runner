-- examples/templated_values_task.lua

TaskDefinitions = {
  templated_values_group = {
    description = "Demonstrates templating in values.yaml.",
    tasks = {
      {
        name = "print_templated_value",
        description = "Prints a value from values.yaml that was templated with an environment variable.",
        command = function()
          log.info("Templated value: " .. values.my_value)
          return true, "Printed templated value."
        end
      }
    }
  }
}