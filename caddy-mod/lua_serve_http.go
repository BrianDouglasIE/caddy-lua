package LOOTBOX

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
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
	createUrlTable(L, r.URL)            // __LOOTBOX_URL
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
	createUrlTable(L, r.URL)            // __LOOTBOX_URL
	setCaddyNextMethod(L, w, r, next)   // __LOOTBOX_NXT

	if err := L.DoFile(filepath); err != nil {
		http.Error(w, "Lua runtime error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	writeLuaResponse(L, w, respTable)
	return nil
}

func writeLuaResponse(L *lua.LState, w http.ResponseWriter, respTable *lua.LTable) {
	status := getResponseStatus(L.GetField(respTable, "status"))

	respHeaders := http.Header{}
	luaHeaders := L.GetField(respTable, "header")
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
	L.SetField(reqTable, "method", lua.LString(r.Method))
	L.SetField(reqTable, "url", createUrlTable(L, r.URL))
	L.SetField(reqTable, "proto", lua.LString(r.Proto))
	L.SetField(reqTable, "host", lua.LString(r.Host))
	L.SetField(reqTable, "remote_addr", lua.LString(r.RemoteAddr))

	headers := L.NewTable()
	for k, vv := range r.Header {
		arr := L.NewTable()
		for i, v := range vv {
			arr.RawSetInt(i+1, lua.LString(v))
		}
		headers.RawSetString(k, arr)
	}
	L.SetField(reqTable, "header", headers)
}

func createUrlTable(L *lua.LState, URL *url.URL) *lua.LTable {
	urlTable := L.NewTable()

	L.SetField(urlTable, "protocol", lua.LString(URL.Scheme))
	L.SetField(urlTable, "username", lua.LString(URL.User.Username()))
	if pass, ok := URL.User.Password(); ok {
		L.SetField(urlTable, "password", lua.LString(pass))
	} else {
		L.SetField(urlTable, "password", lua.LString(""))
	}
	L.SetField(urlTable, "hostname", lua.LString(URL.Hostname()))
	L.SetField(urlTable, "port", lua.LString(URL.Port()))
	L.SetField(urlTable, "pathname", lua.LString(URL.Path))
	L.SetField(urlTable, "search", lua.LString(URL.RawQuery))
	L.SetField(urlTable, "hash", lua.LString(URL.Fragment))

	L.SetField(urlTable, "href", lua.LString(URL.String()))

	return urlTable
}

func createResponseTable(L *lua.LState) *lua.LTable {
	respTable := L.NewTable()
	L.SetGlobal("__LOOTBOX_RES", respTable)
	L.SetField(respTable, "status", lua.LNumber(http.StatusOK))
	L.SetField(respTable, "header", L.NewTable())
	L.SetField(respTable, "body", lua.LString(""))
	return respTable
}

func createServerInfoTable(L *lua.LState, r *http.Request) {
	serverInfoTable := L.NewTable()
	L.SetGlobal("__LOOTBOX_SRV", serverInfoTable)
	L.SetField(serverInfoTable, "version", lua.LString(caddy.AppVersion))
	L.SetField(serverInfoTable, "module", lua.LString("http.handlers.lua"))
	L.SetField(serverInfoTable, "hostname", lua.LString(r.Host))
	L.SetField(serverInfoTable, "tls", lua.LBool(r.TLS != nil))
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
		var decoded any
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
