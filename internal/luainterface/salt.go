package luainterface

import (
	"bytes"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

const (
	luaSaltClientTypeName = "salt_client"
	luaSaltTargetTypeName = "salt_target"
)

// SaltClient represents a client to a Salt master.
type SaltClient struct {
	ConfigPath string
}

// SaltTarget represents a target for Salt commands.
type SaltTarget struct {
	Client     *SaltClient
	Target     string
	TargetType string
}

// --- Constructors ---

// salt.client({ config = "/path/to/master" }) -> client
func newSaltClient(L *lua.LState) int {
	opts := L.OptTable(1, L.NewTable())
	config := opts.RawGetString("config").String()

	client := &SaltClient{ConfigPath: config}
	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaSaltClientTypeName))
	L.Push(ud)
	return 1
}

// --- Helper ---

func checkSaltClient(L *lua.LState) *SaltClient {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*SaltClient); ok {
		return v
	}
	L.ArgError(1, "salt client expected")
	return nil
}

func checkSaltTarget(L *lua.LState) *SaltTarget {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*SaltTarget); ok {
		return v
	}
	L.ArgError(1, "salt target expected")
	return nil
}

// --- Client Methods ---

// client:target("minion*", "glob") -> target
func (sc *SaltClient) target(L *lua.LState) int {
	tgt := L.CheckString(2)
	tgtType := L.OptString(3, "glob") // Default to glob targeting

	target := &SaltTarget{
		Client:     sc,
		Target:     tgt,
		TargetType: tgtType,
	}
	ud := L.NewUserData()
	ud.Value = target
	L.SetMetatable(ud, L.GetTypeMetatable(luaSaltTargetTypeName))
	L.Push(ud)
	return 1
}

// --- Target Methods ---

// target:cmd("state.apply", "users") -> result
func (st *SaltTarget) cmd(L *lua.LState) int {
	cmdStr := L.CheckString(2)
	var args []string
	for i := 3; i <= L.GetTop(); i++ {
		args = append(args, L.ToString(i))
	}

	fullArgs := []string{"--out=json"}
	if st.Client.ConfigPath != "" {
		fullArgs = append(fullArgs, "--config-dir="+st.Client.ConfigPath)
	}
	fullArgs = append(fullArgs, "-L", st.Target, cmdStr)
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command("salt", fullArgs...)
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

// --- Loaders ---

var saltClientMethods = map[string]lua.LGFunction{
	"target": func(L *lua.LState) int {
		client := checkSaltClient(L)
		return client.target(L)
	},
}

var saltTargetMethods = map[string]lua.LGFunction{
	"cmd": func(L *lua.LState) int {
		target := checkSaltTarget(L)
		return target.cmd(L)
	},
}

func SaltLoader(L *lua.LState) int {
	// Register client type
	clientMT := L.NewTypeMetatable(luaSaltClientTypeName)
	L.SetField(clientMT, "__index", L.SetFuncs(L.NewTable(), saltClientMethods))

	// Register target type
	targetMT := L.NewTypeMetatable(luaSaltTargetTypeName)
	L.SetField(targetMT, "__index", L.SetFuncs(L.NewTable(), saltTargetMethods))

	// Register module
	mod := L.NewTable()
	L.SetField(mod, "client", L.NewFunction(newSaltClient))
	L.Push(mod)
	return 1
}

func OpenSalt(L *lua.LState) {
	L.PreloadModule("salt", SaltLoader)
}