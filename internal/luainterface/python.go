package luainterface

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yuin/gopher-lua"
)

const pythonVenvTypeName = "python_venv"

// PythonVenv é o struct Go que representa um ambiente virtual Python.
// Ele armazena o caminho para o diretório do venv.
type PythonVenv struct {
	VenvPath string
}

// runCommand é uma função auxiliar para executar comandos do sistema de forma segura,
// capturando stdout e stderr. Retorna o sucesso da operação e as saídas.
func runCommand(command string, args ...string) (bool, string, string) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	return success, strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String())
}

// newPythonVenv é a função construtora exposta ao Lua como `python:venv(path)`.
// Ela cria um userdata do tipo PythonVenv.
func newPythonVenv(L *lua.LState) int {
	path := L.CheckString(1)
	venv := &PythonVenv{VenvPath: path}

	ud := L.NewUserData()
	ud.Value = venv
	L.SetMetatable(ud, L.GetTypeMetatable(pythonVenvTypeName))
	L.Push(ud)
	return 1
}

// venvExists verifica se o ambiente virtual parece existir.
// A verificação é feita pela presença do arquivo 'bin/activate'.
func venvExists(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*PythonVenv)
	activatePath := filepath.Join(venv.VenvPath, "bin", "activate")

	_, err := os.Stat(activatePath)
	L.Push(lua.LBool(err == nil))
	return 1
}

// venvCreate executa `python3 -m venv <path>` para criar o ambiente virtual.
func venvCreate(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*PythonVenv)
	success, stdout, stderr := runCommand("python3", "-m", "venv", venv.VenvPath)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(success))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// venvPip executa um comando `pip` dentro do contexto do venv.
// Ex: venv:pip("install -r requirements.txt")
func venvPip(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*PythonVenv)
	argsStr := L.CheckString(2)
	args := strings.Fields(argsStr) // Divide a string de argumentos em um slice

	pipPath := filepath.Join(venv.VenvPath, "bin", "pip")
	success, stdout, stderr := runCommand(pipPath, args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(success))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// venvExec executa um comando `python` dentro do contexto do venv.
// Ex: venv:exec("app.py --port 8080")
func venvExec(L *lua.LState) int {
	venv := L.CheckUserData(1).Value.(*PythonVenv)
	argsStr := L.CheckString(2)
	args := strings.Fields(argsStr)

	pythonPath := filepath.Join(venv.VenvPath, "bin", "python")
	success, stdout, stderr := runCommand(pythonPath, args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(success))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// Métodos que serão registrados para o tipo PythonVenv em Lua.
var pythonVenvMethods = map[string]lua.LGFunction{
	"exists": venvExists,
	"create": venvCreate,
	"pip":    venvPip,
	"exec":   venvExec,
}

// registerPythonVenvType cria e registra a metatable para o nosso tipo customizado.
func registerPythonVenvType(L *lua.LState) {
	mt := L.NewTypeMetatable(pythonVenvTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), pythonVenvMethods))
}

// PythonLoader is the function that the gopher-lua usará para carregar o módulo `python`.
func PythonLoader(L *lua.LState) int {
	// Cria a tabela principal do módulo
	mod := L.NewTable()

	// Registra o tipo PythonVenv e seus métodos
	registerPythonVenvType(L)

	// Define a função `python:venv(path)`
	L.SetField(mod, "venv", L.NewFunction(newPythonVenv))

	// Retorna a tabela do módulo
	L.Push(mod)
	return 1
}

func OpenPython(L *lua.LState) {
	L.PreloadModule("python", PythonLoader)
}
