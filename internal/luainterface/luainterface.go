package luainterface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v2"
)

var ExecCommand = exec.Command

func newLuaImportFunction(baseDir string) lua.LGFunction {
	return func(L *lua.LState) int {
		relPath := L.CheckString(1)
		absPath := filepath.Join(baseDir, relPath)
		content, err := ioutil.ReadFile(absPath)
		if err != nil {
			L.RaiseError("cannot read imported file: %s", err.Error())
			return 0
		}
		if err := L.DoString(string(content)); err != nil {
			L.RaiseError("error executing imported file: %s", err.Error())
			return 0
		}
		return 1
	}
}

func OpenImport(L *lua.LState, configFilePath string) {
	baseDir := filepath.Dir(configFilePath)
	L.SetGlobal("import", L.NewFunction(newLuaImportFunction(baseDir)))
}

func GoValueToLua(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		arr := L.NewTable()
		for i, elem := range v {
			arr.RawSetInt(i+1, GoValueToLua(L, elem))
		}
		return arr
	case map[string]interface{}:
		tbl := L.NewTable()
		for key, elem := range v {
			tbl.RawSetString(key, GoValueToLua(L, elem))
		}
		return tbl
	case map[interface{}]interface{}:
		tbl := L.NewTable()
		for key, elem := range v {
			if strKey, ok := key.(string); ok {
				tbl.RawSetString(strKey, GoValueToLua(L, elem))
			} else {
				slog.Warn("Non-string key encountered in YAML map, skipping", "key", key)
			}
		}
		return tbl
	case nil:
		return lua.LNil
	default:
		return lua.LString(fmt.Sprintf("unsupported Go type: %T", v))
	}
}

func LuaToGoValue(L *lua.LState, value lua.LValue) interface{} {
	switch value.Type() {
	case lua.LTBool:
		return lua.LVAsBool(value)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(value))
	case lua.LTString:
		return lua.LVAsString(value)
	case lua.LTTable:
		tbl := value.(*lua.LTable)
		if tbl.Len() > 0 {
			arr := make([]interface{}, 0, tbl.Len())
			for i := 1; i <= tbl.Len(); i++ {
				arr = append(arr, LuaToGoValue(L, tbl.RawGetInt(i)))
			}
			return arr
		} else {
			m := make(map[string]interface{})
			tbl.ForEach(func(key, val lua.LValue) {
				m[key.String()] = LuaToGoValue(L, val)
			})
			return m
		}
	case lua.LTNil:
		return nil
	default:
		return value.String()
	}
}

// --- Data Module ---
func luaDataParseJson(L *lua.LState) int {
	jsonString := L.CheckString(1)
	var goValue interface{}
	err := json.Unmarshal([]byte(jsonString), &goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(GoValueToLua(L, goValue))
	L.Push(lua.LNil)
	return 2
}

func luaDataToJson(L *lua.LState) int {
	luaTable := L.CheckTable(1)
	goValue := LuaToGoValue(L, luaTable)
	jsonBytes, err := json.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(jsonBytes)))
	L.Push(lua.LNil)
	return 2
}

func luaDataParseYaml(L *lua.LState) int {
	yamlString := L.CheckString(1)
	var goValue interface{}
	var mapValue map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlString), &mapValue)
	if err == nil {
		goValue = mapValue
	} else {
		err = yaml.Unmarshal([]byte(yamlString), &goValue)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
	}
	L.Push(GoValueToLua(L, goValue))
	L.Push(lua.LNil)
	return 2
}

func luaDataToYaml(L *lua.LState) int {
	luaTable := L.CheckTable(1)
	goValue := LuaToGoValue(L, luaTable)
	yamlBytes, err := yaml.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(yamlBytes)))
	L.Push(lua.LNil)
	return 2
}

func DataLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"parse_json": luaDataParseJson,
		"to_json":    luaDataToJson,
		"parse_yaml": luaDataParseYaml,
		"to_yaml":    luaDataToYaml,
	})
	L.Push(mod)
	return 1
}
func OpenData(L *lua.LState) {
	L.PreloadModule("data", DataLoader)
	if err := L.DoString(`data = require("data")`); err != nil {
		panic(err)
	}
}

// --- FS Module ---
func luaFsRead(L *lua.LState) int {
	path := L.CheckString(1)
	content, err := os.ReadFile(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(content)))
	L.Push(lua.LNil)
	return 2
}

func luaFsWrite(L *lua.LState) int {
	path := L.CheckString(1)
	content := L.CheckString(2)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func luaFsAppend(L *lua.LState) int {
	path := L.CheckString(1)
	content := L.CheckString(2)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaFsExists(L *lua.LState) int {
	path := L.CheckString(1)
	_, err := os.Stat(path)
	if err == nil {
		L.Push(lua.LBool(true))
		return 1
	}
	if os.IsNotExist(err) {
		L.Push(lua.LBool(false))
		return 1
	}
	L.Push(lua.LBool(false))
	return 1
}

func luaFsMkdir(L *lua.LState) int {
	path := L.CheckString(1)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func luaFsRm(L *lua.LState) int {
	path := L.CheckString(1)
	err := os.Remove(path)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func luaFsRmR(L *lua.LState) int {
	path := L.CheckString(1)
	err := os.RemoveAll(path)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func luaFsLs(L *lua.LState) int {
	path := L.CheckString(1)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	luaTable := L.NewTable()
	for i, file := range files {
		luaTable.RawSetInt(i+1, lua.LString(file.Name()))
	}
	L.Push(luaTable)
	L.Push(lua.LNil)
	return 2
}

func luaFsTmpName(L *lua.LState) int {
	// In Go, creating a temp file/dir is the idiomatic way to get a unique temp name.
	// We create a directory and immediately remove it just to get the name.
	dir, err := ioutil.TempDir("", "sloth-runner-*")
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	os.Remove(dir) // We only want the name, not the directory itself yet.
	L.Push(lua.LString(dir))
	L.Push(lua.LNil)
	return 2
}

func FsLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"read":    luaFsRead,
		"write":   luaFsWrite,
		"append":  luaFsAppend,
		"exists":  luaFsExists,
		"mkdir":   luaFsMkdir,
		"rm":      luaFsRm,
		"rm_r":    luaFsRmR,
		"ls":      luaFsLs,
		"tmpname": luaFsTmpName,
	})
	L.Push(mod)
	return 1
}
func OpenFs(L *lua.LState) {
	L.PreloadModule("fs", FsLoader)
	if err := L.DoString(`fs = require("fs")`); err != nil {
		panic(err)
	}
}

// --- Net Module ---
func luaNetHttpGet(L *lua.LState) int {
	url := L.CheckString(1)

	resp, err := http.Get(url)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LNumber(0))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 4
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LNumber(resp.StatusCode))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 4
	}

	headersTable := L.NewTable()
	for name, values := range resp.Header {
		headerValues := L.NewTable()
		for i, val := range values {
			headerValues.RawSetInt(i+1, lua.LString(val))
		}
		headersTable.RawSetString(name, headerValues)
	}

	L.Push(lua.LString(string(bodyBytes)))
	L.Push(lua.LNumber(resp.StatusCode))
	L.Push(headersTable)
	L.Push(lua.LNil) // No error
	return 4
}

func luaNetHttpPost(L *lua.LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)
	headersTable := L.OptTable(3, L.NewTable()) // Optional headers table

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LNumber(0))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 4
	}

	headersTable.ForEach(func(key, value lua.LValue) {
		req.Header.Set(key.String(), value.String())
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LNumber(0))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 4
	}
	defer resp.Body.Close()

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LNumber(resp.StatusCode))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 4
	}

	respHeadersTable := L.NewTable()
	for name, values := range resp.Header {
		headerValues := L.NewTable()
		for i, val := range values {
			headerValues.RawSetInt(i+1, lua.LString(val))
		}
		respHeadersTable.RawSetString(name, headerValues)
	}

	L.Push(lua.LString(string(respBodyBytes)))
	L.Push(lua.LNumber(resp.StatusCode))
	L.Push(respHeadersTable)
	L.Push(lua.LNil) // No error
	return 4
}

func luaNetDownload(L *lua.LState) int {
	url := L.CheckString(1)
	destinationPath := L.CheckString(2)

	resp, err := http.Get(url)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		L.Push(lua.LString(fmt.Sprintf("failed to download file: status code %d", resp.StatusCode)))
		return 1
	}

	out, err := os.Create(destinationPath)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func NetLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"http_get":  luaNetHttpGet,
		"http_post": luaNetHttpPost,
		"download":  luaNetDownload,
	})
	L.Push(mod)
	return 1
}
func OpenNet(L *lua.LState) {
	L.PreloadModule("net", NetLoader)
	if err := L.DoString(`net = require("net")`); err != nil {
		panic(err)
	}
}

// --- Exec Module ---
func luaExecRun(L *lua.LState) int {
	commandStr := L.CheckString(1)
	opts := L.OptTable(2, L.NewTable())

	ctx := L.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	slog.Debug("executing command", "source", "lua", "command", commandStr)

	cmd := exec.CommandContext(ctx, "bash", "-c", commandStr)

	// Set workdir from options
	if workdir := opts.RawGetString("workdir"); workdir.Type() == lua.LTString {
		cmd.Dir = workdir.String()
	}

	// Set environment variables from options
	cmd.Env = os.Environ() // Inherit current environment
	if envTbl := opts.RawGetString("env"); envTbl.Type() == lua.LTTable {
		envTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key.String(), value.String()))
		})
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	success := err == nil

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if stdoutStr != "" {
		slog.Info(stdoutStr, "source", "lua", "stream", "stdout")
	}
	if stderrStr != "" {
		slog.Warn(stderrStr, "source", "lua", "stream", "stderr")
	}

	result := L.NewTable()
	result.RawSetString("success", lua.LBool(success))
	result.RawSetString("stdout", lua.LString(stdoutStr))
	result.RawSetString("stderr", lua.LString(stderrStr))
	L.Push(result)
	return 1
}

func ExecLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"run": luaExecRun,
	})
	L.Push(mod)
	return 1
}
func OpenExec(L *lua.LState) {
	L.PreloadModule("exec", ExecLoader)
	if err := L.DoString(`exec = require("exec")`); err != nil {
		panic(err)
	}
}

// --- Log Module ---
func luaLogInfo(L *lua.LState) int {
	message := L.CheckString(1)
	slog.Info(message, "source", "lua")
	return 0
}

func luaLogWarn(L *lua.LState) int {
	message := L.CheckString(1)
	slog.Warn(message, "source", "lua")
	return 0
}

func luaLogError(L *lua.LState) int {
	message := L.CheckString(1)
	slog.Error(message, "source", "lua")
	return 0
}

func luaLogDebug(L *lua.LState) int {
	message := L.CheckString(1)
	slog.Debug(message, "source", "lua")
	return 0
}

func LogLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"info":  luaLogInfo,
		"warn":  luaLogWarn,
		"error": luaLogError,
		"debug": luaLogDebug,
	})
	L.Push(mod)
	return 1
}
func OpenLog(L *lua.LState) {
	L.PreloadModule("log", LogLoader)
	if err := L.DoString(`log = require("log")`); err != nil {
		panic(err)
	}
}

// OpenAll preloads all available sloth-runner modules into the Lua state.
func OpenAll(L *lua.LState) {
	// RegisterSharedSessionType(L)
	OpenExec(L)
	OpenFs(L)
	OpenNet(L)
	OpenData(L)
	OpenLog(L)
	OpenSalt(L)
	OpenPulumi(L)
	OpenGit(L)
	OpenGCP(L)
	OpenPython(L)
	OpenNotifications(L)
	OpenAWS(L)
	OpenDigitalOcean(L)
	OpenAzure(L)
}

func OpenAzure(L *lua.LState) {
	mod := NewAzureModule()
	L.PreloadModule("azure", mod.Loader)
	if err := L.DoString(`azure = require("azure")`); err != nil {
		panic(err)
	}
}

func OpenDigitalOcean(L *lua.LState) {
	mod := NewDigitalOceanModule()
	L.PreloadModule("digitalocean", mod.Loader)
	if err := L.DoString(`digitalocean = require("digitalocean")`); err != nil {
		panic(err)
	}
}

func OpenNotifications(L *lua.LState) {
	mod := NewNotificationsModule()
	L.PreloadModule("notifications", mod.Loader)
	if err := L.DoString(`notifications = require("notifications")`); err != nil {
		panic(err)
	}
}

func OpenAWS(L *lua.LState) {
	mod := NewAWSModule()
	L.PreloadModule("aws", mod.Loader)
	if err := L.DoString(`aws = require("aws")`); err != nil {
		panic(err)
	}
}

// --- Parallel Module ---
func newParallelFunction(tr types.TaskRunner) lua.LGFunction {
	return func(L *lua.LState) int {
		tasksTable := L.CheckTable(1)
		inputTable := L.OptTable(2, L.NewTable())
		var tasksToRun []*types.Task
		tasksTable.ForEach(func(_, taskValue lua.LValue) {
			if taskValue.Type() == lua.LTTable {
				task := parseLuaTask(taskValue.(*lua.LTable))
				tasksToRun = append(tasksToRun, &task)
			} else {
				L.RaiseError("invalid item in parallel task list: expected a table, got %s", taskValue.Type().String())
			}
		})
		if len(tasksToRun) == 0 {
			L.Push(L.NewTable())
			L.Push(lua.LNil)
			return 2
		}
		results, err := tr.RunTasksParallel(tasksToRun, inputTable)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("parallel execution failed: %v", err)))
			return 2
		}
		resultsTable := L.NewTable()
		for _, res := range results {
			taskResultTable := L.NewTable()
			taskResultTable.RawSetString("name", lua.LString(res.Name))
			taskResultTable.RawSetString("status", lua.LString(res.Status))
			taskResultTable.RawSetString("duration", lua.LString(res.Duration.String()))
			if res.Error != nil {
				taskResultTable.RawSetString("error", lua.LString(res.Error.Error()))
			}
			resultsTable.Append(taskResultTable)
		}
		L.Push(resultsTable)
		L.Push(lua.LNil)
		return 2
	}
}

func OpenParallel(L *lua.LState, tr types.TaskRunner) {
	L.PreloadModule("parallel", func(L *lua.LState) int {
		mod := L.NewFunction(newParallelFunction(tr))
		L.Push(mod)
		return 1
	})
}

// --- Core Execution Logic ---
func ExecuteLuaFunction(L *lua.LState, fn *lua.LFunction, params map[string]string, secondArg lua.LValue, nRet int, ctx context.Context, args ...lua.LValue) (bool, string, *lua.LTable, error) {
	if ctx != nil {
		L.SetContext(ctx)
	}
	L.Push(fn)
	luaParams := L.NewTable()
	for k, v := range params {
		luaParams.RawSetString(k, lua.LString(v))
	}
	L.Push(luaParams)
	numArgs := 1
	if secondArg != nil {
		L.Push(secondArg)
		numArgs = 2
	}
	// Push additional args
	for _, arg := range args {
		L.Push(arg)
		numArgs++
	}

	if err := L.PCall(numArgs, lua.MultRet, nil); err != nil {
		return false, "", nil, fmt.Errorf("error executing Lua function: %w", err)
	}
	top := L.GetTop()
	var success bool
	var message string
	var outputTable *lua.LTable
	if top >= 1 {
		if L.Get(1).Type() == lua.LTBool {
			success = lua.LVAsBool(L.Get(1))
		} else {
			success = false
			message = fmt.Sprintf("unexpected first return type from Lua: %s", L.Get(1).Type().String())
		}
	}
	if top >= 2 {
		if L.Get(2).Type() == lua.LTString {
			message = lua.LVAsString(L.Get(2))
		}
	}
	if top >= 3 {
		if L.Get(3).Type() == lua.LTTable {
			outputTable = L.Get(3).(*lua.LTable)
		}
	}
	L.SetTop(0)
	return success, message, outputTable, nil
}

func LoadTaskDefinitions(L *lua.LState, luaScriptContent string, configFilePath string) (map[string]types.TaskGroup, error) {
	if err := L.DoString(luaScriptContent); err != nil {
		return nil, fmt.Errorf("error loading Lua script content: %w", err)
	}
	globalTaskDefs := L.GetGlobal("TaskDefinitions")
	if globalTaskDefs.Type() != lua.LTTable {
		return nil, fmt.Errorf("expected 'TaskDefinitions' to be a table, got %s", globalTaskDefs.Type().String())
	}
	loadedTaskGroups := make(map[string]types.TaskGroup)
	globalTaskDefs.(*lua.LTable).ForEach(func(groupKey, groupValue lua.LValue) {
		groupName := groupKey.String()
		if groupValue.Type() != lua.LTTable {
			slog.Warn("Expected group to be a table, skipping", "group", groupName)
			return
		}
		groupTable := groupValue.(*lua.LTable)
		description := groupTable.RawGetString("description").String()
		workdir := groupTable.RawGetString("workdir").String() // Add this line

		// Parse workdir lifecycle fields
		createWorkdir := lua.LVAsBool(groupTable.RawGetString("create_workdir_before_run"))
		cleanWorkdirFunc, _ := groupTable.RawGetString("clean_workdir_after_run").(*lua.LFunction)

		var tasks []types.Task
		luaTasks := groupTable.RawGetString("tasks")
		if luaTasks.Type() == lua.LTTable {
			luaTasks.(*lua.LTable).ForEach(func(taskKey, taskValue lua.LValue) {
				if taskValue.Type() != lua.LTTable {
					slog.Warn("Expected task entry to be a table, skipping", "group", groupName)
					return
				}
				taskTable := taskValue.(*lua.LTable)
				var finalTask types.Task
				usesField := taskTable.RawGetString("uses")
				if usesField.Type() == lua.LTTable {
					baseTaskTable := usesField.(*lua.LTable)
					baseTask := parseLuaTask(baseTaskTable)
					localOverrides := parseLuaTask(taskTable)
					finalTask = baseTask
					if localOverrides.Description != "" {
						finalTask.Description = localOverrides.Description
					}
					if localOverrides.CommandFunc != nil {
						finalTask.CommandFunc = localOverrides.CommandFunc
					}
					if localOverrides.CommandStr != "" {
						finalTask.CommandStr = localOverrides.CommandStr
					}
					// ... (merge other fields)
					finalTask.Name = taskKey.String()
				} else {
					finalTask = parseLuaTask(taskTable)
				}
				tasks = append(tasks, finalTask)
			})
		}
		loadedTaskGroups[groupName] = types.TaskGroup{
			Description:              description,
			Tasks:                    tasks,
			Workdir:                  workdir, // Add this line
			CreateWorkdirBeforeRun:   createWorkdir,
			CleanWorkdirAfterRunFunc: cleanWorkdirFunc,
		}
	})
	return loadedTaskGroups, nil
}

func parseLuaTask(taskTable *lua.LTable) types.Task {
	name := taskTable.RawGetString("name").String()
	desc := taskTable.RawGetString("description").String()
	var cmdFunc *lua.LFunction
	var cmdStr string
	luaCommand := taskTable.RawGetString("command")
	if luaCommand.Type() == lua.LTString {
		cmdStr = luaCommand.String()
	} else if luaCommand.Type() == lua.LTFunction {
		cmdFunc = luaCommand.(*lua.LFunction)
	}

	// Parse params
	params := make(map[string]string)
	luaParams := taskTable.RawGetString("params")
	if luaParams.Type() == lua.LTTable {
		luaParams.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			params[k.String()] = v.String()
		})
	}

	// Parse depends_on
	var dependsOn []string
	luaDependsOn := taskTable.RawGetString("depends_on")
	if luaDependsOn.Type() == lua.LTString {
		dependsOn = []string{luaDependsOn.String()}
	} else if luaDependsOn.Type() == lua.LTTable {
		luaDependsOn.(*lua.LTable).ForEach(func(_, v lua.LValue) {
			dependsOn = append(dependsOn, v.String())
		})
	}

	// Parse next_if_fail
	var nextIfFail []string
	luaNextIfFail := taskTable.RawGetString("next_if_fail")
	if luaNextIfFail.Type() == lua.LTString {
		nextIfFail = []string{luaNextIfFail.String()}
	} else if luaNextIfFail.Type() == lua.LTTable {
		luaNextIfFail.(*lua.LTable).ForEach(func(_, v lua.LValue) {
			nextIfFail = append(nextIfFail, v.String())
		})
	}

	// Parse retries
	retries := 0
	luaRetries := taskTable.RawGetString("retries")
	if luaRetries.Type() == lua.LTNumber {
		retries = int(luaRetries.(lua.LNumber))
	}

	// Parse timeout
	timeout := ""
	luaTimeout := taskTable.RawGetString("timeout")
	if luaTimeout.Type() == lua.LTString {
		timeout = luaTimeout.String()
	}

	// Parse async
	async := false
	luaAsync := taskTable.RawGetString("async")
	if luaAsync.Type() == lua.LTBool {
		async = lua.LVAsBool(luaAsync)
	}

	// Parse pre_exec and post_exec
	var preExec, postExec *lua.LFunction
	luaPreExec := taskTable.RawGetString("pre_exec")
	if luaPreExec.Type() == lua.LTFunction {
		preExec = luaPreExec.(*lua.LFunction)
	}
	luaPostExec := taskTable.RawGetString("post_exec")
	if luaPostExec.Type() == lua.LTFunction {
		postExec = luaPostExec.(*lua.LFunction)
	}

	return types.Task{
		Name:        name,
		Description: desc,
		CommandFunc: cmdFunc,
		CommandStr:  cmdStr,
		Params:      params,
		DependsOn:   dependsOn,
		NextIfFail:  nextIfFail,
		Retries:     retries,
		Timeout:     timeout,
		Async:       async,
		PreExec:     preExec,
		PostExec:    postExec,
	}
}

func LuaTableToGoMap(L *lua.LState, table *lua.LTable) map[string]interface{} {
	result := make(map[string]interface{})
	table.ForEach(func(key, value lua.LValue) {
		k := key.String()
		switch value.Type() {
		case lua.LTBool:
			result[k] = lua.LVAsBool(value)
		case lua.LTNumber:
			result[k] = lua.LVAsNumber(value)
		case lua.LTString:
			result[k] = lua.LVAsString(value)
		case lua.LTTable:
			result[k] = LuaTableToGoMap(L, value.(*lua.LTable))
		default:
			result[k] = value.String()
		}
	})
	return result
}

// CopyTable performs a deep copy of a table from one Lua state to another.
func CopyTable(src *lua.LTable, destL *lua.LState) *lua.LTable {
	destT := destL.NewTable()
	src.ForEach(func(key, value lua.LValue) {
		destKey := CopyValue(key, destL)
		destValue := CopyValue(value, destL)
		destL.SetTable(destT, destKey, destValue)
	})
	return destT
}

// CopyValue copies a Lua value from one state to another.
func CopyValue(value lua.LValue, destL *lua.LState) lua.LValue {
	switch value.Type() {
	case lua.LTBool:
		return lua.LBool(lua.LVAsBool(value))
	case lua.LTNumber:
		return lua.LNumber(lua.LVAsNumber(value))
	case lua.LTString:
		return lua.LString(lua.LVAsString(value))
	case lua.LTTable:
		return CopyTable(value.(*lua.LTable), destL)
	case lua.LTUserData:
		// Userdata cannot be safely copied directly. We copy the underlying value if possible.
		srcUD := value.(*lua.LUserData)
		destUD := destL.NewUserData()
		destUD.Value = srcUD.Value // This is a shallow copy of the value
		return destUD
	default:
		// For other types (functions, etc.), we return nil as they cannot be copied.
		return lua.LNil
	}
}