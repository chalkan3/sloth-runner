package luainterface

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaRepoTypeName = "repo"
)

// --- Métodos do Objeto Repo ---

// repo:pull()
func repoPull(L *lua.LState) int {
	repo := checkRepo(L, 1)
	// Lógica de pull (exemplo)
	fmt.Printf("Executando git pull no repositório em %s\n", repo.Path)
	// Aqui entraria a lógica real do go-git para pull
	L.Push(L.Get(1)) // Retorna o próprio objeto para encadeamento
	return 1
}

// repo:commit(message)
func repoCommit(L *lua.LState) int {
	repo := checkRepo(L, 1)
	message := L.CheckString(2)
	// Lógica de commit (exemplo)
	fmt.Printf("Executando git commit em %s com a mensagem: '%s'\n", repo.Path, message)
	// Aqui entraria a lógica real do go-git para commit
	L.Push(L.Get(1)) // Retorna o próprio objeto
	return 1
}

// repo:push()
func repoPush(L *lua.LState) int {
	repo := checkRepo(L, 1)
	// Lógica de push (exemplo)
	fmt.Printf("Executando git push no repositório em %s\n", repo.Path)
	// Aqui entraria a lógica real do go-git para push
	L.Push(L.Get(1)) // Retorna o próprio objeto
	return 1
}

// --- Métodos do Módulo Git ---

// git:clone(url, path) -> { success, stdout, stderr }
func gitClone(L *lua.LState) int {
	url := L.CheckString(1)
	path := L.CheckString(2)

	// Usa os/exec para ter uma saída mais clara, similar aos outros módulos.
	cmd := exec.Command("git", "clone", url, path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(success))
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	L.Push(result)
	return 1
}

// git:repo(path) -> repo (Este método parece obsoleto, mas vamos mantê-lo por enquanto)
func gitRepo(L *lua.LState) int {
	path := L.CheckString(1)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		L.RaiseError("diretório do repositório não encontrado em: %s", path)
		return 0
	}
	// A API original deste método não está clara, retornando sucesso por padrão.
	result := L.NewTable()
	result.RawSetString("success", lua.LBool(true))
	L.Push(result)
	return 1
}

// --- Funções Auxiliares ---

// Estrutura Go para armazenar os dados do repositório
type RepoData struct {
	Path          string
	URL           string
	CurrentBranch string
}

// Cria e retorna o objeto repo (tabela Lua)
func createRepoObject(L *lua.LState, path string) int {
	absPath, err := filepath.Abs(path)
	if err != nil {
		L.RaiseError("não foi possível obter o caminho absoluto para: %s", path)
		return 0
	}

	repoData := &RepoData{Path: absPath}

	// Tenta abrir o repositório para preencher mais dados
	r, err := git.PlainOpen(absPath)
	if err == nil {
		// Obter URL remota
		remote, err := r.Remote("origin")
		if err == nil {
			repoData.URL = remote.Config().URLs[0]
		}
		// Obter branch atual
		head, err := r.Head()
		if err == nil {
			repoData.CurrentBranch = head.Name().Short()
		}
	}

	// Cria a tabela principal para o objeto repo
	repoTable := L.NewTable()

	// Cria e preenche a tabela 'local'
	localTable := L.NewTable()
	L.SetField(localTable, "path", lua.LString(repoData.Path))
	L.SetField(repoTable, "local", localTable)

	// Cria e preenche a tabela 'remote'
	remoteTable := L.NewTable()
	L.SetField(remoteTable, "url", lua.LString(repoData.URL))
	L.SetField(repoTable, "remote", remoteTable)

	// Cria e preenche a tabela 'current'
	currentTable := L.NewTable()
	L.SetField(currentTable, "branch", lua.LString(repoData.CurrentBranch))
	L.SetField(repoTable, "current", currentTable)

	// Anexa o RepoData como userdata para uso interno nos métodos
	ud := L.NewUserData()
	ud.Value = repoData
	L.SetField(repoTable, "__internal", ud)

	// Define a metatable que aponta para os métodos do repo
	L.SetMetatable(repoTable, L.GetTypeMetatable(luaRepoTypeName))

	L.Push(repoTable)
	return 1
}

// Verifica se o argumento é um objeto repo e retorna o RepoData interno
func checkRepo(L *lua.LState, n int) *RepoData {
	tbl := L.CheckTable(n)
	ud, ok := L.GetField(tbl, "__internal").(*lua.LUserData)
	if !ok {
		L.ArgError(n, "objeto repo inválido")
	}
	repo, ok := ud.Value.(*RepoData)
	if !ok {
		L.ArgError(n, "userdata de repo inválido")
	}
	return repo
}

var gitMethods = map[string]lua.LGFunction{
	"clone": gitClone,
	"repo":  gitRepo,
}

var repoMethods = map[string]lua.LGFunction{
	"pull":   repoPull,
	"commit": repoCommit,
	"push":   repoPush,
}

func GitLoader(L *lua.LState) int {
	// Registra o tipo 'repo' com seus métodos
	mt := L.NewTypeMetatable(luaRepoTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), repoMethods))

	// Registra o módulo 'git' com seus métodos
	mod := L.SetFuncs(L.NewTable(), gitMethods)
	L.Push(mod)
	return 1
}

func OpenGit(L *lua.LState) {
	L.PreloadModule("git", GitLoader)
}
