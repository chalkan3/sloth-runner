# Net Module

The `net` module provides functions for making HTTP requests and downloading files, allowing your tasks to interact with web services and remote resources.

---

## `net.http_get(url)`

Performs an HTTP GET request to the specified URL.

*   **Parameters:**
    *   `url` (string): The URL to send the GET request to.
*   **Returns:**
    *   `body` (string): The response body as a string.
    *   `status_code` (number): The HTTP status code of the response.
    *   `headers` (table): A table containing the response headers.
    *   `error` (string): An error message if the request failed.

---

## `net.http_post(url, body, [headers])`

Performs an HTTP POST request to the specified URL.

*   **Parameters:**
    *   `url` (string): The URL to send the POST request to.
    *   `body` (string): The request body to send.
    *   `headers` (table, optional): A table of request headers to set.
*   **Returns:**
    *   `body` (string): The response body as a string.
    *   `status_code` (number): The HTTP status code of the response.
    *   `headers` (table): A table containing the response headers.
    *   `error` (string): An error message if the request failed.

---

## `net.download(url, destination_path)`

Downloads a file from a URL and saves it to a local path.

*   **Parameters:**
    *   `url` (string): The URL of the file to download.
    *   `destination_path` (string): The local file path to save the downloaded content.
*   **Returns:**
    *   `error`: An error object if the download fails.

### Example

```lua
command = function()
  local net = require("net")
  
  -- Example GET request
  log.info("Performing GET request to httpbin.org...")
  local body, status, headers, err = net.http_get("https://httpbin.org/get")
  if err then
    log.error("GET request failed: " .. err)
    return false, "GET request failed"
  end
  log.info("GET request successful! Status: " .. status)
  -- print("Response Body: " .. body)

  -- Example POST request
  log.info("Performing POST request to httpbin.org...")
  local post_body = '{"name": "sloth-runner", "awesome": true}'
  local post_headers = { ["Content-Type"] = "application/json" }
  body, status, headers, err = net.http_post("https://httpbin.org/post", post_body, post_headers)
  if err then
    log.error("POST request failed: " .. err)
    return false, "POST request failed"
  end
  log.info("POST request successful! Status: " .. status)
  -- print("Response Body: " .. body)

  -- Example Download
  local download_path = "/tmp/sloth-runner-logo.svg"
  log.info("Downloading file to " .. download_path)
  local err = net.download("https://raw.githubusercontent.com/chalkan3/sloth-runner/master/assets/sloth-runner-logo.svg", download_path)
  if err then
    log.error("Download failed: " .. err)
    return false, "Download failed"
  end
  log.info("File downloaded successfully.")
  fs.rm(download_path) -- Clean up

  return true, "Net module operations successful."
end
```
