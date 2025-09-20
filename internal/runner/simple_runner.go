// Package runner provides a simplified, isolated task execution mechanism.
// It is designed to run a single task without handling the full dependency graph
// or lifecycle hooks. Its primary use is for the testing framework, allowing
// individual tasks to be executed and their outputs to be asserted. This package
// was created to break an import cycle between the 'luainterface' and 'taskrunner'
// packages.
package runner

import (
	"context"
	"time"

	"github.com/chalkan3/sloth-runner/internal/lua/utils"
	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
)

// RunSingleTask executes a single task in a controlled, isolated Lua state.
// It is designed for testing purposes and does not handle task dependencies,
// hooks, or workdir management. It creates a new Lua state for the task,
// executes its command function, and then copies the output back to the original
// Lua state provided.
//
// Parameters:
//   L: The original Lua state, used as the destination for the output table.
//   task: The task object to be executed.
//
// Returns:
//   - bool: True if the task's command function returned true.
//   - string: The message returned by the task.
//   - *lua.LTable: The output table from the task, copied to the original state L.
//   - time.Duration: The total execution time of the task.
//   - error: Any error that occurred during the execution of the Lua function.
func RunSingleTask(L *lua.LState, task *types.Task) (bool, string, *lua.LTable, time.Duration, error) {
	startTime := time.Now()

	// We need a new Lua state to isolate the execution
	taskL := lua.NewState()
	defer taskL.Close()
	// Note: We cannot call luainterface.OpenAll(taskL) here as it would
	// re-introduce the import cycle. The test runner assumes that the task's
	// command function is self-contained or uses modules that are opened
	// selectively. For now, this limitation is acceptable for unit testing.

	success, msg, output, err := utils.ExecuteLuaFunction(
		taskL,
		task.CommandFunc,
		task.Params,
		nil, // No dependency input for single task run
		3,
		context.Background(),
	)
	duration := time.Since(startTime)

	// We need to copy the output table back to the original state
	var finalOutput *lua.LTable
	if output != nil {
		finalOutput = utils.CopyTable(output, L)
	}

	return success, msg, finalOutput, duration, err
}