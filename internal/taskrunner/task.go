package taskrunner

import (
	lua "github.com/yuin/gopher-lua"
)

type TaskGroup struct {
	Description string
	Tasks       []Task
}

// Global map to store task groups defined by Lua
// This will be populated by the luainterface package
var TaskGroups = make(map[string]TaskGroup)
