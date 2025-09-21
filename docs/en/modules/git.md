# Git Module

The `git` module provides a fluent API to interact with Git repositories, allowing you to automate common version control operations like cloning, committing, and pushing.

---

## `git.clone(url, path)`

Clones a Git repository to a local path.

*   **Parameters:**
    *   `url` (string): The URL of the repository to clone.
    *   `path` (string): The local directory to clone into.
*   **Returns:**
    *   `repo` (object): A `GitRepo` object on success.
    *   `error`: An error object if the clone fails.

---

## `git.repo(path)`

Opens an existing local Git repository.

*   **Parameters:**
    *   `path` (string): The path to the existing local repository.
*   **Returns:**
    *   `repo` (object): A `GitRepo` object on success.
    *   `error`: An error object if the path is not a valid Git repository.

---

## The `GitRepo` Object

This object represents a local repository and provides chainable methods for performing Git operations.

### `repo:checkout(ref)`

Checks out a specific branch, tag, or commit.

*   **Parameters:** `ref` (string).

### `repo:pull(remote, branch)`

Pulls changes from a remote.

*   **Parameters:** `remote` (string), `branch` (string).

### `repo:add(pattern)`

Stages files for a commit.

*   **Parameters:** `pattern` (string), e.g., `"."` or `"path/to/file.txt"`.

### `repo:commit(message)`

Creates a commit.

*   **Parameters:** `message` (string).

### `repo:tag(name, [message])`

Creates a new tag.

*   **Parameters:** `name` (string), `message` (string, optional).

### `repo:push(remote, branch, [options])`

Pushes commits to a remote.

*   **Parameters:**
    *   `remote` (string).
    *   `branch` (string).
    *   `options` (table, optional): e.g., `{ follow_tags = true }`.

### `repo:result()`

This method is called at the end of a chain to get the result of the *last* operation.

*   **Returns:**
    *   `result` (table): A table containing `success` (boolean), `stdout` (string), and `stderr` (string).

### Example

This example demonstrates a full CI/CD-like workflow: clone, create a version file, add, commit, tag, and push.

```lua
command = function()
  local git = require("git")
  local repo_path = "/tmp/git-example-repo"
  
  -- Clean up previous runs
  fs.rm_r(repo_path)

  -- 1. Clone the repository
  log.info("Cloning repository...")
  local repo, err = git.clone("https://github.com/chalkan3/sloth-runner.git", repo_path)
  if err then
    return false, "Failed to clone: " .. err
  end

  -- 2. Create and write a version file
  fs.write(repo_path .. "/VERSION", "1.2.3")

  -- 3. Chain Git commands: add -> commit -> tag -> push
  log.info("Adding, committing, tagging, and pushing...")
  repo:add("."):commit("ci: Bump version to 1.2.3"):tag("v1.2.3"):push("origin", "main", { follow_tags = true })

  -- 4. Get the result of the final operation (push)
  local result = repo:result()

  if not result.success then
    log.error("Git push failed: " .. result.stderr)
    return false, "Git push failed."
  end

  log.info("Successfully pushed new version tag.")
  return true, "Git operations successful."
end
```
