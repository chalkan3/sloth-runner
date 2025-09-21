package luainterface

import (
	"bytes"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

type gcp struct{}

func OpenGCP(L *lua.LState) {
	L.PreloadModule("gcp", func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"exec": gcpExec,
		})
		L.Push(mod)
		return 1
	})
}

// gcpExec executes a gcloud command.
// Lua usage: gcp.exec({"arg1", "arg2", ...})
// Example: gcp.exec({"compute", "instances", "list", "--project", "my-project"})
func gcpExec(L *lua.LState) int {
	argsTable := L.CheckTable(1)
	var args []string
	argsTable.ForEach(func(_, value lua.LValue) {
		args = append(args, lua.LVAsString(value))
	})

	cmd := ExecCommand("gcloud", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Non-exit error (e.g., command not found)
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2 // result, error
		}
	}

	// Create a Lua table to return the results
	resultTable := L.NewTable()
	resultTable.RawSetString("stdout", lua.LString(stdout.String()))
	resultTable.RawSetString("stderr", lua.LString(stderr.String()))
	resultTable.RawSetString("exit_code", lua.LNumber(exitCode))

	L.Push(resultTable)
	return 1 // result
}