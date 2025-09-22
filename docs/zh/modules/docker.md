# Docker 模块

`docker` 模块提供了一个方便的界面，用于与 Docker 守护进程交互，允许您在管道中构建、运行和推送 Docker 镜像。

## 配置

此模块需要安装 `docker` CLI，并且 Docker 守护进程正在运行且可访问。

## 函数

### `docker.exec(args)`

执行任何原始的 `docker` 命令。

- `args` (table): **必需。** 要传递给 `docker` 命令的参数列表（例如 `{"ps", "-a"}`）。
- **返回:** 包含 `success`、`stdout`、`stderr` 和 `exit_code` 的结果表。

### `docker.build(params)`

使用 `docker build` 构建 Docker 镜像。

- `params` (table):
    - `tag` (string): **必需。** 镜像的标签（例如 `my-app:latest`）。
    - `path` (string): **必需。** 构建上下文路径。
    - `dockerfile` (string): **可选。** Dockerfile 的路径。
    - `build_args` (table): **可选。** 构建参数表（例如 `{VERSION = "1.0"}`）。
- **返回:** 结果表。

### `docker.push(params)`

使用 `docker push` 将 Docker 镜像推送到注册表。

- `params` (table):
    - `tag` (string): **必需。** 要推送的镜像的标签。
- **返回:** 结果表。

### `docker.run(params)`

使用 `docker run` 运行 Docker 容器。

- `params` (table):
    - `image` (string): **必需。** 要运行的镜像。
    - `name` (string): **可选。** 容器的名称。
    - `detach` (boolean): **可选。** 如果为 `true`，则在后台运行容器 (`-d`)。
    - `ports` (table): **可选。** 端口映射列表（例如 `{"8080:80"}`）。
    - `env` (table): **可选。** 环境变量表（例如 `{MY_VAR = "value"}`）。
- **返回:** 结果表。

## 示例

```lua
local image_tag = "my-test-image:latest"

-- 任务 1: Build
local result_build = docker.build({
  tag = image_tag,
  path = "./app"
})
if not result_build.success then return false, "构建失败" end

-- 任务 2: Run
local result_run = docker.run({
  image = image_tag,
  name = "my-test-container",
  ports = {"8080:80"}
})
if not result_run.success then return false, "运行失败" end

-- 任务 3: Push (测试成功后)
local result_push = docker.push({tag = image_tag})
if not result_push.success then return false, "推送失败" end
```
