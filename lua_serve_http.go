package lute

import (
	"encoding/json"
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
	L.SetGlobal("__CADDY_REQUEST", reqTable)
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

	// __CADDY_RESPONSE
	respTable := L.NewTable()
	L.SetGlobal("__CADDY_RESPONSE", respTable)
	L.SetField(respTable, "Status", lua.LNumber(http.StatusOK))
	L.SetField(respTable, "Header", L.NewTable())
	L.SetField(respTable, "Body", lua.LString(""))

	// __CADDY_SERVER_INFO
	serverInfoTable := L.NewTable()
	L.SetGlobal("__CADDY_SERVER_INFO", serverInfoTable)
	L.SetField(serverInfoTable, "Version", lua.LString(caddy.AppVersion))
	L.SetField(serverInfoTable, "Module", lua.LString("http.handlers.lua"))
	L.SetField(serverInfoTable, "Hostname", lua.LString(r.Host))
	L.SetField(serverInfoTable, "TLS", lua.LBool(r.TLS != nil))

	// __CADDY_UTIL
	utilTable := L.NewTable()
	L.SetGlobal("__CADDY_UTIL", utilTable)

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
	L.SetGlobal("__CADDY_ENV", env)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		L.SetField(env, parts[0], lua.LString(parts[1]))
	}

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
