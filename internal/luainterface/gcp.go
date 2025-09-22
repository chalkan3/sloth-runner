package luainterface

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const (
	luaGCPClientTypeName         = "gcp_client"
	luaGCPServiceAccountTypeName = "gcp_service_account"
)

// GCPClient represents a client for GCP operations.
type GCPClient struct {
	Project string
}

// GCPServiceAccount represents a GCP Service Account.
type GCPServiceAccount struct {
	Client  *GCPClient
	Name    string
	Email   string
	Project string
}

// --- Constructors ---

// gcp.client({ project = "my-project" }) -> client
func newGCPClient(L *lua.LState) int {
	opts := L.OptTable(1, L.NewTable())
	project := opts.RawGetString("project").String()

	client := &GCPClient{Project: project}
	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPClientTypeName))
	L.Push(ud)
	return 1
}

// --- Helpers ---

func checkGCPClient(L *lua.LState) *GCPClient {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPClient); ok {
		return v
	}
	L.ArgError(1, "gcp client expected")
	return nil
}

func checkGCPServiceAccount(L *lua.LState) *GCPServiceAccount {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPServiceAccount); ok {
		return v
	}
	L.ArgError(1, "gcp service account expected")
	return nil
}

func (c *GCPClient) runGCloudCommand(args ...string) (string, string, error) {
	allArgs := []string{}
	project := c.Project
	// Allow overriding project per command
	for i, arg := range args {
		if arg == "--project" && i+1 < len(args) {
			project = args[i+1]
		}
	}
	if project != "" {
		allArgs = append(allArgs, "--project", project)
	}
	allArgs = append(allArgs, args...)

	cmd := exec.Command("gcloud", allArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// --- Client Methods ---

// client:service_account("my-sa") -> sa
func (c *GCPClient) serviceAccount(L *lua.LState) int {
	name := L.CheckString(2)
	project := L.OptString(3, c.Project)
	if project == "" {
		L.RaiseError("project must be specified either in gcp.client or service_account method")
	}

	sa := &GCPServiceAccount{
		Client:  c,
		Name:    name,
		Project: project,
		Email:   fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, project),
	}
	ud := L.NewUserData()
	ud.Value = sa
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPServiceAccountTypeName))
	L.Push(ud)
	return 1
}

// --- Service Account Methods ---

// sa:create({ display_name = "..." }) -> self
func (sa *GCPServiceAccount) create(L *lua.LState) int {
	opts := L.OptTable(2, L.NewTable())
	displayName := opts.RawGetString("display_name").String()

	args := []string{"iam", "service-accounts", "create", sa.Name}
	if displayName != "" {
		args = append(args, "--display-name", displayName)
	}

	_, stderr, err := sa.Client.runGCloudCommand(args...)
	if err != nil {
		// Ignore "already exists" error to make it idempotent
		if !strings.Contains(stderr, "already exists") {
			L.RaiseError("gcloud iam service-accounts create failed: %s", stderr)
		}
	}

	L.Push(L.Get(1)) // return self
	return 1
}

// sa:add_iam_binding({ member = "...", role = "..." }) -> self
func (sa *GCPServiceAccount) addIAMBinding(L *lua.LState) int {
	opts := L.CheckTable(2)
	member := opts.RawGetString("member").String()
	role := opts.RawGetString("role").String()

	if member == "" || role == "" {
		L.RaiseError("member and role are required for add_iam_binding")
	}

	args := []string{
		"projects", "add-iam-policy-binding", sa.Project,
		"--member", member,
		"--role", role,
	}

	_, stderr, err := sa.Client.runGCloudCommand(args...)
	if err != nil {
		L.RaiseError("gcloud projects add-iam-policy-binding failed: %s", stderr)
	}

	L.Push(L.Get(1)) // return self
	return 1
}

// --- Loaders ---

var gcpClientMethods = map[string]lua.LGFunction{
	"service_account": func(L *lua.LState) int {
		client := checkGCPClient(L)
		return client.serviceAccount(L)
	},
}

var gcpServiceAccountMethods = map[string]lua.LGFunction{
	"create": func(L *lua.LState) int {
		sa := checkGCPServiceAccount(L)
		return sa.create(L)
	},
	"add_iam_binding": func(L *lua.LState) int {
		sa := checkGCPServiceAccount(L)
		return sa.addIAMBinding(L)
	},
}

func GCPLoader(L *lua.LState) int {
	// Register client type
	clientMT := L.NewTypeMetatable(luaGCPClientTypeName)
	L.SetField(clientMT, "__index", L.SetFuncs(L.NewTable(), gcpClientMethods))

	// Register service account type
	saMT := L.NewTypeMetatable(luaGCPServiceAccountTypeName)
	L.SetField(saMT, "__index", L.SetFuncs(L.NewTable(), gcpServiceAccountMethods))

	// Register module
	mod := L.NewTable()
	L.SetField(mod, "client", L.NewFunction(newGCPClient))
	L.Push(mod)
	return 1
}

func OpenGCP(L *lua.LState) {
	L.PreloadModule("gcp", GCPLoader)
}
