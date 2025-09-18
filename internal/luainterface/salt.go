package luainterface

import (
	"bytes"
	"log"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// luaSaltTargetTypeName is the name of the Lua userdata type for SaltTarget.
const luaSaltTargetTypeName = "salt_target"

// SaltTarget holds the state for a fluent Salt API call.
type SaltTarget struct {
	Target      string
	TargetType  string // e.g., "glob", "list", "pcre"
	lastSuccess bool
	lastStdout  string
	lastStderr  string
	lastError   error // Go error
}

// OpenSalt registers the 'salt' module with the Lua state.
func OpenSalt(L *lua.LState) {
	// Create the metatable for the SaltTarget type.
	mt := L.NewTypeMetatable(luaSaltTargetTypeName)
	L.SetGlobal(luaSaltTargetTypeName, mt) // Optional: make metatable available globally.

	// Register methods for the SaltTarget type.
	methods := map[string]lua.LGFunction{
		"ping":   saltTargetPing,
		"cmd":    saltTargetCmd,
		"result": saltTargetResult, // Method to get results of the last operation
	}
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))

	// Create the main 'salt' module table.
	saltModule := L.NewTable()

	// Register top-level functions like salt.target().
	saltFuncs := map[string]lua.LGFunction{
		"target": saltTarget,
	}
	L.SetFuncs(saltModule, saltFuncs)

	// Make the 'salt' module available globally.
	L.SetGlobal("salt", saltModule)
}

// checkSaltTarget retrieves the SaltTarget struct from a Lua userdata.
func checkSaltTarget(L *lua.LState) *SaltTarget {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*SaltTarget); ok {
		return v
	}
	L.ArgError(1, "salt_target expected")
	return nil
}

// runSaltCommand executes a Salt CLI command and updates the SaltTarget's last operation status.
func runSaltCommand(L *lua.LState, target *SaltTarget, args []string) {
	cmdArgs := []string{target.TargetType, target.Target}
	cmdArgs = append(cmdArgs, args...)

	cmd := ExecCommand("salt", cmdArgs...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	log.Printf("Executing Salt command: salt %s", strings.Join(cmdArgs, " "))

	err := cmd.Run()

	target.lastStdout = stdoutBuf.String()
	target.lastStderr = stderrBuf.String()
	if err != nil {
		target.lastSuccess = false
		target.lastError = err
	} else {
		target.lastSuccess = true
		target.lastError = nil
	}
}

// saltTarget implements the salt.target(target, target_type) function.
func saltTarget(L *lua.LState) int {
	targetStr := L.CheckString(1)
	targetType := L.OptString(2, "glob") // Default to 'glob'

	saltT := &SaltTarget{
		Target:     targetStr,
		TargetType: targetType,
	}
	ud := L.NewUserData()
	ud.Value = saltT
	L.SetMetatable(ud, L.GetTypeMetatable(luaSaltTargetTypeName))
	L.Push(ud)
	return 1
}

// saltTargetPing implements the SaltTarget:ping() method.
func saltTargetPing(L *lua.LState) int {
	target := checkSaltTarget(L)
	if target == nil {
		return 0
	}
	runSaltCommand(L, target, []string{"test.ping"})
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// saltTargetCmd implements the SaltTarget:cmd(module_function, args...) method.
func saltTargetCmd(L *lua.LState) int {
	target := checkSaltTarget(L)
	if target == nil {
		return 0
	}

	moduleFunc := L.CheckString(2)
	var args []string
	for i := 3; i <= L.GetTop(); i++ {
		args = append(args, L.CheckString(i))
	}

	cmdArgs := []string{moduleFunc}
	cmdArgs = append(cmdArgs, args...)

	runSaltCommand(L, target, cmdArgs)
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// saltTargetResult implements the SaltTarget:result() method, returning the last command's output.
func saltTargetResult(L *lua.LState) int {
	target := checkSaltTarget(L)
	if target == nil {
		return 0
	}

	resultTable := L.NewTable()
	resultTable.RawSetString("success", lua.LBool(target.lastSuccess))
	resultTable.RawSetString("stdout", lua.LString(target.lastStdout))
	resultTable.RawSetString("stderr", lua.LString(target.lastStderr))
	if target.lastError != nil {
		resultTable.RawSetString("error", lua.LString(target.lastError.Error()))
	} else {
		resultTable.RawSetString("error", lua.LNil)
	}
	L.Push(resultTable)
	return 1
}