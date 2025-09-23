-- examples/comprehensive_git_showcase.lua
--
-- This file provides a comprehensive showcase of the 'git' module's capabilities
-- and the recommended patterns for git-based workflows in sloth-runner.

TaskDefinitions = {
  ["comprehensive-git-showcase"] = {
    description = "Demonstrates cloning, inspecting, and interacting with a Git repository.",
    -- We keep the workdir on failure for easier debugging.
    clean_workdir_after_run = function(last_result)
      if not last_result.success then
        log.error("A task failed. The workdir will be kept for debugging at: " .. last_result.output.workdir)
      end
      return last_result.success
    end,
    tasks = {
      {
        name = "clone_repository",
        description = "Clones a public repository using the built-in git.clone function.",
        command = function(params)
          log.info("--- Task: " .. params.task_name .. " ---")
          local workdir = params.workdir
          local repo_url = "https://github.com/chalkan3/sloth-runner.git"
          local clone_path = workdir .. "/sloth-runner-clone"

          log.info("Cloning '" .. repo_url .. "' into '" .. clone_path .. "'...")

          local git = require("git")
          -- git.clone executes 'git clone <url> <path>' and returns a result table.
          local result = git.clone(repo_url, clone_path)

          if not result.success then
            log.error("Failed to clone repository: " .. result.stderr)
            return false, "Git clone failed."
          end

          log.info("Repository cloned successfully.")
          -- The output of this task makes the cloned path available to other tasks.
          return true, "Repository cloned.", { repo_path = clone_path }
        end
      },
      {
        name = "inspect_with_exec",
        description = "Inspects the cloned repo using the 'exec' module (Recommended).",
        depends_on = "clone_repository",
        command = function(params, inputs)
          log.info("--- Task: " .. params.task_name .. " ---")
          local repo_path = inputs.clone_repository.repo_path
          if not repo_path then return false, "Repo path not received from dependency." end

          log.info("Getting the last 3 commit logs from '" .. repo_path .. "'...")

          -- The recommended way to run git commands is with the 'exec' module.
          -- The '-C' flag tells git to run the command in the specified directory.
          local log_cmd = "git -C " .. repo_path .. " log -n 3 --oneline"
          local stdout, stderr, err = exec.run(log_cmd)

          if err then
            log.error("Failed to get git log: " .. stderr)
            return false, "Could not get git log."
          end

          log.info("Last 3 commits:\n" .. stdout)
          return true, "Repository inspected successfully."
        end
      },
      {
        name = "checkout_branch_with_exec",
        description = "Checks out a new branch using the 'exec' module.",
        depends_on = "inspect_with_exec",
        command = function(params, inputs)
          log.info("--- Task: " .. params.task_name .. " ---")
          local repo_path = inputs.clone_repository.repo_path
          if not repo_path then return false, "Repo path not received from dependency." end

          log.info("Checking out a new branch 'feature/my-new-branch'...")
          local checkout_cmd = "git -C " .. repo_path .. " checkout -b feature/my-new-branch"
          local _, stderr, err = exec.run(checkout_cmd)
          if err then
            log.error("Failed to checkout new branch: " .. stderr)
            return false, "Branch checkout failed."
          end

          log.info("Successfully checked out new branch.")

          log.info("Verifying current branch...")
          local branch_cmd = "git -C " .. repo_path .. " rev-parse --abbrev-ref HEAD"
          local stdout, stderr, err = exec.run(branch_cmd)
          if err then
            log.error("Failed to verify current branch: " .. stderr)
            return false, "Branch verification failed."
          end

          log.info("Current branch is: " .. stdout)
          return true, "Branch checkout verified."
        end
      },
      {
        name = "show_repo_object_placeholders",
        description = "Demonstrates the 'git.repo' object (NOTE: methods are placeholders).",
        depends_on = "clone_repository",
        command = function(params, inputs)
          log.info("--- Task: " .. params.task_name .. " ---")
          local repo_path = inputs.clone_repository.repo_path
          if not repo_path then return false, "Repo path not received from dependency." end

          local git = require("git")
          -- git.repo() creates a Lua table with info about the repository.
          local repo = git.repo(repo_path)

          log.info("Repo object created. Accessing properties:")
          log.info("  - Local Path: " .. repo.local.path)
          log.info("  - Remote URL: " .. repo.remote.url)
          log.info("  - Current Branch: " .. repo.current.branch)

          log.warn("NOTE: The following methods (pull, commit, push) are placeholders and do NOT perform real git operations. Use exec.run for these actions.")
          -- These calls will only print messages to the console.
          repo:pull()
          repo:commit("This is a test commit message.")
          repo:push()

          return true, "Repo object demonstration complete."
        end
      }
    }
  }
}
