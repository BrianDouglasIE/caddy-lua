package LOOTBOX

import (
	"encoding/json"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	lua "github.com/yuin/gopher-lua"
)

func luaServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler, luaBlock string) error {
	L := lua.NewState()
	defer L.Close()

	respTable := createResponseTable(L) // __LOOTBOX_RES
	createRequestTable(L, r)            // __LOOTBOX_REQ
	createServerInfoTable(L, r)         // __LOOTBOX_SRV
	createEnvTable(L)                   // __LOOTBOX_ENV
	createUtilTable(L)                  // __LOOTBOX_UTL
	setCaddyNextMethod(L, w, r, next)   // __LOOTBOX_NXT

	if err := L.DoString(luaBlock); err != nil {
		http.Error(w, "Lua script error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	writeLuaResponse(L, w, respTable)
	return nil
}

func luaServeHTTPFile(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler, filepath string) error {
	L := lua.NewState()
	defer L.Close()

	respTable := createResponseTable(L) // __LOOTBOX_RES
	createRequestTable(L, r)            // __LOOTBOX_REQ
	createServerInfoTable(L, r)         // __LOOTBOX_SRV
	createEnvTable(L)                   // __LOOTBOX_ENV
	createUtilTable(L)                  // __LOOTBOX_UTL
	setCaddyNextMethod(L, w, r, next)   // __LOOTBOX_NXT

	if err := L.DoFile(filepath); err != nil {
		http.Error(w, "Lua runtime error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	writeLuaResponse(L, w, respTable)
	return nil
}

func writeLuaResponse(L *lua.LState, w http.ResponseWriter, respTable *lua.LTable) {
	status := getResponseStatus(L.GetField(respTable, "Status"))

	respHeaders := http.Header{}
	luaHeaders := L.GetField(respTable, "Header")
	if luaHeaders != lua.LNil {
		tbl, ok := luaHeaders.(*lua.LTable)
		if !ok {
			L.RaiseError("response Header must be a table, got %s", luaHeaders.Type().String())
			return
		}

		tbl.ForEach(func(k, v lua.LValue) {
			key := textproto.CanonicalMIMEHeaderKey(k.String())
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

	var bodyBytes []byte
	body := L.GetField(respTable, "Body")
	if body == lua.LNil {
		bodyBytes = nil
	} else if bodyStr, ok := body.(lua.LString); ok {
		bodyBytes = []byte(string(bodyStr))
	} else {
		L.RaiseError("response Body must be a string, got %s", body.Type().String())
		return
	}

	for k, vv := range respHeaders {
		for _, v := range vv {
			w.Header()[k] = append(w.Header()[k], v)
		}
	}

	w.WriteHeader(status)

	if len(bodyBytes) > 0 && status >= 200 && status != http.StatusNoContent && status != http.StatusNotModified {
		if _, err := w.Write(bodyBytes); err != nil {
			log.Printf("error writing response body: %v", err)
		}
	}
}

func getResponseStatus(statusField lua.LValue) int {
	if n, ok := statusField.(lua.LNumber); ok {
		code := int(n)
		if code >= 100 && code <= 599 {
			return code
		}
		return http.StatusOK
	}

	if s, ok := statusField.(lua.LString); ok {
		if parsed, err := strconv.Atoi(string(s)); err == nil && parsed >= 100 && parsed <= 599 {
			return parsed
		}
	}

	return http.StatusOK
}

func createRequestTable(L *lua.LState, r *http.Request) {
	reqTable := L.NewTable()
	L.SetGlobal("__LOOTBOX_REQ", reqTable)
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
}

func createResponseTable(L *lua.LState) *lua.LTable {
	respTable := L.NewTable()
	L.SetGlobal("__LOOTBOX_RES", respTable)
	L.SetField(respTable, "Status", lua.LNumber(http.StatusOK))
	L.SetField(respTable, "Header", L.NewTable())
	L.SetField(respTable, "Body", lua.LString(""))
	return respTable
}

func createServerInfoTable(L *lua.LState, r *http.Request) {
	serverInfoTable := L.NewTable()
	L.SetGlobal("__LOOTBOX_SRV", serverInfoTable)
	L.SetField(serverInfoTable, "Version", lua.LString(caddy.AppVersion))
	L.SetField(serverInfoTable, "Module", lua.LString("http.handlers.lua"))
	L.SetField(serverInfoTable, "Hostname", lua.LString(r.Host))
	L.SetField(serverInfoTable, "TLS", lua.LBool(r.TLS != nil))
}

func createEnvTable(L *lua.LState) {
	env := L.NewTable()
	L.SetGlobal("__LOOTBOX_ENV", env)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		L.SetField(env, parts[0], lua.LString(parts[1]))
	}
}

func createUtilTable(L *lua.LState) {
	utilTable := L.NewTable()
	L.SetGlobal("__LOOTBOX_UTL", utilTable)

	L.SetField(utilTable, "json_encode", L.NewFunction(func(L *lua.LState) int {
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

	L.SetField(utilTable, "json_decode", L.NewFunction(func(L *lua.LState) int {
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
}

func setCaddyNextMethod(L *lua.LState, w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) {
	L.SetGlobal("__LOOTBOX_NXT", L.NewFunction(func(L *lua.LState) int {
		err := next.ServeHTTP(w, r)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		return 0
	}))
}
