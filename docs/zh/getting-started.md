# 快速入门

欢迎使用 Sloth-Runner！本指南将帮助您快速开始使用该工具。

## 安装

要在您的系统上安装 `sloth-runner`，您可以使用提供的 `install.sh` 脚本。此脚本会自动检测您的操作系统和架构，从 GitHub 下载最新版本，并将 `sloth-runner` 可执行文件放置在 `/usr/local/bin` 中。

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**注意：** `install.sh` 脚本需要 `sudo` 权限才能将可执行文件移动到 `/usr/local/bin`。

## 基本用法

要运行 Lua 任务文件：

```bash
sloth-runner run -f examples/basic_pipeline.lua
```

要列出文件中的任务：

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

## 下一步

现在您已经安装并运行了 Sloth-Runner，请探索[核心概念](./core-concepts.md)以了解如何定义任务，或者直接深入了解新的[内置模块](../index.md#内置模块)以使用 Git、Pulumi 和 Salt 进行高级自动化。

---
**可用语言：**
[English](../en/getting-started.md) | [Português](../pt/getting-started.md) | [中文](./getting-started.md)