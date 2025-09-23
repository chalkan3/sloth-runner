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
	OpenGCP(L)
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

func TestGCP_ServiceAccount(t *testing.T) {
	L, cleanup := setupTest(t)
	defer cleanup()

	mockExitCode = "0"
	mockStdout = ""
	mockStderr = ""

	script := `
		local gcp = require('gcp')
		local client = gcp.client({ project = "my-project" })
		local sa = client:service_account("my-service-account")
		sa:create({ display_name = "My Service Account" })
	`
	err := L.DoString(script)
	assert.NoError(t, err)

	if assert.Equal(t, 1, len(commandsCalled)) {
		assert.Equal(t, []string{"gcloud", "--project", "my-project", "iam", "service-accounts", "create", "my-service-account", "--display-name", "My Service Account"}, commandsCalled[0])
	}
}

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
	mockStderr = "salt error"
	script := `
		local salt = require('salt')
		local client = salt.client()
		local stdout, stderr, err = client:target("*", "glob"):cmd("test.ping")
		assert_equal("", stdout)
		assert_equal("salt error", stderr)
		assert_not_nil(err)
	`
	// Helper function for assertions in Lua
	L.SetGlobal("assert_equal", L.NewFunction(func(L *lua.LState) int {
		expected := L.ToString(1)
		actual := L.ToString(2)
		assert.Equal(t, expected, actual)
		return 0
	}))
	L.SetGlobal("assert_not_nil", L.NewFunction(func(L *lua.LState) int {
		assert.NotNil(t, L.Get(1))
		return 0
	}))
	err := L.DoString(script)
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
		local client = salt.client({config = ""})
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
	assert.Equal(t, []string{"salt", "--out=json", "-L", "keiteguica", "test.ping"}, commandsCalled[0])
	assert.Equal(t, []string{"salt", "--out=json", "-L", "vm-gcp-squid-proxy*", "test.ping"}, commandsCalled[1])
	assert.Equal(t, []string{"salt", "--out=json", "-L", "vm-gcp-squid-proxy*", "pkg.upgrade"}, commandsCalled[2])
	assert.Equal(t, []string{"salt", "--out=json", "-L", "keiteguica", "cmd.run", "ls -la"}, commandsCalled[3])
}
