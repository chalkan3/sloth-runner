package types

import (
	"io"
	"os/exec"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Agent represents a remote agent that can execute tasks.
type Agent struct {
	Address string `yaml:"address"`
}

// Task represents a single unit of work in the runner.
type Task struct {
	Name        string
	Description string
	CommandFunc *lua.LFunction
	CommandStr  string
	DependsOn   []string
	Artifacts   []string
	Consumes    []string
	NextIfFail  []string
	Params      map[string]string
	Retries     int
	Timeout     string
	Async       bool
	PreExec     *lua.LFunction
	PostExec    *lua.LFunction
	RunIf       string
	RunIfFunc   *lua.LFunction
	AbortIf     string
	AbortIfFunc *lua.LFunction
	Output      *lua.LTable
	Agent       string // New field
}

// TaskGroup represents a collection of related tasks.
type TaskGroup struct {
	Description              string
	Tasks                    []Task
	Workdir                  string
	CreateWorkdirBeforeRun   bool
	CleanWorkdirAfterRunFunc *lua.LFunction
	Agents                   map[string]Agent `yaml:"agents"` // New field
}

// TaskResult holds the outcome of a single task execution.
type TaskResult struct {
	Name     string
	Status   string
	Duration time.Duration
	Error    error
}

// SharedSession holds data that can be shared between tasks in a group.
type SharedSession struct {
	Workdir string
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Stderr  io.ReadCloser
}

// TaskRunner is the interface for the main task execution engine.
type TaskRunner interface {
	Run() error
	RunTasksParallel(tasks []*Task, input *lua.LTable) ([]TaskResult, error)
}

// Exporter defines an interface for exporting data from a Lua script.
type Exporter interface {
	Export(data map[string]interface{})
}

// PythonVenv represents a Python virtual environment.
type PythonVenv struct {
	Path string
}
