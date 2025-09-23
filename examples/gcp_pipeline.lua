TaskDefinitions = {
  ["deploy_gcp_hub_spoke_and_host"] = {
    description = "Clones GCP repositories for Hub/Spoke and Host Manager.",
    -- Corrected function name and logic
    clean_workdir_after_run = function (last_result)
      -- Only try to access output if the task failed, as it might not exist on success.
      if not last_result.success then
        -- Check if output and workdir exist before trying to access them.
        if last_result.output and last_result.output.workdir then
          log.error("Task failed. The workdir will not be deleted: " .. last_result.output.workdir)
        else
          log.error("Task failed and workdir info is unavailable. The workdir will be kept for debugging.")
        end
      end
      return last_result.success
    end,
    tasks = {
      {
        name = "git:clone:https://github.com/chalkan3/gcp-hub-spoke",
        description = "This task will clone https://github.com/chalkan3/gcp-hub-spoke",
        command = function(params)
          local workdir = params.workdir
          local repository_url = 'https://github.com/chalkan3/gcp-hub-spoke'
          local clone_path = workdir .. '/gcp-hub-spoke'
          local git = require("git")
          local clone_result = git.clone(repository_url, clone_path)

          -- Corrected property name
          if not clone_result.success then
            local err = "Unable to clone " .. repository_url .. " with stderr: " .. clone_result.stderr
            log.error(err)
            -- Return an output table even on failure
            return false, err, { workdir = workdir }
          end
          -- Return workdir on success as well
          return true, repository_url .. " cloned", { gcp_hub_spoke_path = clone_path, workdir = workdir }
        end
      },
      {
        name = "git:clone:https://github.com/chalkan3/gcp-host-manager",
        description = "This task will clone https://github.com/chalkan3/gcp-host-manager",
        command = function(params)
          local workdir = params.workdir
          local repository_url = 'https://github.com/chalkan3/gcp-host-manager'
          local clone_path = workdir .. '/gcp-host-manager'
          local git = require("git")
          local clone_result = git.clone(repository_url, clone_path)

          -- Corrected property name
          if not clone_result.success then
            local err = "Unable to clone " .. repository_url .. " with stderr: " .. clone_result.stderr
            log.error(err)
            -- Return an output table even on failure
            return false, err, { workdir = workdir }
          end
          -- Return workdir on success as well
          return true, repository_url .. " cloned", { gcp_host_manager_path = clone_path, workdir = workdir }
        end
      }
    }
  }
}