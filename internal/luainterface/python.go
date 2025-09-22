package luainterface

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chalkan3/sloth-runner/internal/types"
	"github.com/yuin/gopher-lua"
)

const pythonVenvTypeName = "python_venv"

// runCommand is a helper function to execute system commands safely,
// capturing stdout and stderr. It returns the success of the operation and the outputs.
func runCommand(command string, args ...string) (bool, string, string) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	return success, strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String())
}

// newPythonVenv is the constructor function exposed to Lua as `python:venv(path)`.
// It creates a userdata of type types.PythonVenv.
func newPythonVenv(L *lua.LState) int {
	path := L.CheckString(1)
	venv := &types.PythonVenv{Path: path}

	ud := L.NewUserData()
	ud.Value = venv
	L.SetMetatable(ud, L.GetTypeMetatable(pythonVenvTypeName))
	L.Push(ud)
	return 1
}

// venvExists checks if the virtual environment seems to exist.
// The check is done by the presence of the 'bin/activate' file.
func venvExists(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*types.PythonVenv)
	activatePath := filepath.Join(venv.Path, "bin", "activate")

	_, err := os.Stat(activatePath)
	L.Push(lua.LBool(err == nil))
	return 1
}

// venvCreate executes `python3 -m venv <path>` to create the virtual environment.
func venvCreate(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*types.PythonVenv)
	success, _, stderr := runCommand("python3", "-m", "venv", venv.Path)
	if !success {
		L.RaiseError("failed to create python venv: %s", stderr)
	}
	L.Push(L.Get(1)) // Return self for chaining
	return 1
}

// venvPip executes a `pip` command within the context of the venv.
// Ex: venv:pip("install -r requirements.txt")
func venvPip(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*types.PythonVenv)
	argsStr := L.CheckString(2)
	args := strings.Fields(argsStr) // Split the argument string into a slice

	pipPath := filepath.Join(venv.Path, "bin", "pip")
	success, _, stderr := runCommand(pipPath, args...)
	if !success {
		L.RaiseError("failed to run pip command: %s", stderr)
	}
	L.Push(L.Get(1)) // Return self for chaining
	return 1
}

// venvExec executes a `python` command within the context of the venv.
// Ex: venv:exec("app.py --port 8080")
func venvExec(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*types.PythonVenv)
	argsStr := L.CheckString(2)
	args := strings.Fields(argsStr)

	pythonPath := filepath.Join(venv.Path, "bin", "python")
	success, stdout, stderr := runCommand(pythonPath, args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(success))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// Methods that will be registered for the types.PythonVenv type in Lua.
var pythonVenvMethods = map[string]lua.LGFunction{
	"exists": venvExists,
	"create": venvCreate,
	"pip":    venvPip,
	"exec":   venvExec,
}

// registerPythonVenvType creates and registers the metatable for our custom type.
func registerPythonVenvType(L *lua.LState) {
	mt := L.NewTypeMetatable(pythonVenvTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), pythonVenvMethods))
}

// PythonLoader is the function that gopher-lua will use to load the `python` module.
func PythonLoader(L *lua.LState) int {
	// Create the main module table
	mod := L.NewTable()

	// Register the types.PythonVenv type and its methods
	registerPythonVenvType(L)

	// Define the `python:venv(path)` function
	L.SetField(mod, "venv", L.NewFunction(newPythonVenv))

	// Return the module table
	L.Push(mod)
	return 1
}

func OpenPython(L *lua.LState) {
	L.PreloadModule("python", PythonLoader)
}
