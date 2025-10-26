package lute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	lua "github.com/yuin/gopher-lua"
)

func luaServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler, luaArg string, isFilePath bool) error {
	L := lua.NewState()
	defer L.Close()

	// __CADDY_REQUEST
	reqTable := L.NewTable()
	L.SetField(reqTable, "Method", lua.LString(r.Method))
	L.SetField(reqTable, "URL", lua.LString(r.URL.String()))
	L.SetField(reqTable, "Proto", lua.LString(r.Proto))
	L.SetField(reqTable, "Host", lua.LString(r.Host))
	L.SetField(reqTable, "RemoteAddr", lua.LString(r.RemoteAddr))

	headers := L.NewTable()
	for k, vv := range r.Header {
		arr := L.NewTable()
		for i, v := range vv {
			arr.RawSetInt(i+1, lua.LString(v))
		}
		headers.RawSetString(k, arr)
	}
	L.SetField(reqTable, "Header", headers)
	L.SetGlobal("__CADDY_REQUEST", reqTable)

	// __CADDY_RESPONSE
	respTable := L.NewTable()
	L.SetField(respTable, "Status", lua.LNumber(http.StatusOK))
	L.SetField(respTable, "Header", L.NewTable())
	L.SetField(respTable, "Body", lua.LString(""))
	L.SetGlobal("__CADDY_RESPONSE", respTable)

	// __CADDY_SERVER_INFO
	L.SetGlobal("__CADDY_SERVER_INFO", L.NewTable())
	server := L.GetGlobal("__CADDY_SERVER_INFO").(*lua.LTable)
	L.SetField(server, "Version", lua.LString(caddy.AppVersion))
	L.SetField(server, "Module", lua.LString("http.handlers.lua"))
	L.SetField(server, "Hostname", lua.LString(r.Host))
	L.SetField(server, "TLS", lua.LBool(r.TLS != nil))

	// __CADDY_UTIL
	util := L.NewTable()
	L.SetGlobal("__CADDY_UTIL", util)

	L.SetField(util, "json_encode", L.NewFunction(func(L *lua.LState) int {
		val := L.CheckAny(1)
		goVal := luaToGo(val)
		jsonBytes, err := json.Marshal(goVal)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(jsonBytes))
		return 1
	}))

	L.SetField(util, "json_decode", L.NewFunction(func(L *lua.LState) int {
		str := L.CheckString(1)
		var decoded interface{}
		if err := json.Unmarshal([]byte(str), &decoded); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(goToLua(L, decoded))
		return 1
	}))

	// __CADDY_NEXT
	L.SetGlobal("__CADDY_NEXT", L.NewFunction(func(L *lua.LState) int {
		err := next.ServeHTTP(w, r)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		return 0
	}))

	// __CADDY_ENV
	env := L.NewTable()
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		L.SetField(env, parts[0], lua.LString(parts[1]))
	}
	L.SetGlobal("__CADDY_ENV", env)

	if isFilePath {
		if err := L.DoFile(luaArg); err != nil {
			http.Error(w, "Lua script error: "+err.Error(), http.StatusInternalServerError)
			return nil
		}
	} else {
		if err := L.DoString(luaArg); err != nil {
			http.Error(w, "Lua script error: "+err.Error(), http.StatusInternalServerError)
			return nil
		}
	}

	status := int(lua.LVAsNumber(L.GetField(respTable, "Status")))
	if status == 0 {
		status = http.StatusOK
	}

	respHeaders := http.Header{}
	luaHeaders := L.GetField(respTable, "Header")
	if luaHeaders != lua.LNil {
		if tbl, ok := luaHeaders.(*lua.LTable); ok {
			tbl.ForEach(func(k, v lua.LValue) {
				key := k.String()
				switch vv := v.(type) {
				case *lua.LTable:
					vv.ForEach(func(_, val lua.LValue) {
						respHeaders.Add(key, val.String())
					})
				default:
					respHeaders.Set(key, vv.String())
				}
			})
		}
	}

	body := L.GetField(respTable, "Body")
	var bodyBytes []byte
	switch b := body.(type) {
	case lua.LString:
		bodyBytes = []byte(b)
	case *lua.LTable:
		var parts []string
		b.ForEach(func(_, v lua.LValue) {
			parts = append(parts, v.String())
		})
		bodyBytes = []byte(strings.Join(parts, ""))
	}

	for k, vv := range respHeaders {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(status)
	if bodyBytes != nil {
		_, _ = w.Write(bodyBytes)
	}
	return nil
}

func luaToGo(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(v)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(v))
	case lua.LTString:
		return lua.LVAsString(v)
	case lua.LTTable:
		tbl := v.(*lua.LTable)
		isArray := true
		var maxKey int
		tbl.ForEach(func(k, _ lua.LValue) {
			if key, ok := k.(lua.LNumber); ok {
				if int(key) > maxKey {
					maxKey = int(key)
				}
			} else {
				isArray = false
			}
		})
		if isArray {
			arr := make([]interface{}, maxKey)
			tbl.ForEach(func(k, v lua.LValue) {
				idx := int(k.(lua.LNumber))
				if idx > 0 && idx <= len(arr) {
					arr[idx-1] = luaToGo(v)
				}
			})
			return arr
		}
		obj := map[string]interface{}{}
		tbl.ForEach(func(k, v lua.LValue) {
			obj[k.String()] = luaToGo(v)
		})
		return obj
	default:
		return v.String()
	}
}

func goToLua(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case float64:
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case []interface{}:
		tbl := L.NewTable()
		for i, elem := range val {
			tbl.RawSetInt(i+1, goToLua(L, elem))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, elem := range val {
			tbl.RawSetString(k, goToLua(L, elem))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}
