package luainterface

import (
	"bytes"
	"encoding/json"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

const (
	luaAWSClientTypeName = "aws_client"
	luaAWSS3TypeName     = "aws_s3"
)

// AWSClient represents a client for AWS operations.
type AWSClient struct {
	Profile string
}

// AWSS3 represents the S3 service.
type AWSS3 struct {
	Client *AWSClient
}

// --- Constructor ---

// aws.client({ profile = "my-profile" }) -> client
func newAWSClient(L *lua.LState) int {
	opts := L.OptTable(1, L.NewTable())
	profile := opts.RawGetString("profile").String()

	client := &AWSClient{Profile: profile}
	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaAWSClientTypeName))
	L.Push(ud)
	return 1
}

// --- Helpers ---

func checkAWSClient(L *lua.LState) *AWSClient {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*AWSClient); ok {
		return v
	}
	L.ArgError(1, "aws client expected")
	return nil
}

func checkAWSS3(L *lua.LState) *AWSS3 {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*AWSS3); ok {
		return v
	}
	L.ArgError(1, "aws s3 object expected")
	return nil
}

func (c *AWSClient) runAWSCommand(args ...string) (string, string, error) {
	var cmd *exec.Cmd
	if c.Profile != "" {
		cmdArgs := append([]string{"exec", c.Profile, "--", "aws"}, args...)
		cmd = exec.Command("aws-vault", cmdArgs...)
	} else {
		cmd = exec.Command("aws", args...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// --- Client Methods ---

// client:s3() -> s3_obj
func (c *AWSClient) s3(L *lua.LState) int {
	s3 := &AWSS3{Client: c}
	ud := L.NewUserData()
	ud.Value = s3
	L.SetMetatable(ud, L.GetTypeMetatable(luaAWSS3TypeName))
	L.Push(ud)
	return 1
}

// client:get_secret("secret-id") -> secret_string
func (c *AWSClient) getSecret(L *lua.LState) int {
	secretID := L.CheckString(2)
	args := []string{"secretsmanager", "get-secret-value", "--secret-id", secretID}

	stdout, stderr, err := c.runAWSCommand(args...)
	if err != nil {
		L.RaiseError("aws secretsmanager get-secret-value failed: %s", stderr)
	}

	var secretResponse struct {
		SecretString string `json:"SecretString"`
	}
	if err := json.Unmarshal([]byte(stdout), &secretResponse); err != nil {
		L.RaiseError("failed to parse secret JSON: %v", err)
	}

	L.Push(lua.LString(secretResponse.SecretString))
	return 1
}

// --- S3 Methods ---

// s3:sync({ from = "...", to = "...", delete = false }) -> self
func (s3 *AWSS3) sync(L *lua.LState) int {
	opts := L.CheckTable(2)
	from := opts.RawGetString("from").String()
	to := opts.RawGetString("to").String()
	del := lua.LVAsBool(opts.RawGetString("delete"))

	if from == "" || to == "" {
		L.RaiseError("from and to are required for s3:sync")
	}

	args := []string{"s3", "sync", from, to}
	if del {
		args = append(args, "--delete")
	}

	_, stderr, err := s3.Client.runAWSCommand(args...)
	if err != nil {
		L.RaiseError("aws s3 sync failed: %s", stderr)
	}

	L.Push(L.Get(1)) // return self
	return 1
}

// --- Loaders ---

var awsClientMethods = map[string]lua.LGFunction{
	"s3": func(L *lua.LState) int {
		client := checkAWSClient(L)
		return client.s3(L)
	},
	"get_secret": func(L *lua.LState) int {
		client := checkAWSClient(L)
		return client.getSecret(L)
	},
}

var awsS3Methods = map[string]lua.LGFunction{
	"sync": func(L *lua.LState) int {
		s3 := checkAWSS3(L)
		return s3.sync(L)
	},
}

func AWSLoader(L *lua.LState) int {
	// Register client type
	clientMT := L.NewTypeMetatable(luaAWSClientTypeName)
	L.SetField(clientMT, "__index", L.SetFuncs(L.NewTable(), awsClientMethods))

	// Register S3 type
	s3MT := L.NewTypeMetatable(luaAWSS3TypeName)
	L.SetField(s3MT, "__index", L.SetFuncs(L.NewTable(), awsS3Methods))

	// Register module
	mod := L.NewTable()
	L.SetField(mod, "client", L.NewFunction(newAWSClient))
	L.Push(mod)
	return 1
}

func OpenAWS(L *lua.LState) {
	L.PreloadModule("aws", AWSLoader)
}