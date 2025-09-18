# Getting Started

Welcome to Sloth-Runner! This guide will help you get started with the tool quickly.

## Installation

To install `sloth-runner` on your system, you can use the provided `install.sh` script. This script automatically detects your operating system and architecture, downloads the latest release from GitHub, and places the `sloth-runner` executable in `/usr/local/bin`.

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**Note:** The `install.sh` script requires `sudo` privileges to move the executable to `/usr/local/bin`.

## Basic Usage

To run a Lua task file:

```bash
sloth-runner run -f examples/basic_pipeline.lua
```

To list tasks in a file:

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

## Next Steps

Now that you have Sloth-Runner installed and running, explore the [Core Concepts](./core-concepts.md) to understand how to define your tasks, or dive directly into the new [Built-in Modules](./index.md#built-in-modules) for advanced automation with Git, Pulumi, and Salt.

---
[English](./getting-started.md) | [Português](../pt/getting-started.md) | [中文](../zh/getting-started.md)