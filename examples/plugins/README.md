# Plugin Examples

This directory contains example plugins for `sloth-runner`.

## `hello-plugin`

This is a simple plugin that provides a `greet` function.

### Installation

To install this plugin, you can use the `plugin install` command with the path to the plugin directory:

```bash
sloth-runner plugin install examples/plugins/hello-plugin
```

### Usage

Once installed, you can use the plugin in your tasks:

```lua
local hello = require("hello-plugin")

TaskDefinitions = {
  my_group = {
    tasks = {
      {
        name = "test_plugin",
        command = function()
          hello.greet("world")
          return true
        end
      }
    }
  }
}
```
