# 高级功能

本文档介绍 `sloth-runner` 的一些更高级的功能，旨在增强您的开发、调试和配置工作流。

## 交互式任务运行器

对于复杂的工作流，逐个执行任务、检查其输出并决定是继续、跳过还是重试任务可能很有用。交互式任务运行器为调试和开发任务管道提供了一种强大的方法。

要使用交互式运行器，请将 `--interactive` 标志添加到 `sloth-runner run` 命令中：

```bash
sloth-runner run -f examples/basic_pipeline.lua --yes --interactive
```

启用后，运行器将在执行每个任务之前暂停并提示您执行操作：

```
? 任务: fetch_data (模拟获取原始数据)
> 运行
  跳过
  中止
  继续
```

**操作:**

*   **运行:** (默认) 继续执行当前任务。
*   **跳过:** 跳过当前任务并转到执行顺序中的下一个任务。
*   **中止:** 立即中止整个任务执行。
*   **继续:** 执行当前任务和所有后续任务，不再提示，从而有效地为余下的运行禁用交互模式。

## 增强的 `values.yaml` 模板

您可以通过使用 Go 模板语法注入环境变量来使 `values.yaml` 文件更加动态。这对于提供敏感信息（如令牌或密钥）或特定于环境的配置特别有用，而无需对其进行硬编码。

`sloth-runner` 将 `values.yaml` 作为 Go 模板处理，使任何环境变量都可以在 `.Env` 映射下使用。

**示例:**

1.  **创建一个带有模板占位符的 `values.yaml` 文件：**

    ```yaml
    # values.yaml
    api_key: "{{ .Env.MY_API_KEY }}"
    region: "{{ .Env.AWS_REGION | default "us-east-1" }}"
    ```
    *注意：如果未设置环境变量，您可以使用 `default` 提供备用值。*

2.  **创建一个使用这些值的 Lua 任务：**

    ```lua
    -- my_task.lua
    TaskDefinitions = {
      my_group = {
        tasks = {
          {
            name = "deploy",
            command = function()
              log.info("部署到区域: " .. values.region)
              log.info("使用 API 密钥 (前 5 个字符): " .. string.sub(values.api_key, 1, 5) .. "...")
              return true, "部署成功。"
            end
          }
        }
      }
    }
    ```

3.  **在设置环境变量的情况下运行任务：**

    ```bash
    export MY_API_KEY="supersecretkey12345"
    export AWS_REGION="us-west-2"

    sloth-runner run -f my_task.lua -v values.yaml --yes
    ```

**输出:**

输出将显示环境变量中的值已正确替换：

```
INFO 部署到区域: us-west-2
INFO 使用 API 密钥 (前 5 个字符): super...
```
