package types

import (
	lua "github.com/yuin/gopher-lua"
)

type Task struct {
	Name        string
	Description string
	CommandFunc *lua.LFunction // Stores the Lua function if command is dynamic
	CommandStr  string         // Stores the command string if static
	Params      map[string]string
	PreExec     *lua.LFunction // Stores the Lua function for pre-execution hook
	PostExec    *lua.LFunction // Stores the Lua function for post-execution hook
	Async       bool           // Whether the task should run asynchronously
	DependsOn   []string       // Names of the tasks this one depends on
	Output      *lua.LTable    // Stores the output of the task after execution
}

type TaskGroup struct {
	Description string
	Tasks       []Task
}