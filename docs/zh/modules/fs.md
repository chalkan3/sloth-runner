# FS 模块

`fs` 模块提供了从您的 Lua 脚本直接与文件系统交互的基本功能。

---\n
## `fs.read(path)`

读取文件的全部内容。

*   **参数:**
    *   `path` (string): 文件路径。
*   **返回:**
    *   `string`: 文件内容。
    *   `error`: 如果读取失败，则返回一个错误对象。

---\n
## `fs.write(path, content)`

将内容写入文件，如果文件已存在则覆盖它。

*   **参数:**
    *   `path` (string): 文件路径。
    *   `content` (string): 要写入的内容。
*   **返回:**
    *   `error`: 如果写入失败，则返回一个错误对象。

---\n
## `fs.append(path, content)`

将内容追加到文件末尾。如果文件不存在，则创建它。

*   **参数:**
    *   `path` (string): 文件路径。
    *   `content` (string): 要追加的内容。
*   **返回:**
    *   `error`: 如果追加失败，则返回一个错误对象。

---\n
## `fs.exists(path)`

检查给定路径的文件或目录是否存在。

*   **参数:**
    *   `path` (string): 要检查的路径。
*   **返回:**
    *   `boolean`: 如果路径存在，则为 `true`，否则为 `false`。

---\n
## `fs.mkdir(path)`

在给定路径创建一个目录，包括任何必要的父目录 (类似于 `mkdir -p`)。

*   **参数:**
    *   `path` (string): 要创建的目录路径。
*   **返回:**
    *   `error`: 如果创建失败，则返回一个错误对象。

---\n
## `fs.rm(path)`

删除单个文件。

*   **参数:**
    *   `path` (string): 要删除的文件的路径。
*   **返回:**
    *   `error`: 如果删除失败，则返回一个错误对象。

---\n
## `fs.rm_r(path)`

递归地删除文件或目录 (类似于 `rm -rf`)。

*   **参数:**
    *   `path` (string): 要删除的路径。
*   **返回:**
    *   `error`: 如果删除失败，则返回一个错误对象。

---\n
## `fs.ls(path)`

列出目录的内容。

*   **参数:**
    *   `path` (string): 目录的路径。
*   **返回:**
    *   `table`: 包含文件和子目录名称的表。
    *   `error`: 如果列出失败，则返回一个错误对象。

---\n
## `fs.tmpname()`

生成一个唯一的临时目录路径。注意：此函数仅返回名称，不创建目录。

*   **返回:**
    *   `string`: 适合用作临时目录的唯一路径。
    *   `error`: 如果无法生成名称，则返回一个错误对象。

### 示例

```lua
command = function()
  local fs = require("fs")
  
  local tmp_dir = "/tmp/fs-example"
  log.info("正在创建目录: " .. tmp_dir)
  fs.mkdir(tmp_dir)

  local file_path = tmp_dir .. "/my_file.txt"
  log.info("正在写入文件: " .. file_path)
  fs.write(file_path, "你好, Sloth Runner!\n")

  log.info("正在追加到文件...")
  fs.append(file_path, "这是一个新行。")

  if fs.exists(file_path) then
    log.info("文件内容: " .. fs.read(file_path))
  end

  log.info("正在列出 " .. tmp_dir .. " 的内容")
  local contents = fs.ls(tmp_dir)
  for i, name in ipairs(contents) do
    print("- " .. name)
  end

  log.info("正在清理...")
  fs.rm_r(tmp_dir)
  
  return true, "FS 模块操作成功。"
end
```

```