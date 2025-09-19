package luainterface

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// --- MOCKING INFRASTRUCTURE ---

var mockExitCode string
var mockStdout string
var mockStderr string
var commandsCalled [][]string

func mockExecCommand(command string, args ...string) *exec.Cmd {
	commandsCalled = append(commandsCalled, append([]string{command}, args...))
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_TEST_MODE=helper",
		fmt.Sprintf("GO_TEST_EXIT_CODE=%s", mockExitCode),
		fmt.Sprintf("GO_TEST_STDOUT=%s", mockStdout),
		fmt.Sprintf("GO_TEST_STDERR=%s", mockStderr),
	}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_MODE") != "helper" {
		return
	}
	fmt.Fprint(os.Stdout, os.Getenv("GO_TEST_STDOUT"))
	fmt.Fprint(os.Stderr, os.Getenv("GO_TEST_STDERR"))
	exitCode := 0
	fmt.Sscanf(os.Getenv("GO_TEST_EXIT_CODE"), "%d", &exitCode)
	os.Exit(exitCode)
}

func setupTest(t *testing.T) (*lua.LState, func()) {
	originalExecCommand := ExecCommand
	ExecCommand = mockExecCommand

	L := lua.NewState()
	L.PreloadModule("git", GitLoader)
	L.PreloadModule("pulumi", PulumiLoader)
	L.PreloadModule("salt", SaltLoader)
	// Open modules required by the examples
	OpenLog(L)
	OpenData(L)

	cleanup := func() {
		ExecCommand = originalExecCommand
		L.Close()
		mockExitCode = "0"
		mockStdout = ""
		mockStderr = ""
		commandsCalled = nil
	}

	return L, cleanup
}

// --- BASIC TESTS ---

func TestGitClone_Basic(t *testing.T) {
	L, cleanup := setupTest(t)
	defer cleanup()
	clonePath := t.TempDir()
	script := fmt.Sprintf(`
		local git = require('git')
		git.clone("https://a.com/b.git", "%s")
	`, clonePath)
	err := L.DoString(script)
	assert.NoError(t, err)
}

func TestPulumiStack_MissingWorkdir_NoError(t *testing.T) {
	L, cleanup := setupTest(t)
	defer cleanup()
	script := `
		local pulumi = require('pulumi')
		local stack, err = pulumi.stack("dev", {})
	`
	err := L.DoString(script)
	assert.NoError(t, err)
}

func TestSaltTarget_Cmd_Basic(t *testing.T) {
	L, cleanup := setupTest(t)
	defer cleanup()
	mockExitCode = "1"
	err := L.DoString(`
		local salt = require('salt')
		salt.target("*", "glob"):cmd("test.ping")
	`)
	assert.NoError(t, err)
}

// --- SALT EXAMPLE TESTS ---

func TestSaltExample_FluentAPI(t *testing.T) {
	L, cleanup := setupTest(t)
	defer cleanup()

	// Read the example file
	luaScript, err := ioutil.ReadFile("../../examples/fluent_salt_api_test.lua")
	assert.NoError(t, err)

	// Prepend require statements to the script content
	fullScript := `
		local salt = require('salt')
		local log = require('log')
		local data = require('data')
	` + string(luaScript)

	// Execute the script's command function
	err = L.DoString(fullScript)
	assert.NoError(t, err)

	fn := L.GetGlobal("command").(*lua.LFunction)
	err = L.CallByParam(lua.P{Fn: fn, NRet: 2})
	assert.NoError(t, err)

	// Assert that the correct salt commands were called in order
	assert.Equal(t, 4, len(commandsCalled))
	assert.Equal(t, []string{"salt", "glob", "keiteguica", "test.ping"}, commandsCalled[0])
	assert.Equal(t, []string{"salt", "glob", "vm-gcp-squid-proxy*", "test.ping"}, commandsCalled[1])
	assert.Equal(t, []string{"salt", "glob", "vm-gcp-squid-proxy*", "pkg.upgrade"}, commandsCalled[2])
	assert.Equal(t, []string{"salt", "glob", "keiteguica", "cmd.run", "ls -la"}, commandsCalled[3])
}
