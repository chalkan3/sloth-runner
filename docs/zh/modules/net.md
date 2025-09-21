# Net 模块

`net` 模块提供了发出 HTTP 请求和下载文件的功能，允许您的任务与 Web 服务和远程资源进行交互。

---

## `net.http_get(url)`

向指定的 URL 执行 HTTP GET 请求。

*   **参数:**
    *   `url` (string): 要发送 GET 请求的 URL。
*   **返回:**
    *   `body` (string): 作为字符串的响应体。
    *   `status_code` (number): 响应的 HTTP 状态码。
    *   `headers` (table): 包含响应头的表。
    *   `error` (string): 如果请求失败，则为错误消息。

---

## `net.http_post(url, body, [headers])`

向指定的 URL 执行 HTTP POST 请求。

*   **参数:**
    *   `url` (string): 要发送 POST 请求的 URL。
    *   `body` (string): 要发送的请求体。
    *   `headers` (table, 可选): 要设置的请求头表。
*   **返回:**
    *   `body` (string): 作为字符串的响应体。
    *   `status_code` (number): 响应的 HTTP 状态码。
    *   `headers` (table): 包含响应头的表。
    *   `error` (string): 如果请求失败，则为错误消息。

---

## `net.download(url, destination_path)`

从 URL 下载文件并将其保存到本地路径。

*   **参数:**
    *   `url` (string): 要下载的文件的 URL。
    *   `destination_path` (string): 用于保存下载内容的本地文件路径。
*   **返回:**
    *   `error`: 如果下载失败，则返回一个错误对象。

### 示例

```lua
command = function()
  local net = require("net")
  
  -- GET 请求示例
  log.info("正在向 httpbin.org 执行 GET 请求...")
  local body, status, headers, err = net.http_get("https://httpbin.org/get")
  if err then
    log.error("GET 请求失败: " .. err)
    return false, "GET 请求失败"
  end
  log.info("GET 请求成功！状态: " .. status)
  -- print("响应体: " .. body)

  -- POST 请求示例
  log.info("正在向 httpbin.org 执行 POST 请求...")
  local post_body = '{"name": "sloth-runner", "awesome": true}'
  local post_headers = { ["Content-Type"] = "application/json" }
  body, status, headers, err = net.http_post("https://httpbin.org/post", post_body, post_headers)
  if err then
    log.error("POST 请求失败: " .. err)
    return false, "POST 请求失败"
  end
  log.info("POST 请求成功！状态: " .. status)
  -- print("响应体: " .. body)

  -- 下载示例
  local download_path = "/tmp/sloth-runner-logo.svg"
  log.info("正在下载文件到 " .. download_path)
  local err = net.download("https://raw.githubusercontent.com/chalkan3/sloth-runner/master/assets/sloth-runner-logo.svg", download_path)
  if err then
    log.error("下载失败: " .. err)
    return false, "下载失败"
  end
  log.info("文件下载成功。")
  fs.rm(download_path) -- 清理

  return true, "Net 模块操作成功。"
end
```
