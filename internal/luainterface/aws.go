package luainterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/yuin/gopher-lua"
)

// AWSModule provides AWS functionalities to Lua scripts
type AWSModule struct{}

// NewAWSModule creates a new AWSModule
func NewAWSModule() *AWSModule {
	return &AWSModule{}
}

// Loader returns the Lua loader for the aws module
func (mod *AWSModule) Loader(L *lua.LState) int {
	awsTable := L.NewTable()

	// Register aws.exec
	L.SetFuncs(awsTable, map[string]lua.LGFunction{
		"exec": mod.awsExec,
	})

	// Create and register the s3 submodule
	s3Module := L.NewTable()
	L.SetFuncs(s3Module, map[string]lua.LGFunction{
		"sync": mod.s3Sync,
	})
	awsTable.RawSetString("s3", s3Module)

	// Create and register the secretsmanager submodule
	secretsManagerModule := L.NewTable()
	L.SetFuncs(secretsManagerModule, map[string]lua.LGFunction{
		"get_secret": mod.secretsManagerGetSecret,
	})
	awsTable.RawSetString("secretsmanager", secretsManagerModule)

	L.Push(awsTable)
	return 1
}

// awsExec is the generic executor for AWS CLI commands.
// Lua usage: aws.exec({"s3", "ls"}, {profile = "my-profile"})
func (mod *AWSModule) awsExec(L *lua.LState) int {
	argsTable := L.CheckTable(1)
	optsTable := L.OptTable(2, L.NewTable())
	profile := optsTable.RawGetString("profile").String()

	var args []string
	argsTable.ForEach(func(_, val lua.LValue) {
		args = append(args, val.String())
	})

	var command string
	var commandArgs []string

	if profile != "" {
		command = "aws-vault"
		commandArgs = append([]string{"exec", profile, "--", "aws"}, args...)
	} else {
		command = "aws"
		commandArgs = args
	}

	cmd := exec.Command(command, commandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1 // Indicates an error other than a non-zero exit code
		}
	}

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("exit_code", lua.LNumber(exitCode))

	L.Push(result)
	return 1
}

// s3Sync provides a high-level interface for `aws s3 sync`.
// Lua usage: aws.s3.sync({source="...", destination="...", profile="...", delete=true})
func (mod *AWSModule) s3Sync(L *lua.LState) int {
	tbl := L.CheckTable(1)
	source := tbl.RawGetString("source").String()
	destination := tbl.RawGetString("destination").String()
	profile := tbl.RawGetString("profile").String()
	del := lua.LVAsBool(tbl.RawGetString("delete"))

	if source == "" || destination == "" {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("source and destination are required for s3.sync"))
		return 2
	}

	args := []string{"s3", "sync", source, destination}
	if del {
		args = append(args, "--delete")
	}

	// Call the generic exec function
	L.Push(L.NewFunction(mod.awsExec))
	argsTable := L.NewTable()
	for _, arg := range args {
		argsTable.Append(lua.LString(arg))
	}
	L.Push(argsTable)

	optsTable := L.NewTable()
	if profile != "" {
		optsTable.RawSetString("profile", lua.LString(profile))
	}
	L.Push(optsTable)

	L.Call(2, 1)
	result := L.CheckTable(L.GetTop())
	exitCode := int(result.RawGetString("exit_code").(lua.LNumber))

	if exitCode != 0 {
		L.Push(lua.LBool(false))
		L.Push(result.RawGetString("stderr"))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}

// secretsManagerGetSecret retrieves a secret from AWS Secrets Manager.
// Lua usage: aws.secretsmanager.get_secret({secret_id="...", profile="..."})
func (mod *AWSModule) secretsManagerGetSecret(L *lua.LState) int {
	tbl := L.CheckTable(1)
	secretID := tbl.RawGetString("secret_id").String()
	profile := tbl.RawGetString("profile").String()

	if secretID == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("secret_id is required"))
		return 2
	}

	args := []string{"secretsmanager", "get-secret-value", "--secret-id", secretID}

	// Call the generic exec function
	L.Push(L.NewFunction(mod.awsExec))
	argsTable := L.NewTable()
	for _, arg := range args {
		argsTable.Append(lua.LString(arg))
	}
	L.Push(argsTable)

	optsTable := L.NewTable()
	if profile != "" {
		optsTable.RawSetString("profile", lua.LString(profile))
	}
	L.Push(optsTable)

	L.Call(2, 1)
	result := L.CheckTable(L.GetTop())
	exitCode := int(result.RawGetString("exit_code").(lua.LNumber))
	stdout := result.RawGetString("stdout").String()
	stderr := result.RawGetString("stderr").String()

	if exitCode != 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("Failed to get secret: %s", stderr)))
		return 2
	}

	var secretResponse struct {
		SecretString string `json:"SecretString"`
	}

	if err := json.Unmarshal([]byte(stdout), &secretResponse); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("Failed to parse secret JSON: %v", err)))
		return 2
	}

	L.Push(lua.LString(secretResponse.SecretString))
	return 1
}
