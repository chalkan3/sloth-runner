--
-- artifacts_example.lua
--
-- This example demonstrates the artifact management feature.
-- One task creates a file, declares it as an artifact, and a second
-- task consumes that artifact to read its content.
--
-- The pipeline will:
-- 1. Task `generate-report`: Creates a `report.txt` file in its own workdir
--    and declares `artifacts = {"report.txt"}`.
-- 2. The runner saves `report.txt` to a shared artifact storage.
-- 3. Task `publish-report`: Declares `consumes = {"report.txt"}`.
-- 4. The runner copies `report.txt` from shared storage into this task's workdir.
-- 5. The `publish-report` task reads the file and prints its content.
--
-- To run this example:
--    go run ./cmd/sloth-runner -f examples/artifacts_example.lua
--

local log = require("log")

TaskDefinitions = {
  ["artifacts-pipeline"] = {
    description = "A pipeline to demonstrate producing and consuming artifacts.",
    create_workdir_before_run = true,

    tasks = {
      {
        name = "generate-report",
        description = "Creates a report file and declares it as an artifact.",
        artifacts = {"report.txt", "another-file-*.log"},
        command = function(params)
          local report_content = "This is the content of the report."
          fs.write(params.workdir .. "/report.txt", report_content)
          fs.write(params.workdir .. "/another-file-123.log", "some log data")
          log.info("Generated report.txt in workdir: " .. params.workdir)
          return true, "Report generated."
        end
      },
      {
        name = "publish-report",
        description = "Consumes the report artifact and reads its content.",
        depends_on = "generate-report",
        consumes = {"report.txt"},
        command = function(params)
          local report_path = params.workdir .. "/report.txt"
          log.info("Attempting to read consumed artifact at: " .. report_path)

          if not fs.exists(report_path) then
            log.error("Consumed artifact 'report.txt' was not found in the workdir!")
            return false, "Artifact not found."
          end

          local content, err = fs.read(report_path)
          if err then
            log.error("Failed to read artifact file: " .. err)
            return false, "Failed to read artifact."
          end

          log.info("Successfully read content from consumed artifact:")
          print("--- Artifact Content ---")
          print(content)
          print("------------------------")

          return true, "Artifact published."
        end
      }
    }
  }
}
