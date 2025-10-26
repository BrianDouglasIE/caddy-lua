package lute

import (
	"net/http"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func luaServeHTTP(w http.ResponseWriter, r *http.Request, luaArg string, isFilePath bool) error {
	L := lua.NewState()
	defer L.Close()

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

	respTable := L.NewTable()
	L.SetField(respTable, "Status", lua.LNumber(http.StatusOK))
	L.SetField(respTable, "Header", L.NewTable())
	L.SetField(respTable, "Body", lua.LString(""))

	L.SetGlobal("__CADDY_REQUEST", reqTable)
	L.SetGlobal("__CADDY_RESPONSE", respTable)

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
