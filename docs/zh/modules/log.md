# Log 模块

`log` 模块提供了一个简单而必要的接口，用于从您的 Lua 脚本中将消息记录到 `sloth-runner` 控制台。在任务执行期间，使用此模块是提供反馈和调试信息的标准方式。

---

## `log.info(message)`

以 INFO 级别记录一条消息。这是用于一般信息性消息的标准级别。

*   **参数:**
    *   `message` (string): 要记录的消息。

---

## `log.warn(message)`

以 WARN 级别记录一条消息。这适用于应引起用户注意的非关键问题。

*   **参数:**
    *   `message` (string): 要记录的消息。

---

## `log.error(message)`

以 ERROR 级别记录一条消息。这应用于可能导致任务失败的重大错误。

*   **参数:**
    *   `message` (string): 要记录的消息。

---

## `log.debug(message)`

以 DEBUG 级别记录一条消息。除非运行器处于详细或调试模式，否则这些消息通常是隐藏的。它们对于详细的诊断信息很有用。

*   **参数:**
    *   `message` (string): 要记录的消息。

### 示例

```lua
command = function()
  -- log 模块是全局可用的，不需要 require。
  
  log.info("启动日志记录示例任务。")
  
  local user_name = "Sloth"
  log.debug("当前用户是: " .. user_name)

  if user_name ~= "Sloth" then
    log.warn("用户不是预期的用户。")
  end

  log.info("任务正在执行其主要操作...")
  
  local success = true -- 模拟一次成功的操作
  if not success then
    log.error("主要操作意外失败！")
    return false, "主要操作失败"
  end

  log.info("日志记录示例任务成功完成。")
  return true, "日志记录已演示。"
end
```
