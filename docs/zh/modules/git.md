# Git 模块

Sloth-Runner 中的 `git` 模块提供了一个流畅、高级的 API，可直接从您的 Lua 脚本与 Git 存储库进行交互。这使您能够自动化常见的 Git 操作，如克隆、拉取、添加、提交、打标签和推送，从而促进 CI/CD 工作流和版本控制自动化。

## 常见用例

*   **CI/CD 自动化:** 克隆存储库、更新代码、提交脚本生成的更改，并推送到版本控制系统。
*   **配置管理:** 在应用更改之前，从 Git 存储库中拉取最新的配置。
*   **自动化版本控制:** 为新的软件版本创建标签和提交。

## API 参考

### `git.clone(url, path)`

将 Git 存储库从 URL 克隆到本地路径。如果该路径已包含 Git 存储库，则该函数将返回 `nil` 和一条错误消息。

*   `url` (字符串): 要克隆的 Git 存储库的 URL。
*   `path` (字符串): 将在其中克隆存储库的本地路径。

**返回:**
*   `GitRepo` (用户数据): 如果克隆成功，则为 `GitRepo` 对象的实例。
*   `error` (字符串): 如果克隆失败或路径已经是存储库，则为错误消息。

### `git.repo(path)`

打开对现有本地 Git 存储库的引用。

*   `path` (字符串): Git 存储库根目录的本地路径。

**返回:**
*   `GitRepo` (用户数据): 如果路径是有效的 Git 存储库，则为 `GitRepo` 对象的实例。
*   `error` (字符串): 如果路径不是 Git 存储库，则为错误消息。

### `GitRepo` 对象方法 (可链式调用)

以下所有方法都在 `GitRepo` 实例上调用 (例如 `repo:checkout(...)`)，并返回 `GitRepo` 实例本身以允许方法链式调用。要获取上次操作的结果，请使用 `:result()` 方法。

#### `repo:checkout(ref)`

更改存储库的当前分支或提交。

*   `ref` (字符串): 要检出的分支、标签或提交哈希。

#### `repo:pull(remote, branch)`

从远程存储库拉取最新的更改。

*   `remote` (字符串): 远程的名称 (例如 "origin")。
*   `branch` (字符串): 要拉取的分支的名称。

#### `repo:add(pattern)`

将文件添加到 Git 索引 (暂存区)。

*   `pattern` (字符串): 要添加的文件模式 (例如 "."、"path/to/file.txt")。

#### `repo:commit(message)`

使用索引中的更改创建一个新的提交。

*   `message` (字符串): 提交消息。

#### `repo:tag(name, message)`

在存储库中创建一个新标签。

*   `name` (字符串): 标签名称 (例如 "v1.0.0")。
*   `message` (字符串, 可选): 标签的可选消息。

#### `repo:push(remote, branch, options)`

将提交和标签推送到远程存储库。

*   `remote` (字符串): 远程的名称 (例如 "origin")。
*   `branch` (字符串): 要推送的分支的名称。
*   `options` (Lua 表, 可选): 用于附加标志的选项表：
    *   `follow_tags` (布尔值): 如果为 `true`，则将 `--follow-tags` 标志添加到 `git push` 命令。

#### `repo:result()`

返回在 `GitRepo` 实例上执行的最后一个 Git 操作的结果。

**返回:**
*   `result` (Lua 表): 一个包含以下内容的表：
    *   `success` (布尔值): 如果操作成功，则为 `true`；否则为 `false`。
    *   `stdout` (字符串): Git 命令的标准输出。
    *   `stderr` (字符串): Git 命令的标准错误输出。
    *   `error` (字符串或 `nil`): 如果命令执行失败，则为 Go 错误消息。

## 用法示例

### 基本 Git 自动化示例

此示例演示如何克隆存储库、拉取更改、模拟修改、提交和推送更改。

```lua
-- examples/git_example.lua

command = function(params)
    log.info("正在开始 Git 自动化示例...")

    local repo_url = "https://github.com/chalkan3/sloth-runner.git" -- 以 sloth-runner 本身作为示例
    local repo_path = "./sloth-runner-checkout"
    local new_version = params.version or "v1.0.0-test" -- 示例版本
    local repo

    -- 如果存储库尚不存在于本地，则克隆它
    if not fs.exists(repo_path) then
        log.info("正在克隆存储库: " .. repo_url .. " 到 " .. repo_path)
        local cloned_repo, clone_err = git.clone(repo_url, repo_path)
        if clone_err then
            log.error("克隆存储库失败: " .. clone_err)
            return false, "Git 克隆失败。"
        end
        repo = cloned_repo
    else
        log.info("存储库已存在，正在打开本地引用: " .. repo_path)
        local opened_repo, open_err = git.repo(repo_path) -- 只获取本地仓库的对象
        if open_err then
            log.error("打开存储库失败: " .. open_err)
            return false, "Git 仓库打开失败。"
        end
        repo = opened_repo
    end

    if not repo then
        return false, "克隆或打开存储库失败。"
    end

    log.info("正在 " .. repo.RepoPath .. " 上开始 git 操作...")

    -- 流畅地链式执行一系列命令
    -- 注意: 每个操作都返回 'repo' 对象以进行链式调用。
    -- 要检查每个步骤的成功情况，您应该在每个步骤之后调用 :result()，
    -- 或者在链的末尾调用以获取最后一个命令的结果。

    log.info("正在检出 main 分支并拉取最新更改...")
    repo:checkout("main"):pull("origin", "main")
    local pull_result = repo:result() -- 获取最后一个命令 (pull) 的结果
    if not pull_result.success then
        log.error("检出或拉取失败: " .. pull_result.stderr)
        return false, "Git 检出/拉取失败。"
    end
    log.info("检出和拉取成功。Stdout: " .. pull_result.stdout)

    -- 模拟存储库中的更改
    local version_file_path = repo_path .. "/VERSION_EXAMPLE" -- 使用不同的名称以避免冲突
    fs.write(version_file_path, new_version)
    log.info("已将 VERSION_EXAMPLE 文件更新为: " .. new_version)

    -- 以链式方式添加、提交、打标签和推送更改
    local commit_message = "ci: 示例将版本升级到 " .. new_version
    log.info("正在添加、提交、打标签和推送更改...")

    -- 链式调用: add -> commit -> tag -> push
    repo:add(version_file_path)
        :commit(commit_message)
        :tag(new_version, "发布 " .. new_version)
        :push("origin", "main", { follow_tags = true })

    local final_push_result = repo:result() -- 获取最后一个命令 (push) 的结果

    -- 检查链中最后一个操作的结果
    if not final_push_result.success then
        log.error("推送更改失败: " .. final_push_result.stderr)
        return false, "Git 推送失败。"
    end

    log.info("已成功将版本 " .. new_version .. " 推送到 origin。Stdout: " .. final_push_result.stdout)
    log.info("Git 自动化示例成功完成。")
    return true, "Git 自动化示例已完成。"
end

TaskDefinitions = {
    git_automation_example = {
        description = "演示使用 'git' 模块进行存储库自动化。",
        tasks = {
            {
                name = "run_git_automation",
                command = command,
                params = {
                    version = "v1.0.0-test" -- 示例参数
                }
            }
        }
    }
}
```

---
[English](../../en/modules/git.md) | [Português](../../pt/modules/git.md) | [中文](./git.md)