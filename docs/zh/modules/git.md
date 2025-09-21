# Git 模块

`git` 模块提供了一个流畅的 API 来与 Git 存储库进行交互，允许您自动化常见的版本控制操作，如克隆、提交和推送。

---

## `git.clone(url, path)`

将 Git 存储库克隆到本地路径。

*   **参数:**
    *   `url` (string): 要克隆的存储库的 URL。
    *   `path` (string): 要克隆到的本地目录。
*   **返回:**
    *   `repo` (object): 成功时返回一个 `GitRepo` 对象。
    *   `error`: 如果克隆失败，则返回一个错误对象。

---

## `git.repo(path)`

打开一个现有的本地 Git 存储库。

*   **参数:**
    *   `path` (string): 现有本地存储库的路径。
*   **返回:**
    *   `repo` (object): 成功时返回一个 `GitRepo` 对象。
    *   `error`: 如果路径不是有效的 Git 存储库，则返回一个错误对象。

---

## `GitRepo` 对象

此对象表示一个本地存储库，并提供可链接的方法来执行 Git 操作。

### `repo:checkout(ref)`

检出特定的分支、标签或提交。

*   **参数:** `ref` (string)。

### `repo:pull(remote, branch)`

从远程拉取更改。

*   **参数:** `remote` (string), `branch` (string)。

### `repo:add(pattern)`

将文件暂存以进行提交。

*   **参数:** `pattern` (string), 例如 `"."` 或 `"path/to/file.txt"`。

### `repo:commit(message)`

创建一个提交。

*   **参数:** `message` (string)。

### `repo:tag(name, [message])`

创建一个新标签。

*   **参数:** `name` (string), `message` (string, 可选)。

### `repo:push(remote, branch, [options])`

将提交推送到远程。

*   **参数:**
    *   `remote` (string)。
    *   `branch` (string)。
    *   `options` (table, 可选): 例如 `{ follow_tags = true }`。

### `repo:result()`

此方法在链的末尾调用，以获取*最后一个*操作的结果。

*   **返回:**
    *   `result` (table): 一个包含 `success` (boolean)、`stdout` (string) 和 `stderr` (string) 的表。

### 示例

此示例演示了一个完整的类似 CI/CD 的工作流：克隆、创建版本文件、添加、提交、打标签和推送。

```lua
command = function()
  local git = require("git")
  local repo_path = "/tmp/git-example-repo"
  
  -- 清理以前的运行
  fs.rm_r(repo_path)

  -- 1. 克隆存储库
  log.info("正在克隆存储库...")
  local repo, err = git.clone("https://github.com/chalkan3/sloth-runner.git", repo_path)
  if err then
    return false, "克隆失败: " .. err
  end

  -- 2. 创建并写入版本文件
  fs.write(repo_path .. "/VERSION", "1.2.3")

  -- 3. 链接 Git 命令: add -> commit -> tag -> push
  log.info("正在添加、提交、打标签和推送...")
  repo:add("."):commit("ci: Bump version to 1.2.3"):tag("v1.2.3"):push("origin", "main", { follow_tags = true })

  -- 4. 获取最终操作 (push) 的结果
  local result = repo:result()

  if not result.success then
    log.error("Git 推送失败: " .. result.stderr)
    return false, "Git 推送失败。"
  end

  log.info("成功推送新版本标签。")
  return true, "Git 操作成功。"
end
```
