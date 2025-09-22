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
	luaGCPInstancesTypeName      = "gcp_instances"
	luaGCPBucketsTypeName        = "gcp_buckets"
	luaGCPSqlTypeName            = "gcp_sql"
	luaGCPSqlInstancesTypeName   = "gcp_sql_instances"
	luaGCPGkeTypeName            = "gcp_gke"
	luaGCPGkeClustersTypeName    = "gcp_gke_clusters"
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

// GCPCompute represents the compute service client.
type GCPCompute struct {
	Client *GCPClient
	Zone   string
}

// GCPInstances represents the instances service client.
type GCPInstances struct {
	Compute *GCPCompute
}

// GCPStorage represents the storage service client.
type GCPStorage struct {
	Client *GCPClient
}

// GCPBuckets represents the buckets service client.
type GCPBuckets struct {
	Storage *GCPStorage
}

// GCPSql represents the sql service client.
type GCPSql struct {
	Client *GCPClient
}

// GCPSqlInstances represents the sql instances service client.
type GCPSqlInstances struct {
	Sql *GCPSql
}

// GCPGke represents the gke service client.
type GCPGke struct {
	Client *GCPClient
}

// GCPGkeClusters represents the gke clusters service client.
type GCPGkeClusters struct {
	Gke *GCPGke
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

func checkGCPCompute(L *lua.LState) *GCPCompute {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPCompute); ok {
		return v
	}
	L.ArgError(1, "gcp compute expected")
	return nil
}

func checkGCPInstances(L *lua.LState) *GCPInstances {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPInstances); ok {
		return v
	}
	L.ArgError(1, "gcp instances expected")
	return nil
}

func checkGCPStorage(L *lua.LState) *GCPStorage {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPStorage); ok {
		return v
	}
	L.ArgError(1, "gcp storage expected")
	return nil
}

func checkGCPBuckets(L *lua.LState) *GCPBuckets {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPBuckets); ok {
		return v
	}
	L.ArgError(1, "gcp buckets expected")
	return nil
}

func checkGCPSql(L *lua.LState) *GCPSql {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPSql); ok {
		return v
	}
	L.ArgError(1, "gcp sql expected")
	return nil
}

func checkGCPSqlInstances(L *lua.LState) *GCPSqlInstances {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPSqlInstances); ok {
		return v
	}
	L.ArgError(1, "gcp sql instances expected")
	return nil
}

func checkGCPGke(L *lua.LState) *GCPGke {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPGke); ok {
		return v
	}
	L.ArgError(1, "gcp gke expected")
	return nil
}

func checkGCPGkeClusters(L *lua.LState) *GCPGkeClusters {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*GCPGkeClusters); ok {
		return v
	}
	L.ArgError(1, "gcp gke clusters expected")
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

// client:compute({ zone = "..." }) -> compute
func (c *GCPClient) compute(L *lua.LState) int {
	opts := L.OptTable(2, L.NewTable())
	zone := opts.RawGetString("zone").String()

	compute := &GCPCompute{
		Client: c,
		Zone:   zone,
	}
	ud := L.NewUserData()
	ud.Value = compute
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPComputeTypeName))
	L.Push(ud)
	return 1
}

// client:storage() -> storage
func (c *GCPClient) storage(L *lua.LState) int {
	storage := &GCPStorage{
		Client: c,
	}
	ud := L.NewUserData()
	ud.Value = storage
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPStorageTypeName))
	L.Push(ud)
	return 1
}

// client:sql() -> sql
func (c *GCPClient) sql(L *lua.LState) int {
	sql := &GCPSql{
		Client: c,
	}
	ud := L.NewUserData()
	ud.Value = sql
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPSqlTypeName))
	L.Push(ud)
	return 1
}

// client:gke() -> gke
func (c *GCPClient) gke(L *lua.LState) int {
	gke := &GCPGke{
		Client: c,
	}
	ud := L.NewUserData()
	ud.Value = gke
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPGkeTypeName))
	L.Push(ud)
	return 1
}

// --- Compute Methods ---

// compute:instances() -> instances
func (c *GCPCompute) instances(L *lua.LState) int {
	instances := &GCPInstances{
		Compute: c,
	}
	ud := L.NewUserData()
	ud.Value = instances
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPInstancesTypeName))
	L.Push(ud)
	return 1
}

// --- Instances Methods ---

// instances:list() -> { success, stdout, stderr }
func (i *GCPInstances) list(L *lua.LState) int {
	args := []string{"compute", "instances", "list", "--format=json"}
	if i.Compute.Zone != "" {
		args = append(args, "--zones", i.Compute.Zone)
	}

	stdout, stderr, err := i.Compute.Client.runGCloudCommand(args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(err == nil))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// --- Storage Methods ---

// storage:buckets() -> buckets
func (s *GCPStorage) buckets(L *lua.LState) int {
	buckets := &GCPBuckets{
		Storage: s,
	}
	ud := L.NewUserData()
	ud.Value = buckets
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPBucketsTypeName))
	L.Push(ud)
	return 1
}

// --- Buckets Methods ---

// buckets:list() -> { success, stdout, stderr }
func (b *GCPBuckets) list(L *lua.LState) int {
	args := []string{"storage", "buckets", "list", "--format=json"}

	stdout, stderr, err := b.Storage.Client.runGCloudCommand(args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(err == nil))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// --- Sql Methods ---

// sql:instances() -> instances
func (s *GCPSql) instances(L *lua.LState) int {
	instances := &GCPSqlInstances{
		Sql: s,
	}
	ud := L.NewUserData()
	ud.Value = instances
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPSqlInstancesTypeName))
	L.Push(ud)
	return 1
}

// --- Sql Instances Methods ---

// instances:list() -> { success, stdout, stderr }
func (i *GCPSqlInstances) list(L *lua.LState) int {
	args := []string{"sql", "instances", "list", "--format=json"}

	stdout, stderr, err := i.Sql.Client.runGCloudCommand(args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(err == nil))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
	return 1
}

// --- Gke Methods ---

// gke:clusters() -> clusters
func (g *GCPGke) clusters(L *lua.LState) int {
	clusters := &GCPGkeClusters{
		Gke: g,
	}
	ud := L.NewUserData()
	ud.Value = clusters
	L.SetMetatable(ud, L.GetTypeMetatable(luaGCPGkeClustersTypeName))
	L.Push(ud)
	return 1
}

// --- Gke Clusters Methods ---

// clusters:list() -> { success, stdout, stderr }
func (c *GCPGkeClusters) list(L *lua.LState) int {
	args := []string{"container", "clusters", "list", "--format=json"}

	stdout, stderr, err := c.Gke.Client.runGCloudCommand(args...)

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(err == nil))
	result.RawSetString("stdout", lua.LString(stdout))
	result.RawSetString("stderr", lua.LString(stderr))
	L.Push(result)
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
	"compute": func(L *lua.LState) int {
		client := checkGCPClient(L)
		return client.compute(L)
	},
	"storage": func(L *lua.LState) int {
		client := checkGCPClient(L)
		return client.storage(L)
	},
	"sql": func(L *lua.LState) int {
		client := checkGCPClient(L)
		return client.sql(L)
	},
	"gke": func(L *lua.LState) int {
		client := checkGCPClient(L)
		return client.gke(L)
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

var gcpComputeMethods = map[string]lua.LGFunction{
	"instances": func(L *lua.LState) int {
		compute := checkGCPCompute(L)
		return compute.instances(L)
	},
}

var gcpInstancesMethods = map[string]lua.LGFunction{
	"list": func(L *lua.LState) int {
		instances := checkGCPInstances(L)
		return instances.list(L)
	},
}

var gcpStorageMethods = map[string]lua.LGFunction{
	"buckets": func(L *lua.LState) int {
		storage := checkGCPStorage(L)
		return storage.buckets(L)
	},
}

var gcpBucketsMethods = map[string]lua.LGFunction{
	"list": func(L *lua.LState) int {
		buckets := checkGCPBuckets(L)
		return buckets.list(L)
	},
}

var gcpSqlMethods = map[string]lua.LGFunction{
	"instances": func(L *lua.LState) int {
		sql := checkGCPSql(L)
		return sql.instances(L)
	},
}

var gcpSqlInstancesMethods = map[string]lua.LGFunction{
	"list": func(L *lua.LState) int {
		instances := checkGCPSqlInstances(L)
		return instances.list(L)
	},
}

var gcpGkeMethods = map[string]lua.LGFunction{
	"clusters": func(L *lua.LState) int {
		gke := checkGCPGke(L)
		return gke.clusters(L)
	},
}

var gcpGkeClustersMethods = map[string]lua.LGFunction{
	"list": func(L *lua.LState) int {
		clusters := checkGCPGkeClusters(L)
		return clusters.list(L)
	},
}

func GCPLoader(L *lua.LState) int {
	// Register client type
	clientMT := L.NewTypeMetatable(luaGCPClientTypeName)
	L.SetField(clientMT, "__index", L.SetFuncs(L.NewTable(), gcpClientMethods))

	// Register service account type
	saMT := L.NewTypeMetatable(luaGCPServiceAccountTypeName)
	L.SetField(saMT, "__index", L.SetFuncs(L.NewTable(), gcpServiceAccountMethods))

	// Register compute type
	computeMT := L.NewTypeMetatable(luaGCPComputeTypeName)
	L.SetField(computeMT, "__index", L.SetFuncs(L.NewTable(), gcpComputeMethods))

	// Register instances type
	instancesMT := L.NewTypeMetatable(luaGCPInstancesTypeName)
	L.SetField(instancesMT, "__index", L.SetFuncs(L.NewTable(), gcpInstancesMethods))

	// Register storage type
	storageMT := L.NewTypeMetatable(luaGCPStorageTypeName)
	L.SetField(storageMT, "__index", L.SetFuncs(L.NewTable(), gcpStorageMethods))

	// Register buckets type
	bucketsMT := L.NewTypeMetatable(luaGCPBucketsTypeName)
	L.SetField(bucketsMT, "__index", L.SetFuncs(L.NewTable(), gcpBucketsMethods))

	// Register sql type
	sqlMT := L.NewTypeMetatable(luaGCPSqlTypeName)
	L.SetField(sqlMT, "__index", L.SetFuncs(L.NewTable(), gcpSqlMethods))

	// Register sql instances type
	sqlInstancesMT := L.NewTypeMetatable(luaGCPSqlInstancesTypeName)
	L.SetField(sqlInstancesMT, "__index", L.SetFuncs(L.NewTable(), gcpSqlInstancesMethods))

	// Register gke type
	gkeMT := L.NewTypeMetatable(luaGCPGkeTypeName)
	L.SetField(gkeMT, "__index", L.SetFuncs(L.NewTable(), gcpGkeMethods))

	// Register gke clusters type
	gkeClustersMT := L.NewTypeMetatable(luaGCPGkeClustersTypeName)
	L.SetField(gkeClustersMT, "__index", L.SetFuncs(L.NewTable(), gcpGkeClustersMethods))

	// Register module
	mod := L.NewTable()
	L.SetField(mod, "client", L.NewFunction(newGCPClient))
	L.Push(mod)
	return 1
}

func OpenGCP(L *lua.LState) {
	L.PreloadModule("gcp", GCPLoader)
}
