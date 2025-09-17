package luainterface

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

func TestGoValueToLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("bool", func(t *testing.T) {
		val := GoValueToLua(L, true)
		assert.Equal(t, lua.LBool(true), val)
	})

	t.Run("float64", func(t *testing.T) {
		val := GoValueToLua(L, 123.45)
		assert.Equal(t, lua.LNumber(123.45), val)
	})

	t.Run("int", func(t *testing.T) {
		val := GoValueToLua(L, 123)
		assert.Equal(t, lua.LNumber(123), val)
	})

	t.Run("string", func(t *testing.T) {
		val := GoValueToLua(L, "hello")
		assert.Equal(t, lua.LString("hello"), val)
	})

	t.Run("nil", func(t *testing.T) {
		val := GoValueToLua(L, nil)
		assert.Equal(t, lua.LNil, val)
	})

	t.Run("slice", func(t *testing.T) {
		slice := []interface{}{"a", 1}
		val := GoValueToLua(L, slice)
		tbl := val.(*lua.LTable)
		assert.Equal(t, lua.LString("a"), tbl.RawGetInt(1))
		assert.Equal(t, lua.LNumber(1), tbl.RawGetInt(2))
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]interface{}{"a": "b", "c": 1}
		val := GoValueToLua(L, m)
		tbl := val.(*lua.LTable)
		assert.Equal(t, lua.LString("b"), tbl.RawGetString("a"))
		assert.Equal(t, lua.LNumber(1), tbl.RawGetString("c"))
	})
}

func TestLuaToGoValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("bool", func(t *testing.T) {
		val := LuaToGoValue(L, lua.LBool(true))
		assert.Equal(t, true, val)
	})

	t.Run("number", func(t *testing.T) {
		val := LuaToGoValue(L, lua.LNumber(123.45))
		assert.Equal(t, 123.45, val)
	})

	t.Run("string", func(t *testing.T) {
		val := LuaToGoValue(L, lua.LString("hello"))
		assert.Equal(t, "hello", val)
	})

	t.Run("nil", func(t *testing.T) {
		val := LuaToGoValue(L, lua.LNil)
		assert.Nil(t, val)
	})

	t.Run("table (array)", func(t *testing.T) {
		tbl := L.NewTable()
		tbl.RawSetInt(1, lua.LString("a"))
		tbl.RawSetInt(2, lua.LNumber(1))
		val := LuaToGoValue(L, tbl)
		slice := val.([]interface{})
		assert.Equal(t, "a", slice[0])
		assert.Equal(t, float64(1), slice[1])
	})

	t.Run("table (map)", func(t *testing.T) {
		tbl := L.NewTable()
		tbl.RawSetString("a", lua.LString("b"))
		tbl.RawSetString("c", lua.LNumber(1))
		val := LuaToGoValue(L, tbl)
		m := val.(map[string]interface{})
		assert.Equal(t, "b", m["a"])
		assert.Equal(t, float64(1), m["c"])
	})
}

func TestJsonConversion(t *testing.T) {
	t.Run("parse_json", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		jsonStr := `{"a":"b","c":1}`
		L.Push(lua.LString(jsonStr))
		luaDataParseJson(L)
		tbl := L.Get(2).(*lua.LTable)
		assert.Equal(t, lua.LString("b"), tbl.RawGetString("a"))
		assert.Equal(t, lua.LNumber(1), tbl.RawGetString("c"))
	})

	t.Run("to_json", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		jsonStr := `{"a":"b","c":1}`
		tbl := L.NewTable()
		tbl.RawSetString("a", lua.LString("b"))
		tbl.RawSetString("c", lua.LNumber(1))
		L.Push(tbl)
		luaDataToJson(L)
		resJsonStr := L.Get(2).String()
		assert.JSONEq(t, jsonStr, resJsonStr)
	})
}

func TestYamlConversion(t *testing.T) {
	t.Run("parse_yaml", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		yamlStr := `
a: b
c: 1
`
		L.Push(lua.LString(yamlStr))
		luaDataParseYaml(L)
		tbl := L.Get(2).(*lua.LTable)
		assert.Equal(t, lua.LString("b"), tbl.RawGetString("a"))
		assert.Equal(t, lua.LNumber(1), tbl.RawGetString("c"))
	})

	t.Run("to_yaml", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		tbl := L.NewTable()
		tbl.RawSetString("a", lua.LString("b"))
		tbl.RawSetString("c", lua.LNumber(1))
		L.Push(tbl)
		luaDataToYaml(L)
		resYamlStr := L.Get(2).String()
		assert.Contains(t, resYamlStr, "a: b")
		assert.Contains(t, resYamlStr, "c: 1")
	})
}

func TestFsLibrary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-fs")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := tmpDir + "/test.txt"
	content := "hello world"

	t.Run("write", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(filePath))
		L.Push(lua.LString(content))
		luaFsWrite(L)
		assert.Equal(t, lua.LNil, L.Get(-1))
	})

	t.Run("read", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(filePath))
		luaFsRead(L)
		readContent := L.Get(2).String()
		assert.Equal(t, content, readContent)
	})

	t.Run("exists", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(filePath))
		luaFsExists(L)
		exists := L.Get(2).(lua.LBool)
		assert.True(t, bool(exists))
	})

	t.Run("append", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		appendedContent := " more content"
		L.Push(lua.LString(filePath))
		L.Push(lua.LString(appendedContent))
		luaFsAppend(L)
		assert.Equal(t, lua.LNil, L.Get(-1))
	})

	t.Run("read again", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(filePath))
		luaFsRead(L)
		readContent := L.Get(2).String()
		assert.Equal(t, content+" more content", readContent)
	})

	t.Run("ls", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(tmpDir))
		luaFsLs(L)
		filesTable := L.Get(2).(*lua.LTable)
		assert.Equal(t, 1, filesTable.Len())
		assert.Equal(t, "test.txt", filesTable.RawGetInt(1).String())
	})

	t.Run("rm", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(filePath))
		luaFsRm(L)
		assert.Equal(t, lua.LNil, L.Get(-1))
	})

	t.Run("exists again", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		L.Push(lua.LString(filePath))
		luaFsExists(L)
		exists := L.Get(2).(lua.LBool)
		assert.False(t, bool(exists))
	})

	t.Run("mkdir", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		newDirPath := tmpDir + "/newdir"
		L.Push(lua.LString(newDirPath))
		luaFsMkdir(L)
		assert.Equal(t, lua.LNil, L.Get(-1))
	})

	t.Run("rm_r", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenFs(L)
		newDirPath := tmpDir + "/newdir"
		L.Push(lua.LString(newDirPath))
		luaFsRmR(L)
		assert.Equal(t, lua.LNil, L.Get(-1))
	})
}

func TestNetLibrary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("get response"))
		case "/post":
			body, _ := io.ReadAll(r.Body)
			w.Write([]byte("posted: " + string(body)))
		case "/download":
			w.Header().Set("Content-Disposition", "attachment; filename=test.txt")
			w.Write([]byte("file content"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Run("http_get", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenNet(L)
		L.Push(lua.LString(server.URL + "/get"))
		luaNetHttpGet(L)
		body := L.Get(-4).String()
		status := int(L.Get(-3).(lua.LNumber))
		assert.Equal(t, "get response", body)
		assert.Equal(t, http.StatusOK, status)
	})

	t.Run("http_post", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenNet(L)
		L.Push(lua.LString(server.URL + "/post"))
		L.Push(lua.LString("test body"))
		L.Push(L.NewTable()) // Empty headers
		luaNetHttpPost(L)
		body := L.Get(-4).String()
		status := int(L.Get(-3).(lua.LNumber))
		assert.Equal(t, "posted: test body", body)
		assert.Equal(t, http.StatusOK, status)
	})

	t.Run("download", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenNet(L)
		tmpDir, err := os.MkdirTemp("", "test-download")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)
		filePath := tmpDir + "/test.txt"

		L.Push(lua.LString(server.URL + "/download"))
		L.Push(lua.LString(filePath))
		luaNetDownload(L)
		assert.Equal(t, lua.LNil, L.Get(-1))

		content, err := os.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "file content", string(content))
	})
}

func TestExecLibrary(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	OpenExec(L)

	// Test command
	L.Push(lua.LString("echo"))
	L.Push(lua.LString("hello"))
	L.Push(lua.LString("world"))
	luaExecCommand(L)

	stdout := L.Get(-3).String()
	stderr := L.Get(-2).String()
	err := L.Get(-1)

	assert.Equal(t, "hello world\n", stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, lua.LNil, err)
}

func TestLogLibrary(t *testing.T) {
	t.Run("info", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenLog(L)
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()
		L.Push(lua.LString("info message"))
		luaLogInfo(L)
		assert.Contains(t, buf.String(), "[INFO] info message")
	})

	t.Run("warn", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenLog(L)
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()
		L.Push(lua.LString("warn message"))
		luaLogWarn(L)
		assert.Contains(t, buf.String(), "[WARN] warn message")
	})

	t.Run("error", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenLog(L)
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()
		L.Push(lua.LString("error message"))
		luaLogError(L)
		assert.Contains(t, buf.String(), "[ERROR] error message")
	})

	t.Run("debug", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()
		OpenLog(L)
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()
		L.Push(lua.LString("debug message"))
		luaLogDebug(L)
		assert.Contains(t, buf.String(), "[DEBUG] debug message")
	})
}

func TestExecuteLuaFunction(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("3 return values", func(t *testing.T) {
		err := L.DoString(`
			function test_func(params, inputs)
				return true, "success", {result = "output"}
			end
		`)
		assert.NoError(t, err)

		fn := L.GetGlobal("test_func").(*lua.LFunction)
		params := map[string]string{"p1": "v1"}
		inputs := L.NewTable()

		success, msg, output, err := ExecuteLuaFunction(L, fn, params, inputs, 3)
		assert.NoError(t, err)
		assert.True(t, success)
		assert.Equal(t, "success", msg)
		assert.Equal(t, "output", output.RawGetString("result").String())
	})

	t.Run("2 return values", func(t *testing.T) {
		err := L.DoString(`
			function test_func_2(params, inputs)
				return false, "failure"
			end
		`)
		assert.NoError(t, err)

		fn := L.GetGlobal("test_func_2").(*lua.LFunction)
		params := map[string]string{}

		success, msg, output, err := ExecuteLuaFunction(L, fn, params, nil, 2)
		assert.NoError(t, err)
		assert.False(t, success)
		assert.Equal(t, "failure", msg)
		assert.Nil(t, output)
	})

	t.Run("error in function", func(t *testing.T) {
		err := L.DoString(`
			function test_func_3(params, inputs)
				error("test error")
			end
		`)
		assert.NoError(t, err)

		fn := L.GetGlobal("test_func_3").(*lua.LFunction)
		params := map[string]string{}

		_, _, _, err = ExecuteLuaFunction(L, fn, params, nil, 0)
		assert.Error(t, err)
	})
}

func TestLoadTaskDefinitions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	luaScript := `
	TaskDefinitions = {
		group1 = {
			description = "Group 1",
			tasks = {
				{
					name = "task1",
					description = "Task 1",
					command = "echo 'hello'",
					params = {p1 = "v1"},
					depends_on = "task2",
				},
				{
					name = "task2",
					description = "Task 2",
					command = function(params, inputs)
						return true, "command executed", {output = "data"}
					end,
					pre_exec = function(params, inputs)
						return true, "pre-exec ok"
					end,
					post_exec = function(params, output)
						return true, "post-exec ok"
					end,
					async = true,
				},
			},
		},
	}
	`

	taskGroups, err := LoadTaskDefinitions(L, luaScript)
	assert.NoError(t, err)
	assert.Len(t, taskGroups, 1)

	group1 := taskGroups["group1"]
	assert.Equal(t, "Group 1", group1.Description)
	assert.Len(t, group1.Tasks, 2)

	task1 := group1.Tasks[0]
	assert.Equal(t, "task1", task1.Name)
	assert.Equal(t, "echo 'hello'", task1.CommandStr)
	assert.Equal(t, "v1", task1.Params["p1"])
	assert.Equal(t, []string{"task2"}, task1.DependsOn)

	task2 := group1.Tasks[1]
	assert.Equal(t, "task2", task2.Name)
	assert.NotNil(t, task2.CommandFunc)
	assert.NotNil(t, task2.PreExec)
	assert.NotNil(t, task2.PostExec)
	assert.True(t, task2.Async)
}
