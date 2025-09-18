package types

import (
	"time"
	lua "github.com/yuin/gopher-lua"
)

// TaskResult holds the outcome of a single task execution.
type TaskResult struct {
	Name     string
	Status   string
	Duration time.Duration
	Error    error
}

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
	Retries     int            // Number of times to retry a failed task
	Timeout     string         // Timeout for the task (e.g., "30s", "1m")
	RunIf       string         // A shell command that must succeed for the task to run
	AbortIf     string         // A shell command that, if it succeeds, will abort the entire execution
	RunIfFunc   *lua.LFunction // A Lua function that must return true for the task to run
	AbortIfFunc *lua.LFunction // A Lua function that, if it returns true, will abort the entire execution
	NextIfFail  []string       // Names of the tasks that must fail for this one to run
}

type TaskGroup struct {
	Description string
	Tasks       []Task
}

type TaskRunner interface {
	RunTasksParallel(tasks []*Task, input *lua.LTable) ([]TaskResult, error)
}