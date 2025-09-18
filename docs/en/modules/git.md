# Git Module

The `git` module in Sloth-Runner provides a fluent, high-level API to interact with Git repositories directly from your Lua scripts. This allows you to automate common Git operations such as cloning, pulling, adding, committing, tagging, and pushing, facilitating CI/CD workflows and versioning automation.

## Common Use Cases

*   **CI/CD Automation:** Clone repositories, update code, commit script-generated changes, and push to version control.
*   **Configuration Management:** Pull the latest configurations from a Git repository before applying changes.
*   **Automated Versioning:** Create tags and commits for new software releases.

## API Reference

### `git.clone(url, path)`

Clones a Git repository from a URL to a local path. If the path already contains a Git repository, the function will return `nil` and an error message.

*   `url` (string): The URL of the Git repository to clone.
*   `path` (string): The local path where the repository will be cloned.

**Returns:**
*   `GitRepo` (userdata): An instance of the `GitRepo` object if the clone is successful.
*   `error` (string): An error message if the clone fails or the path is already a repository.

### `git.repo(path)`

Opens a reference to an existing local Git repository.

*   `path` (string): The local path to the root directory of the Git repository.

**Returns:**
*   `GitRepo` (userdata): An instance of the `GitRepo` object if the path is a valid Git repository.
*   `error` (string): An error message if the path is not a Git repository.

### `GitRepo` Object Methods (Chainable)

All methods below are called on the `GitRepo` instance (e.g., `repo:checkout(...)`) and return the `GitRepo` instance itself to allow method chaining. To get the result of the last operation, use the `:result()` method.

#### `repo:checkout(ref)`

Changes the current branch or commit of the repository.

*   `ref` (string): The branch, tag, or commit hash to checkout.

#### `repo:pull(remote, branch)`

Pulls the latest changes from a remote repository.

*   `remote` (string): The name of the remote (e.g., "origin").
*   `branch` (string): The name of the branch to pull.

#### `repo:add(pattern)`

Adds files to the Git index (staging area).

*   `pattern` (string): The file pattern to add (e.g., ".", "path/to/file.txt").

#### `repo:commit(message)`

Creates a new commit with the changes in the index.

*   `message` (string): The commit message.

#### `repo:tag(name, message)`

Creates a new tag in the repository.

*   `name` (string): The tag name (e.g., "v1.0.0").
*   `message` (string, optional): An optional message for the tag.

#### `repo:push(remote, branch, options)`

Pushes commits and tags to a remote repository.

*   `remote` (string): The name of the remote (e.g., "origin").
*   `branch` (string): The name of the branch to push.
*   `options` (Lua table, optional): A table of options for additional flags:
    *   `follow_tags` (boolean): If `true`, adds the `--follow-tags` flag to the `git push` command.

#### `repo:result()`

Returns the result of the last Git operation executed on the `GitRepo` instance.

**Returns:**
*   `result` (Lua table): A table containing:
    *   `success` (boolean): `true` if the operation was successful, `false` otherwise.
    *   `stdout` (string): The standard output of the Git command.
    *   `stderr` (string): The standard error output of the Git command.
    *   `error` (string or `nil`): A Go error message if the command execution failed.

## Usage Examples

### Basic Git Automation Example

This example demonstrates how to clone a repository, pull changes, simulate a modification, commit, and push the changes.

```lua
-- examples/git_example.lua

command = function(params)
    log.info("Starting Git automation example...")

    local repo_url = "https://github.com/chalkan3/sloth-runner.git" -- Using sloth-runner itself as an example
    local repo_path = "./sloth-runner-checkout"
    local new_version = params.version or "v1.0.0-test" -- Example version
    local repo

    -- Clone the repository if it doesn't exist locally yet
    if not fs.exists(repo_path) then
        log.info("Cloning repository: " .. repo_url .. " into " .. repo_path)
        local cloned_repo, clone_err = git.clone(repo_url, repo_path)
        if clone_err then
            log.error("Failed to clone repository: " .. clone_err)
            return false, "Git clone failed."
        end
        repo = cloned_repo
    else
        log.info("Repository already exists, opening local reference: " .. repo_path)
        local opened_repo, open_err = git.repo(repo_path) -- Just get the object for the local repo
        if open_err then
            log.error("Failed to open repository: " .. open_err)
            return false, "Git repo open failed."
        end
        repo = opened_repo
    end

    if not repo then
        return false, "Failed to clone or open repository."
    end

    log.info("Starting git operations on " .. repo.RepoPath .. "...")

    -- Execute a sequence of commands fluently and chained
    -- Note: Each operation returns the 'repo' object for chaining.
    -- To check the success of each step, you should call :result() after each one,
    -- or at the end of the chain for the last command.

    log.info("Checking out main branch and pulling latest changes...")
    repo:checkout("main"):pull("origin", "main")
    local pull_result = repo:result() -- Get the result of the last command (pull)
    if not pull_result.success then
        log.error("Failed to checkout or pull: " .. pull_result.stderr)
        return false, "Git checkout/pull failed."
    end
    log.info("Checkout and pull successful. Stdout: " .. pull_result.stdout)

    -- Simulate a change in the repository
    local version_file_path = repo_path .. "/VERSION_EXAMPLE" -- Use a different name to avoid conflict
    fs.write(version_file_path, new_version)
    log.info("Updated VERSION_EXAMPLE file to: " .. new_version)

    -- Add, commit, tag, and push changes in a chained manner
    local commit_message = "ci: Example bump version to " .. new_version
    log.info("Adding, committing, tagging, and pushing changes...")

    -- Chaining: add -> commit -> tag -> push
    repo:add(version_file_path)
        :commit(commit_message)
        :tag(new_version, "Release " .. new_version)
        :push("origin", "main", { follow_tags = true })

    local final_push_result = repo:result() -- Get the result of the last command (push)

    -- Check the result of the last operation in the chain
    if not final_push_result.success then
        log.error("Failed to push changes: " .. final_push_result.stderr)
        return false, "Git push failed."
    end

    log.info("Successfully pushed version " .. new_version .. " to origin. Stdout: " .. final_push_result.stdout)
    log.info("Git automation example finished successfully.")
    return true, "Git automation example finished."
end

TaskDefinitions = {
    git_automation_example = {
        description = "Demonstrates using the 'git' module for repository automation.",
        tasks = {
            {
                name = "run_git_automation",
                command = command,
                params = {
                    version = "v1.0.0-test" -- Example parameter
                }
            }
        }
    }
}
```

---
[English](./git.md) | [Português](../../pt/modules/git.md) | [中文](../../zh/modules/git.md)