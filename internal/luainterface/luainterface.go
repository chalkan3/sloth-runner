package luainterface

import (
	"bytes" // Added for command output
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec" // Added for executing shell commands
	"strings"

	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v2"

	"sloth-runner/internal/types"
)

// GoValueToLua converts a Go interface{} value to a Lua LValue.
func GoValueToLua(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case bool:
		return lua.LBool(v)
	case float64: // JSON numbers are float64 in Go
		return lua.LNumber(v)
	case int: // Added: Handle Go int type
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}: // JSON array
		arr := L.NewTable()
		for i, elem := range v {
			arr.RawSetInt(i+1, GoValueToLua(L, elem))
		}
		return arr
	case map[string]interface{}: // JSON object
		tbl := L.NewTable()
		for key, elem := range v {
			tbl.RawSetString(key, GoValueToLua(L, elem))
		}
		return tbl
	case map[interface{}]interface{}: // Handle YAML's default map type
		tbl := L.NewTable()
		for key, elem := range v {
			if strKey, ok := key.(string); ok { // Ensure key is a string
				tbl.RawSetString(strKey, GoValueToLua(L, elem))
			} else {
				// Handle non-string keys if necessary, or log a warning
				log.Printf("Warning: Non-string key encountered in YAML map: %v", key)
			}
		}
		return tbl
	case nil:
		return lua.LNil
	default:
		// Fallback for unsupported types, convert to string or handle as error
		return lua.LString(fmt.Sprintf("unsupported Go type: %T", v))
	}
}

// LuaToGoValue converts a Lua LValue to a Go interface{} value.
func LuaToGoValue(L *lua.LState, value lua.LValue) interface{} {
	switch value.Type() {
	case lua.LTBool:
		return lua.LVAsBool(value)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(value)) // Convert to float64 for JSON compatibility
	case lua.LTString:
		return lua.LVAsString(value)
	case lua.LTTable:
		tbl := value.(*lua.LTable)
		if tbl.Len() > 0 { // Check if it's an array-like table
			arr := make([]interface{}, 0, tbl.Len())
			for i := 1; i <= tbl.Len(); i++ {
				arr = append(arr, LuaToGoValue(L, tbl.RawGetInt(i)))
			}
			return arr
		} else { // Otherwise, it's a map-like table
			m := make(map[string]interface{})
			tbl.ForEach(func(key, val lua.LValue) {
				m[key.String()] = LuaToGoValue(L, val)
			})
			return m
		}
	case lua.LTNil:
		return nil
	default:
		return value.String() // Fallback for other types
	}
}

// luaDataParseJson parses a JSON string into a Lua table.
// It takes one argument: json_string (string).
// It returns (lua_table table, err string).
func luaDataParseJson(L *lua.LState) int {
	jsonString := L.CheckString(1)

	var goValue interface{}
	err := json.Unmarshal([]byte(jsonString), &goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	luaValue := GoValueToLua(L, goValue)
	L.Push(luaValue)
	L.Push(lua.LNil)
	return 2
}

// luaDataToJson converts a Lua table to a JSON string.
// It takes one argument: lua_table (table).
// It returns (json_string string, err string).
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

// luaDataParseYaml parses a YAML string into a Lua table.
// It takes one argument: yaml_string (string).
// It returns (lua_table table, err string).
func luaDataParseYaml(L *lua.LState) int {
	yamlString := L.CheckString(1)

	var goValue interface{}
	// Attempt to unmarshal into a map first, as most task configurations will be maps
	var mapValue map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlString), &mapValue)
	if err == nil {
		goValue = mapValue
	} else {
		// If it's not a map, try unmarshaling into a generic interface{}
		err = yaml.Unmarshal([]byte(yamlString), &goValue)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
	}

	luaValue := GoValueToLua(L, goValue)
	L.Push(luaValue)
	L.Push(lua.LNil)
	return 2
}

// luaDataToYaml converts a Lua table to a YAML string.
// It takes one argument: lua_table (table).
// It returns (yaml_string string, err string).
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

// OpenData opens the 'data' library to the Lua state.
func OpenData(L *lua.LState) {
	// Create a new table for the 'data' module
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"parse_json": luaDataParseJson,
		"to_json":    luaDataToJson,
		"parse_yaml": luaDataParseYaml,
		"to_yaml":    luaDataToYaml,
	})
	// Set the 'data' module in the global table
	L.SetGlobal("data", mod)
}

// luaFsRead reads the content of a file.
// It takes one argument: path (string).
// It returns (content string, err string).
func luaFsRead(L *lua.LState) int {
	path := L.CheckString(1)
	content, err := os.ReadFile(path) // Using os.ReadFile for newer Go versions

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(content)))
	L.Push(lua.LNil)
	return 2
}

// luaFsWrite writes content to a file.
// It takes two arguments: path (string), content (string).
// It returns (err string).
func luaFsWrite(L *lua.LState) int {
	path := L.CheckString(1)
	content := L.CheckString(2)
	err := os.WriteFile(path, []byte(content), 0644) // 0644 is common file permission

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaFsAppend appends content to a file.
// It takes two arguments: path (string), content (string).
// It returns (err string).
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

// luaFsExists checks if a file or directory exists.
// It takes one argument: path (string).
// It returns (exists bool).
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
	// Other error, treat as not existing for simplicity in Lua, or push error string
	L.Push(lua.LBool(false))
	return 1
}

// luaFsMkdir creates a directory.
// It takes one argument: path (string).
// It returns (err string).
func luaFsMkdir(L *lua.LState) int {
	path := L.CheckString(1)
	err := os.MkdirAll(path, 0755) // 0755 is common directory permission

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaFsRm removes a file or empty directory.
// It takes one argument: path (string).
// It returns (err string).
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

// luaFsRmR recursively removes a directory and its contents.
// It takes one argument: path (string).
// It returns (err string).
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

// luaFsLs lists files and directories in a path.
// It takes one argument: path (string).
// It returns (files table, err string).
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

// OpenFs opens the 'fs' library to the Lua state.
func OpenFs(L *lua.LState) {
	// Create a new table for the 'fs' module
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"read":   luaFsRead,
		"write":  luaFsWrite,
		"append": luaFsAppend,
		"exists": luaFsExists,
		"mkdir":  luaFsMkdir,
		"rm":     luaFsRm,
		"rm_r":   luaFsRmR,
		"ls":     luaFsLs,
	})
	// Set the 'fs' module in the global table
	L.SetGlobal("fs", mod)
}

// luaNetHttpGet performs an HTTP GET request.
// It takes one argument: url (string).
// It returns (body string, status_code number, headers table, err string).
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

// luaNetHttpPost performs an HTTP POST request.
// It takes three arguments: url (string), body (string), headers (table, optional).
// It returns (body string, status_code number, headers table, err string).
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

	// Set headers from Lua table
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

// luaNetDownload downloads a file from a URL to a destination path.
// It takes two arguments: url (string), destination_path (string).
// It returns (err string).
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

// OpenNet opens the 'net' library to the Lua state.
func OpenNet(L *lua.LState) {
	// Create a new table for the 'net' module
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"http_get":  luaNetHttpGet,
		"http_post": luaNetHttpPost,
		"download":  luaNetDownload,
	})
	// Set the 'net' module in the global table
	L.SetGlobal("net", mod)
}

// luaExecCommand executes a shell command.
// It takes one or more arguments: command (string), [args...] (string).
// It returns (stdout string, stderr string, err string).
func luaExecCommand(L *lua.LState) int {
	cmdName := L.CheckString(1)
	var args []string
	if L.GetTop() > 1 {
		for i := 2; i <= L.GetTop(); i++ {
			args = append(args, L.CheckString(i))
		}
	}

	cmd := exec.Command(cmdName, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		L.Push(lua.LString(stdout.String()))
		L.Push(lua.LString(stderr.String()))
		L.Push(lua.LString(err.Error()))
		return 3
	}

	L.Push(lua.LString(stdout.String()))
	L.Push(lua.LString(stderr.String()))
	L.Push(lua.LNil) // No error
	return 3
}

// OpenExec opens the 'exec' library to the Lua state.
func OpenExec(L *lua.LState) {
	// Create a new table for the 'exec' module
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"command": luaExecCommand,
	})
	// Set the 'exec' module in the global table
	L.SetGlobal("exec", mod)
}

// luaLogInfo logs an informational message.
func luaLogInfo(L *lua.LState) int {
	message := L.CheckString(1)
	log.Printf("[INFO] %s", message)
	return 0
}

// luaLogWarn logs a warning message.
func luaLogWarn(L *lua.LState) int {
	message := L.CheckString(1)
	log.Printf("[WARN] %s", message)
	return 0
}

// luaLogError logs an error message.
func luaLogError(L *lua.LState) int {
	message := L.CheckString(1)
	log.Printf("[ERROR] %s", message)
	return 0
}

// luaLogDebug logs a debug message.
func luaLogDebug(L *lua.LState) int {
	message := L.CheckString(1)
	log.Printf("[DEBUG] %s", message)
	return 0
}

// OpenLog opens the 'log' library to the Lua state.
func OpenLog(L *lua.LState) {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"info":  luaLogInfo,
		"warn":  luaLogWarn,
		"error": luaLogError,
		"debug": luaLogDebug,
	})
	L.SetGlobal("log", mod)
}

// luaSaltCmd executes a SaltStack command (salt or salt-call).
// It takes one or more arguments: command_type (string, "salt" or "salt-call"), [target (string)], [function (string)], [args...] (string).
// It returns (stdout string, stderr string, err string).
func luaSaltCmd(L *lua.LState) int {
	commandType := L.CheckString(1) // "salt" or "salt-call"
	var cmdArgs []string

	if commandType != "salt" && commandType != "salt-call" {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("invalid command type: %s. Expected 'salt' or 'salt-call'", commandType)))
		return 3
	}

	// Collect all arguments from Lua stack
	for i := 2; i <= L.GetTop(); i++ {
		cmdArgs = append(cmdArgs, L.CheckString(i))
	}

	cmd := exec.Command(commandType, cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		L.Push(lua.LString(stdout.String()))
		L.Push(lua.LString(stderr.String()))
		L.Push(lua.LString(err.Error()))
		return 3
	}

	L.Push(lua.LString(stdout.String()))
	L.Push(lua.LString(stderr.String()))
	L.Push(lua.LNil) // No error
	return 3
}

// OpenSalt opens the 'salt' library to the Lua state.
func OpenSalt(L *lua.LState) {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"cmd": luaSaltCmd,
	})
	L.SetGlobal("salt", mod)
}


// ExecuteLuaFunction calls a Lua function with given parameters and captures return values.
// It expects the Lua function to return (bool success, string message, [table output]).
// It returns (bool success, string message, *lua.LTable output, error goError).
func ExecuteLuaFunction(L *lua.LState, fn *lua.LFunction, params map[string]string, secondArg lua.LValue, nRet int) (bool, string, *lua.LTable, error) {
	L.Push(fn)

	// Push params as a Lua table
	luaParams := L.NewTable()
	for k, v := range params {
		luaParams.RawSetString(k, lua.LString(v))
	}
	L.Push(luaParams)

	// Push secondArg if provided
	numArgs := 1 // params is always the first arg
	if secondArg != nil {
		L.Push(secondArg)
		numArgs = 2
	}

	// Call the function with protected call (pcall)
	// We expect at least 2 return values (status, message) and optionally a third (output table)
	// nRet specifies the number of results to push onto the stack.
	// If nRet is 0, no results are pushed. If nRet is lua.MultRet, all results are pushed.
	if err := L.PCall(numArgs, nRet, nil); err != nil {
		return false, "", nil, fmt.Errorf("error executing Lua function: %w", err)
	}

	// Pop results from the stack.
	// The results are on the stack in the order they were returned by Lua.
	// So, if Lua returns (success, message, output), they are on stack as success, message, output.
	// We need to pop them in reverse order of how we want to process them.

	var success bool
	var message string
	var outputTable *lua.LTable

	// Get output table (if nRet allows for it and it's present)
	if nRet >= 3 && L.GetTop() >= 1 && L.Get(-1).Type() == lua.LTTable {
		outputTable = L.Get(-1).(*lua.LTable)
		L.Pop(1)
	} else if nRet >= 3 && L.GetTop() >= 1 && L.Get(-1).Type() != lua.LTNil {
		// If nRet expects an output table but it's not a table or nil, something is wrong
		return false, fmt.Sprintf("unexpected third return type from Lua: %s", L.Get(-1).String()), nil, nil
	} else if nRet >= 3 && L.GetTop() >= 1 && L.Get(-1).Type() == lua.LTNil {
		L.Pop(1) // Pop the nil if it was expected but not provided
	}


	// Get message (if nRet allows for it and it's present)
	if nRet >= 2 && L.GetTop() >= 1 && L.Get(-1).Type() == lua.LTString {
		message = lua.LVAsString(L.Get(-1))
		L.Pop(1)
	} else if nRet >= 2 && L.GetTop() >= 1 && L.Get(-1).Type() != lua.LTNil {
		// If nRet expects a message but it's not a string or nil, something is wrong
		return false, fmt.Sprintf("unexpected second return type from Lua: %s", L.Get(-1).String()), nil, nil
	} else if nRet >= 2 && L.GetTop() >= 1 && L.Get(-1).Type() == lua.LTNil {
		L.Pop(1) // Pop the nil if it was expected but not provided
	}


	// Get success (if nRet allows for it and it's present)
	if nRet >= 1 && L.GetTop() >= 1 && L.Get(-1).Type() == lua.LTBool {
		success = lua.LVAsBool(L.Get(-1))
		L.Pop(1)
	} else if nRet >= 1 && L.GetTop() >= 1 && L.Get(-1).Type() != lua.LTNil {
		// If nRet expects a success but it's not a bool or nil, something is wrong
		return false, fmt.Sprintf("unexpected first return type from Lua: %s", L.Get(-1).String()), nil, nil
	} else if nRet >= 1 && L.GetTop() >= 1 && L.Get(-1).Type() == lua.LTNil {
		L.Pop(1) // Pop the nil if it was expected but not provided
	}


	return success, message, outputTable, nil
}

// LoadTaskDefinitions loads Lua script content and parses the task definitions.
func LoadTaskDefinitions(L *lua.LState, luaScriptContent string) (map[string]types.TaskGroup, error) {
	// Load the Lua script content. This will populate the global TaskDefinitions table in Lua.
	if err := L.DoString(luaScriptContent); err != nil {
		return nil, fmt.Errorf("error loading Lua script content: %w", err)
	}

	// Get the 'TaskDefinitions' global table from Lua
	globalTaskDefs := L.GetGlobal("TaskDefinitions")
	if globalTaskDefs.Type() != lua.LTTable {
		return nil, fmt.Errorf("expected 'TaskDefinitions' to be a table, got %s", globalTaskDefs.Type().String())
	}

	loadedTaskGroups := make(map[string]types.TaskGroup)

	// Iterate over the top-level TaskDefinitions table (groups)
	globalTaskDefs.(*lua.LTable).ForEach(func(groupKey, groupValue lua.LValue) {
		groupName := groupKey.String()
		if groupValue.Type() != lua.LTTable {
			log.Printf("Warning: Expected group '%s' to be a table, skipping.", groupName)
			return
		}

		groupTable := groupValue.(*lua.LTable)
		description := groupTable.RawGetString("description").String()
		var tasks []types.Task

		// Get the 'tasks' table within the group
		luaTasks := groupTable.RawGetString("tasks")
		if luaTasks.Type() == lua.LTTable {
			luaTasks.(*lua.LTable).ForEach(func(taskKey, taskValue lua.LValue) {
				if taskValue.Type() != lua.LTTable {
					log.Printf("Warning: Expected task entry in group '%s' to be a table, skipping.", groupName)
					return
				}
				taskTable := taskValue.(*lua.LTable)

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

				params := make(map[string]string)
				luaParams := taskTable.RawGetString("params")
				if luaParams.Type() == lua.LTTable {
					luaParams.(*lua.LTable).ForEach(func(paramKey, paramValue lua.LValue) {
						params[paramKey.String()] = paramValue.String()
					})
				}

				var preExecFunc *lua.LFunction
				luaPreExec := taskTable.RawGetString("pre_exec")
				if luaPreExec.Type() == lua.LTFunction {
					preExecFunc = luaPreExec.(*lua.LFunction)
				}

				var postExecFunc *lua.LFunction
				luaPostExec := taskTable.RawGetString("post_exec")
				if luaPostExec.Type() == lua.LTFunction {
					postExecFunc = luaPostExec.(*lua.LFunction)
				}

				async := false
				luaAsync := taskTable.RawGetString("async")
				if luaAsync.Type() == lua.LTBool {
					async = luaAsync.(lua.LBool).String() == "true"
				}

				var dependsOn []string
				luaDependsOn := taskTable.RawGetString("depends_on")
				if luaDependsOn.Type() == lua.LTString {
					dependsOn = []string{luaDependsOn.String()}
				} else if luaDependsOn.Type() == lua.LTTable {
					luaDependsOn.(*lua.LTable).ForEach(func(depKey, depValue lua.LValue) {
						if depValue.Type() == lua.LTString {
							dependsOn = append(dependsOn, depValue.String())
						} else {
							log.Printf("Warning: Non-string dependency found in depends_on table for task '%s'. Skipping.", name)
						}
					})
				} else if luaDependsOn.Type() == lua.LTNil {
					dependsOn = []string{} // No dependencies
				} else {
					log.Printf("Warning: Unexpected type for depends_on field for task '%s'. Expected string or table, got %s. Treating as no dependencies.", name, luaDependsOn.Type().String())
					dependsOn = []string{}
				}

				tasks = append(tasks, types.Task{
					Name:        name,
					Description: desc,
					CommandFunc: cmdFunc,
					CommandStr:  cmdStr,
					Params:      params,
					PreExec:     preExecFunc,
					PostExec:    postExecFunc,
					Async:       async,
					DependsOn:   dependsOn,
				})
			})
		}

		loadedTaskGroups[groupName] = types.TaskGroup{
			Description: description,
			Tasks:       tasks,
		}
	})
	return loadedTaskGroups, nil
}

// LuaTableToGoMap converts a lua.LTable to a Go map[string]interface{}.
// It handles nested tables and basic Lua types.
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
			result[k] = value.String() // Fallback for other types (e.g., functions, userdata)
		}
	})
	return result
}
