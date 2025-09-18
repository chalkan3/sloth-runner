package luainterface

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// luaGitRepoTypeName is the name of the Lua userdata type for GitRepo.
const luaGitRepoTypeName = "git_repo"

// GitRepo holds the state for a fluent Git API call.
type GitRepo struct {
	RepoPath string
	// Store the result of the last operation for inspection in Lua
	lastSuccess bool
	lastStdout  string
	lastStderr  string
	lastError   error // Go error
}

// OpenGit registers the 'git' module with the Lua state.
func OpenGit(L *lua.LState) {
	// Create the metatable for the GitRepo type.
	mt := L.NewTypeMetatable(luaGitRepoTypeName)
	L.SetGlobal(luaGitRepoTypeName, mt) // Optional: make metatable available globally.

	// Register methods for the GitRepo type.
	methods := map[string]lua.LGFunction{
		"checkout": gitRepoCheckout,
		"pull":     gitRepoPull,
		"add":      gitRepoAdd,
		"commit":   gitRepoCommit,
		"tag":      gitRepoTag,
		"push":     gitRepoPush,
		"result":   gitRepoResult, // New method to get results of the last operation
	}
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))

	// Create the main 'git' module table.
	gitModule := L.NewTable()

	// Register top-level functions like git.clone() and git.repo().
	gitFuncs := map[string]lua.LGFunction{
		"clone": gitClone,
		"repo":  gitRepo,
	}
	L.SetFuncs(gitModule, gitFuncs)

	// Make the 'git' module available globally.
	L.SetGlobal("git", gitModule)
}

// checkGitRepo retrieves the GitRepo struct from a Lua userdata.
func checkGitRepo(L *lua.LState) *GitRepo {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GitRepo); ok {
		return v
	}
	L.ArgError(1, "git_repo expected")
	return nil
}

// runGitCommand executes a Git CLI command and updates the GitRepo's last operation status.
func runGitCommand(L *lua.LState, repo *GitRepo, args []string) {
	cmd := ExecCommand("git", args...)
	cmd.Dir = repo.RepoPath // Set working directory for the git command
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	log.Printf("Executing Git command in %s: git %s", repo.RepoPath, strings.Join(args, " "))

	err := cmd.Run()

	repo.lastStdout = stdoutBuf.String()
	repo.lastStderr = stderrBuf.String()
	if err != nil {
		repo.lastSuccess = false
		repo.lastError = err
	} else {
		repo.lastSuccess = true
		repo.lastError = nil
	}
}

// gitClone implements the git.clone(url, path) function.
func gitClone(L *lua.LState) int {
	repoURL := L.CheckString(1)
	repoPath := L.CheckString(2)

	// Check if path exists and is already a git repo
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		L.Push(lua.LNil) // Return nil for the repo object
		L.Push(lua.LString(fmt.Sprintf("path %s already contains a git repository", repoPath))) // Return error message
		return 2
	}

	cmd := ExecCommand("git", "clone", repoURL, repoPath)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	log.Printf("Cloning repository: git clone %s %s", repoURL, repoPath)

	err := cmd.Run()

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("failed to clone repository: %s, stdout: %s, stderr: %s", err.Error(), stdoutBuf.String(), stderrBuf.String())))
		return 2
	}

	repo := &GitRepo{
		RepoPath:    repoPath,
		lastSuccess: true,
		lastStdout:  stdoutBuf.String(),
		lastStderr:  stderrBuf.String(),
		lastError:   nil,
	}
	ud := L.NewUserData()
	ud.Value = repo
	L.SetMetatable(ud, L.GetTypeMetatable(luaGitRepoTypeName))
	L.Push(ud)
	L.Push(lua.LNil) // No error
	return 2
}

// gitRepo implements the git.repo(path) function.
func gitRepo(L *lua.LState) int {
	repoPath := L.CheckString(1)

	// Check if it's a valid git repository
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("path %s is not a git repository", repoPath)))
		return 2
	}

	repo := &GitRepo{
		RepoPath:    repoPath,
		lastSuccess: true, // Assume success if it's a valid repo
		lastStdout:  "",
		lastStderr:  "",
		lastError:   nil,
	}
	ud := L.NewUserData()
	ud.Value = repo
	L.SetMetatable(ud, L.GetTypeMetatable(luaGitRepoTypeName))
	L.Push(ud)
	L.Push(lua.LNil) // No error
	return 2
}

// gitRepoCheckout implements the GitRepo:checkout(ref) method.
func gitRepoCheckout(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}
	ref := L.CheckString(2)
	runGitCommand(L, repo, []string{"checkout", ref})
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// gitRepoPull implements the GitRepo:pull(remote, branch) method.
func gitRepoPull(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}
	remote := L.CheckString(2)
	branch := L.CheckString(3)
	runGitCommand(L, repo, []string{"pull", remote, branch})
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// gitRepoAdd implements the GitRepo:add(pattern) method.
func gitRepoAdd(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}
	pattern := L.CheckString(2)
	runGitCommand(L, repo, []string{"add", pattern})
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// gitRepoCommit implements the GitRepo:commit(message) method.
func gitRepoCommit(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}
	message := L.CheckString(2)
	runGitCommand(L, repo, []string{"commit", "-m", message})
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// gitRepoTag implements the GitRepo:tag(name, message) method.
func gitRepoTag(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}
	name := L.CheckString(2)
	message := L.OptString(3, "") // Optional message
	args := []string{"tag", name}
	if message != "" {
		args = append(args, "-m", message)
	}
	runGitCommand(L, repo, args)
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// gitRepoPush implements the GitRepo:push(remote, branch, options) method.
func gitRepoPush(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}
	remote := L.CheckString(2)
	branch := L.CheckString(3)
	optionsTable := L.OptTable(4, L.NewTable()) // Optional options table

	args := []string{"push", remote, branch}

	if followTags := optionsTable.RawGetString("follow_tags"); followTags.Type() == lua.LTBool && lua.LVAsBool(followTags) {
		args = append(args, "--follow-tags")
	}
	// Add other options as needed (e.g., --force, --set-upstream)

	runGitCommand(L, repo, args)
	L.Push(L.CheckUserData(1)) // Return self for chaining
	return 1
}

// gitRepoResult implements the GitRepo:result() method, returning the last command's output.
func gitRepoResult(L *lua.LState) int {
	repo := checkGitRepo(L)
	if repo == nil {
		return 0
	}

	resultTable := L.NewTable()
	resultTable.RawSetString("success", lua.LBool(repo.lastSuccess))
	resultTable.RawSetString("stdout", lua.LString(repo.lastStdout))
	resultTable.RawSetString("stderr", lua.LString(repo.lastStderr))
	if repo.lastError != nil {
		resultTable.RawSetString("error", lua.LString(repo.lastError.Error()))
	} else {
		resultTable.RawSetString("error", lua.LNil)
	}
	L.Push(resultTable)
	return 1
}
