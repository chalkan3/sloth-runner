package runner

import (
	"context"
	"time"

	"github.com/chalkan3/sloth-runner/internal/lua/utils"
	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
)

// RunSingleTask executes a single task in a controlled environment for testing.
// It does not handle dependencies.
func RunSingleTask(L *lua.LState, task *types.Task) (bool, string, *lua.LTable, time.Duration, error) {
	startTime := time.Now()

	// We need a new Lua state to isolate the execution
	taskL := lua.NewState()
	defer taskL.Close()
	// luainterface.OpenAll(taskL) // This would re-introduce the cycle. We need a selective opener.

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
