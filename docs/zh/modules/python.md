# Python 模块

`python` 模块提供了一种方便的方式来管理 Python 虚拟环境 (`venv`) 并从您的 `sloth-runner` 任务中执行脚本。这对于涉及基于 Python 的工具或脚本的工作流特别有用。

---

## `python.venv(path)`

创建一个 Python 虚拟环境对象。请注意，这只在 Lua 中创建对象；环境本身在文件系统上直到您调用 `:create()` 后才被创建。

*   **参数:**
    *   `path` (string): 应在其中创建虚拟环境的文件系统路径 (例如, `./.venv`)。
*   **返回:**
    *   `venv` (object): 一个虚拟环境对象，包含与其交互的方法。

---

### `venv:create()`

在指定路径的文件系统上创建虚拟环境。

*   **返回:**
    *   `error`: 如果创建失败，则返回一个错误对象。

---

### `venv:pip(command)`

在虚拟环境的上下文中执行 `pip` 命令。

*   **参数:**
    *   `command` (string): 要传递给 `pip` 的参数 (例如, `install -r requirements.txt`)。
*   **返回:**
    *   `result` (table): 一个包含 `pip` 命令的 `stdout`、`stderr` 和 `exit_code` 的表。

---

### `venv:exec(script_path)`

使用虚拟环境中的 Python 解释器执行 Python 脚本。

*   **参数:**
    *   `script_path` (string): 要执行的 Python 脚本的路径。
*   **返回:**
    *   `result` (table): 一个包含脚本执行的 `stdout`、`stderr` 和 `exit_code` 的表。

### 示例

此示例演示了一个完整的生命周期：创建虚拟环境、从 `requirements.txt` 文件安装依赖项以及运行 Python 脚本。

```lua
-- examples/python_venv_lifecycle_example.lua

TaskDefinitions = {
  main = {
    description = "一个演示 Python venv 生命周期的任务。",
    create_workdir_before_run = true, -- 使用临时工作目录
    tasks = {
      {
        name = "run-python-script",
        description = "创建 venv，安装依赖项并运行脚本。",
        command = function(params) 
          local python = require("python")
          local workdir = params.workdir -- 从组中获取临时工作目录
          
          -- 1. 将我们的 Python 脚本和依赖项写入工作目录
          fs.write(workdir .. "/requirements.txt", "requests==2.28.1")
          fs.write(workdir .. "/main.py", "import requests\nprint(f'Hello from Python! Using requests version: {requests.__version__}')")

          -- 2. 创建一个 venv 对象
          local venv_path = workdir .. "/.venv"
          log.info("正在设置虚拟环境于: " .. venv_path)
          local venv = python.venv(venv_path)

          -- 3. 在文件系统上创建 venv
          venv:create()

          -- 4. 使用 pip 安装依赖项
          log.info("正在从 requirements.txt 安装依赖项...")
          local pip_result = venv:pip("install -r " .. workdir .. "/requirements.txt")
          if pip_result.exit_code ~= 0 then
            log.error("Pip 安装失败: " .. pip_result.stderr)
            return false, "未能安装 Python 依赖项。"
          end

          -- 5. 执行脚本
          log.info("正在运行 Python 脚本...")
          local exec_result = venv:exec(workdir .. "/main.py")
          if exec_result.exit_code ~= 0 then
            log.error("Python 脚本失败: " .. exec_result.stderr)
            return false, "Python 脚本执行失败。"
          end

          log.info("Python 脚本成功执行。")
          print("--- Python 脚本输出 ---")
          print(exec_result.stdout)
          print("----------------------------")

          return true, "Python venv 生命周期完成。"
        end
      }
    }
  }
}
```

```