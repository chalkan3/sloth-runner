package luainterface

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaPulumiStackTypeName = "pulumiStack"
)

type pulumiStack struct {
	Name     string
	WorkDir  string
	VenvPath string // Novo campo para o caminho do venv
}

// pulumi:stack(name, {workdir="path", venv_path="path"}) -> stack
func pulumiStackFn(L *lua.LState) int {
	name := L.CheckString(1)
	opts := L.CheckTable(2)
	workdir := opts.RawGetString("workdir").String()
	venvPath := opts.RawGetString("venv_path").String() // Lê o novo campo

	if workdir == "" {
		L.RaiseError("o campo 'workdir' é obrigatório para pulumi:stack")
		return 0
	}

	stack := &pulumiStack{
		Name:     name,
		WorkDir:  workdir,
		VenvPath: venvPath, // Armazena o caminho
	}

	ud := L.NewUserData()
	ud.Value = stack
	L.SetMetatable(ud, L.GetTypeMetatable(luaPulumiStackTypeName))
	L.Push(ud)
	return 1
}

// --- Métodos do Objeto Stack ---

func checkPulumiStack(L *lua.LState, n int) *pulumiStack {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*pulumiStack); ok {
		return v
	}
	L.ArgError(n, "esperado objeto stack do pulumi")
	return nil
}

func runPulumiCommand(L *lua.LState, command string) int {
	stack := checkPulumiStack(L, 1)

	// Check if a session object is passed as the third argument
	var session *types.SharedSession
	if L.GetTop() >= 3 {
		if ud, ok := L.Get(3).(*lua.LUserData); ok {
			if s, ok := ud.Value.(*types.SharedSession); ok {
				session = s
			}
		}
	}

	pulumiArgs := []string{command, "--stack", stack.Name}
	var stderrFile lua.LValue
	if L.GetTop() >= 2 {
		opts := L.CheckTable(2)
		if lua.LVAsBool(opts.RawGetString("yes")) {
			pulumiArgs = append(pulumiArgs, "--yes")
		}
		if lua.LVAsBool(opts.RawGetString("skip_preview")) {
			pulumiArgs = append(pulumiArgs, "--skip-preview")
		}
		stderrFile = opts.RawGetString("stderr_file")
	}

	if session != nil {
		// Execute in shared session
		fullCommand := "pulumi " + strings.Join(pulumiArgs, " ")
		stdout, stderr, err := session.ExecuteCommand(fullCommand, stack.WorkDir)
		success := err == nil

		result := L.NewTable()
		result.RawSetString("stdout", lua.LString(stdout))
		result.RawSetString("stderr", lua.LString(stderr))
		result.RawSetString("success", lua.LBool(success))
		L.Push(result)
		return 1
	}

	// Fallback to isolated execution
	pulumiPath := "pulumi"
	cmd := exec.Command(pulumiPath, pulumiArgs...)
	cmd.Dir = stack.WorkDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	if stderrFile != nil && stderrFile != lua.LNil {
		f, err := os.OpenFile(stderrFile.String(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			L.RaiseError("failed to open stderr file: %v", err)
		}
		cmd.Stderr = f
		defer f.Close()
	} else {
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("success", lua.LBool(success))
	L.Push(result)
	return 1
}

// stack:up({ yes=true, skip_preview=true })
func pulumiStackUp(L *lua.LState) int {
	return runPulumiCommand(L, "up")
}

// stack:preview()
func pulumiStackPreview(L *lua.LState) int {
	return runPulumiCommand(L, "preview")
}

// stack:destroy({ yes=true })
func pulumiStackDestroy(L *lua.LState) int {
	return runPulumiCommand(L, "destroy")
}

// stack:outputs() -> table
func pulumiStackOutputs(L *lua.LState) int {
	stack := checkPulumiStack(L, 1)
	fmt.Printf("Obtendo saídas (outputs) para a stack '%s' em '%s'\n", stack.Name, stack.WorkDir)
	// Simula o retorno de uma tabela de saídas
	outputs := L.NewTable()
	L.SetField(outputs, "url", lua.LString("http://example-app.com"))
	L.SetField(outputs, "bucket_name", lua.LString("my-static-content-bucket"))
	L.Push(outputs)
	return 1
}

// stack:config(key, value, is_secret)
func pulumiStackConfig(L *lua.LState) int {
	stack := checkPulumiStack(L, 1)
	key := L.CheckString(2)
	value := L.CheckString(3)
	isSecret := L.OptBool(4, false)

	args := []string{"config", "set", key, value, "--stack", stack.Name}
	if isSecret {
		args = append(args, "--secret")
	}

	// O comando 'pulumi' deve ser executado no diretório de trabalho da stack.
	pulumiPath := "pulumi"

	cmd := exec.Command(pulumiPath, args...)
	cmd.Dir = stack.WorkDir
	cmd.Env = os.Environ()

	if stack.VenvPath != "" {
		cmd.Env = append(cmd.Env, "VIRTUAL_ENV="+stack.VenvPath)
		newPath := filepath.Join(stack.VenvPath, "bin") + ":" + os.Getenv("PATH")
		pathUpdated := false
		for i, v := range cmd.Env {
			if strings.HasPrefix(v, "PATH=") {
				cmd.Env[i] = "PATH=" + newPath
				pathUpdated = true
				break
			}
		}
		if !pathUpdated {
			cmd.Env = append(cmd.Env, "PATH="+newPath)
		}
	}

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
	"stack": pulumiStackFn,
}

var pulumiStackMethods = map[string]lua.LGFunction{
	"up":      pulumiStackUp,
	"preview": pulumiStackPreview,
	"destroy": pulumiStackDestroy,
	"outputs": pulumiStackOutputs,
	"config":  pulumiStackConfig,
	"select":  pulumiStackSelect, // Novo método adicionado
}

// stack:select({ create = true })
func pulumiStackSelect(L *lua.LState) int {
	stack := checkPulumiStack(L, 1)

	args := []string{"stack", "select", stack.Name}
	if lua.LVAsBool(L.OptTable(2, L.NewTable()).RawGetString("create")) {
		args = append(args, "--create")
	}

	pulumiPath := "pulumi"

	cmd := exec.Command(pulumiPath, args...)
	cmd.Dir = stack.WorkDir
	cmd.Env = os.Environ()

	if stack.VenvPath != "" {
		cmd.Env = append(cmd.Env, "VIRTUAL_ENV="+stack.VenvPath)
		newPath := filepath.Join(stack.VenvPath, "bin") + ":" + os.Getenv("PATH")
		pathUpdated := false
		for i, v := range cmd.Env {
			if strings.HasPrefix(v, "PATH=") {
				cmd.Env[i] = "PATH=" + newPath
				pathUpdated = true
				break
			}
		}
		if !pathUpdated {
			cmd.Env = append(cmd.Env, "PATH="+newPath)
		}
	}

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
	// Registra o tipo 'pulumiStack' com seus métodos
	mt := L.NewTypeMetatable(luaPulumiStackTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), pulumiStackMethods))

	// Registra o módulo 'pulumi' com seu método construtor
	mod := L.SetFuncs(L.NewTable(), pulumiMethods)
	L.Push(mod)
	return 1
}

func OpenPulumi(L *lua.LState) {
	L.PreloadModule("pulumi", PulumiLoader)
}
