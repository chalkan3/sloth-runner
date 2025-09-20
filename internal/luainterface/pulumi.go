package luainterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
)

const (

luaPulumiStackTypeName = "sloth.pulumiStack"
)

type pulumiStack struct {
	Name     string
	WorkDir  string
	Venv     *types.PythonVenv
	LoginURL string
}

// setupPulumiCmd creates and new-configured an exec.Cmd for a Pulumi command.
func setupPulumiCmd(stack *pulumiStack, commandParts ...string) *exec.Cmd {

pulumiCmd := "pulumi " + strings.Join(commandParts, " ")

	// Chain commands
	var commands []string

	// Activate venv if present
	if stack.Venv != nil && stack.Venv.Path != "" {
		activateScript := filepath.Join(stack.Venv.Path, "bin", "activate")
		commands = append(commands, fmt.Sprintf("source %s", activateScript))
	}

	// Login if URL is present
	if stack.LoginURL != "" {
		commands = append(commands, fmt.Sprintf("pulumi login %s", stack.LoginURL))
	}

	// Add the actual pulumi command
	commands = append(commands, pulumiCmd)

	fullCommand := strings.Join(commands, " && ")

	cmd := exec.Command("bash", "-c", fullCommand)
	cmd.Dir = stack.WorkDir

	// Prepend Pulumi bin to PATH
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Handle error, maybe log it or return an error
	}
	pulumiPath := filepath.Join(homeDir, ".pulumi", "bin")
	newPath := fmt.Sprintf("PATH=%s:%s", pulumiPath, os.Getenv("PATH"))

	// Create a new environment slice and add the modified PATH
	env := os.Environ()
	found := false
	for i, v := range env {
		if strings.HasPrefix(v, "PATH=") {
			env[i] = newPath
			found = true
			break
		}
	}
	if !found {
		env = append(env, newPath)
	}
	cmd.Env = env

	return cmd
}

// pulumi:stack(name, {workdir="path", venv=venv_obj, login_url="url"}) -> stack
func pulumiStackFn(L *lua.LState) int {
	name := L.CheckString(1)
	opts := L.CheckTable(2)
	workdir := opts.RawGetString("workdir").String()
	loginURL := opts.RawGetString("login_url").String()
	venvUD := opts.RawGetString("venv")

	if workdir == "" {
		L.RaiseError("the 'workdir' field is required for pulumi:stack")
		return 0
	}

	var venv *types.PythonVenv
	if venvUD.Type() == lua.LTUserData {
		if v, ok := venvUD.(*lua.LUserData).Value.(*types.PythonVenv); ok {
			venv = v
		}
	}

	stack := &pulumiStack{
		Name:     name,
		WorkDir:  workdir,
		Venv:     venv,
		LoginURL: loginURL,
	}

	ud := L.NewUserData()
	ud.Value = stack
	L.SetMetatable(ud, L.GetTypeMetatable(luaPulumiStackTypeName))
	L.Push(ud)
	return 1
}

func checkPulumiStack(L *lua.LState, n int) *pulumiStack {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*pulumiStack); ok {
		return v
	}
	L.ArgError(n, "expected pulumi stack object")
	return nil
}

func runPulumiCommand(L *lua.LState, command string) int {
	stack := checkPulumiStack(L, 1)

	pulumiArgs := []string{command, "--stack", stack.Name}
	if L.GetTop() >= 2 {
		opts := L.CheckTable(2)
		if lua.LVAsBool(opts.RawGetString("yes")) {
			pulumiArgs = append(pulumiArgs, "--yes")
		}
		if lua.LVAsBool(opts.RawGetString("skip_preview")) {
			pulumiArgs = append(pulumiArgs, "--skip-preview")
		}
	}

	cmd := setupPulumiCmd(stack, pulumiArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("success", lua.LBool(success))
	L.Push(result)
	return 1
}

func pulumiStackUp(L *lua.LState) int {
	return runPulumiCommand(L, "up")
}

func pulumiStackPreview(L *lua.LState) int {
	return runPulumiCommand(L, "preview")
}

func pulumiStackDestroy(L *lua.LState) int {
	return runPulumiCommand(L, "destroy")
}

func pulumiStackOutputs(L *lua.LState) int {
	stack := checkPulumiStack(L, 1)
	cmd := setupPulumiCmd(stack, "stack", "output", "--json")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{} // Ignore stderr for outputs

	err := cmd.Run()
	if err != nil {
		L.RaiseError("failed to get outputs: %v\n%s", err, cmd.Stderr.(*bytes.Buffer).String())
		return 0
	}

	var outputs map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &outputs); err != nil {
		L.RaiseError("failed to parse pulumi outputs json: %v", err)
		return 0
	}

	L.Push(GoValueToLua(L, outputs))
	return 1
}

func pulumiStackConfig(L *lua.LState) int {
	stack := checkPulumiStack(L, 1)
	key := L.CheckString(2)
	value := L.CheckString(3)
	isSecret := L.OptBool(4, false)

	quotedValue := fmt.Sprintf("'%s'", value)

	configCmdParts := []string{"config", "set", key, quotedValue, "--stack", stack.Name}
	if isSecret {
		configCmdParts = append(configCmdParts, "--secret")
	}

	cmd := setupPulumiCmd(stack, configCmdParts...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("success", lua.LBool(success))
	L.Push(result)
	return 1
}

var pulumiMethods = map[string]lua.LGFunction{
	"stack":          pulumiStackFn,
	"login":          pulumiLoginFn,
	"install_plugin": pulumiInstallPluginFn,
}

func pulumiInstallPluginFn(L *lua.LState) int {
	pluginName := L.CheckString(1)
	fullCommand := fmt.Sprintf("pulumi plugin install language %s --reinstall", pluginName)
	cmd := exec.Command("bash", "-c", fullCommand)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("success", lua.LBool(success))
	L.Push(result)
	return 1
}

func pulumiLoginFn(L *lua.LState) int {
	url := L.OptString(1, "")
	cmd := exec.Command("pulumi", "login", url)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("success", lua.LBool(success))
	L.Push(result)
	return 1
}

var pulumiStackMethods = map[string]lua.LGFunction{
	"up":      pulumiStackUp,
	"preview": pulumiStackPreview,
	"destroy": pulumiStackDestroy,
	"outputs": pulumiStackOutputs,
	"config":  pulumiStackConfig,
	"select":  pulumiStackSelect,
}

func pulumiStackSelect(L *lua.LState) int {
	stack := checkPulumiStack(L, 1)

	args := []string{"stack", "select", stack.Name}
	if lua.LVAsBool(L.OptTable(2, L.NewTable()).RawGetString("create")) {
		args = append(args, "--create")
	}

	cmd := setupPulumiCmd(stack, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("success", lua.LBool(success))
	L.Push(result)
	return 1
}

func PulumiLoader(L *lua.LState) int {
	mt := L.NewTypeMetatable(luaPulumiStackTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), pulumiStackMethods))
	mod := L.SetFuncs(L.NewTable(), pulumiMethods)
	L.Push(mod)
	return 1
}

func OpenPulumi(L *lua.LState) {
	L.PreloadModule("pulumi", PulumiLoader)
}
