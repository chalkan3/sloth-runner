-- examples/git_module_showcase.lua
--
-- This file demonstrates the capabilities of the 'git' module.
-- It shows how to clone a repository, including specific branches,
-- and how to structure a workflow that depends on the cloned code.

TaskDefinitions = {
  ["git-showcase"] = {
    description = "A group of tasks to demonstrate git module functionalities.",
    -- This setting ensures that if the pipeline fails, the working directory
    -- is kept, allowing you to inspect the cloned repository.
    clean_workdir_after_run = function(last_result)
      if not last_result.success then
        log.error("A task failed. The workdir will be kept for debugging at: " .. last_result.output.workdir)
      end
      return last_result.success
    end,
    tasks = {
      {
        name = "clone_public_repo",
        description = "Clones a public repository from GitHub.",
        command = function(params)
          -- Print the new context variables injected by the task runner
          log.info("--- Execution Context ---")
          log.info("Task Name: " .. params.task_name)
          log.info("Group Name: " .. params.group_name)
          log.info("-------------------------")

          local workdir = params.workdir
          local repo_url = "https://github.com/chalkan3/sloth-runner.git"
          local clone_path = workdir .. "/sloth-runner"

          log.info("Cloning repository '" .. repo_url .. "' into: " .. clone_path)

          local git = require("git")
          -- The clone function executes 'git clone <url> <path>'
          local result = git.clone(repo_url, clone_path)

          if not result.success then
            log.error("Failed to clone repository: " .. result.stderr)
            return false, "Git clone failed.", { workdir = workdir }
          end

          log.info("Repository cloned successfully.")
          log.info("Clone output (stdout): " .. result.stdout)

          -- For example, let's list the contents using exec.run
          local ls_stdout, ls_stderr, ls_err = exec.run("ls -l " .. clone_path)
          if ls_err then
            log.error("Failed to list repository contents: " .. ls_stderr)
            return false, "Failed to list repo contents."
          end
          log.info("Contents of the cloned repository:\n" .. ls_stdout)

          return true, "Repository cloned and verified.", { workdir = workdir, repo_path = clone_path }
        end
      },
      {
        name = "inspect_cloned_repo",
        description = "Inspects the repository that was cloned in the previous step.",
        depends_on = "clone_public_repo",
        command = function(params, inputs)
          -- 'inputs' contains the outputs from the tasks this one depends on.
          local repo_path = inputs.clone_public_repo.repo_path

          if not repo_path or not fs.exists(repo_path) then
            log.error("Cloned repository path not found or does not exist: " .. tostring(repo_path))
            return false, "Repository path is invalid."
          end

          log.info("Inspecting repository at: " .. repo_path)

          -- Get the git log using exec.run
          local log_cmd = "git -C " .. repo_path .. " log -n 3 --oneline"
          log.info("Running command: " .. log_cmd)
          local stdout, stderr, err = exec.run(log_cmd)

          if err then
            log.error("Failed to get git log: " .. stderr)
            return false, "Could not get git log."
          end

          log.info("Last 3 commits:\n" .. stdout)

          return true, "Repository inspected successfully."
        end
      }
    }
  }
}
