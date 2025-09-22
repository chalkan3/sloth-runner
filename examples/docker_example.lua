--
-- docker_example.lua
--
-- This example demonstrates how to use the `docker` module to build
-- and run a Docker image.
--
-- The pipeline will:
-- 1. Build a Docker image from the `examples/docker/Dockerfile`.
-- 2. Run a container from the new image to verify its output.
-- 3. Clean up by removing the container and the image.
--
-- To run this example:
--    go run ./cmd/sloth-runner -f examples/docker_example.lua
--

local log = require("log")

-- Configuration --
local image_tag = "sloth-runner-docker-example:latest"
local container_name = "sloth-runner-test-container"
-------------------

TaskDefinitions = {
  ["docker-build-pipeline"] = {
    description = "A pipeline to build and test a Docker image.",

    tasks = {
      {
        name = "build_image",
        description = "Builds the Docker image from the example Dockerfile.",
        command = function()
          log.info("Building Docker image with tag: " .. image_tag)
          local result = docker.build({
            tag = image_tag,
            path = "./examples/docker"
          })

          if not result.success then
            log.error("Docker build failed: " .. result.stderr)
            return false, "Docker build failed."
          end

          log.info("Docker image built successfully.")
          print(result.stdout)
          return true, "Image built."
        end
      },
      {
        name = "run_container_and_verify",
        description = "Runs a container to verify it works as expected.",
        depends_on = "build_image",
        command = function()
          log.info("Running container '" .. container_name .. "' from image " .. image_tag)
          local result = docker.run({
            image = image_tag,
            name = container_name
          })

          if not result.success then
            log.error("Failed to run container: " .. result.stderr)
            return false, "Container run failed."
          end

          log.info("Container ran successfully. Verifying output...")
          local expected_output = "Hello from a Docker container managed by Sloth-Runner!"
          
          -- Trim whitespace from stdout for a reliable comparison
          local actual_output = string.gsub(result.stdout, "^%s*(.-)%s*$", "%1")

          if actual_output ~= expected_output then
            log.error("Container output did not match expected output.")
            log.error("Expected: " .. expected_output)
            log.error("Actual: " .. actual_output)
            return false, "Container verification failed."
          end

          log.info("Container output verified successfully!")
          return true, "Container verified."
        end
      },
      {
        name = "cleanup_container",
        description = "Removes the test container.",
        depends_on = "run_container_and_verify",
        command = function()
          log.info("Removing container: " .. container_name)
          -- Using docker.exec for a simple command
          local result = docker.exec({"rm", container_name})
          if not result.success then
            log.error("Failed to remove container: " .. result.stderr)
            return false, "Container cleanup failed."
          end
          log.info("Container removed.")
          return true, "Container cleaned up."
        end
      },
      {
        name = "cleanup_image",
        description = "Removes the built Docker image.",
        depends_on = "cleanup_container",
        command = function()
          log.info("Removing image: " .. image_tag)
          local result = docker.exec({"rmi", image_tag})
          if not result.success then
            log.error("Failed to remove image: " .. result.stderr)
            return false, "Image cleanup failed."
          end
          log.info("Image removed.")
          return true, "Image cleaned up."
        end
      }
    }
  }
}
