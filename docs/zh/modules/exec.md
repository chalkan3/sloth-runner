# Exec 模块

`exec` 模块是 `sloth-runner` 中最基本的模块之一。它提供了一个强大的函数来执行任意的 shell 命令，让您可以完全控制执行环境。

## `exec.run(command, [options])`

使用 `bash -c` 执行一个 shell 命令。

### 参数

*   `command` (string): 要执行的 shell 命令。
*   `options` (table, 可选): 用于控制执行的选项表。
    *   `workdir` (string): 命令应在其中执行的工作目录。如果未提供，它将在任务组的临时目录（如果可用）或当前目录中运行。
    *   `env` (table): 为命令执行设置的环境变量字典（键值对）。这些变量会添加到现有环境中。

### 返回

一个包含命令执行结果的表：

*   `success` (boolean): 如果命令以代码 `0` 退出，则为 `true`，否则为 `false`。
*   `stdout` (string): 命令的标准输出。
*   `stderr` (string): 命令的标准错误输出。

### 示例

此示例演示如何使用带有自定义工作目录和环境变量的 `exec.run`。

```lua
-- examples/exec_module_example.lua

TaskDefinitions = {
  main = {
    description = "一个演示 exec 模块的任务。",
    tasks = {
      {
        name = "run-with-options",
        description = "使用自定义工作目录和环境执行命令。",
        command = function()
          log.info("准备运行自定义命令...")
          
          local exec = require("exec")
          
          -- 为示例创建一个临时目录
          local temp_dir = "/tmp/sloth-exec-test"
          fs.mkdir(temp_dir)
          fs.write(temp_dir .. "/test.txt", "来自测试文件的问候")

          -- 定义选项
          local options = {
            workdir = temp_dir,
            env = {
              MY_VAR = "SlothRunner",
              ANOTHER_VAR = "is_awesome"
            }
          }

          -- 执行命令
          local result = exec.run("echo 'MY_VAR is $MY_VAR' && ls -l && cat test.txt", options)

          -- 清理临时目录
          fs.rm_r(temp_dir)

          if result.success then
            log.info("命令成功执行！")
            print("--- STDOUT ---")
            print(result.stdout)
            print("--------------")
            return true, "Exec 命令成功。"
          else
            log.error("Exec 命令失败。")
            log.error("Stderr: " .. result.stderr)
            return false, "Exec 命令失败。"
          end
        end
      }
    }
  }
}
```
