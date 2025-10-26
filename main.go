package lute

import (
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	caddy.RegisterModule(LuaHandler{})
	caddy.RegisterModule(LuaFileHandler{})
	httpcaddyfile.RegisterHandlerDirective("lua", parseCaddyfileForLuaBlock)
	httpcaddyfile.RegisterDirectiveOrder("lua", httpcaddyfile.After, "header")
	httpcaddyfile.RegisterHandlerDirective("lua_file", parseCaddyfileForLuaFileHandler)
	httpcaddyfile.RegisterDirectiveOrder("lua_file", httpcaddyfile.After, "header")
}

func parseCaddyfileForLuaBlock(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler LuaHandler
	for h.Next() {
		if !h.NextArg() {
			return nil, h.ArgErr()
		}
		handler.LuaBlock = h.Val()
		if h.NextArg() {
			return nil, h.ArgErr()
		}
	}
	return &handler, nil
}

func parseCaddyfileForLuaFileHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler LuaFileHandler
	for h.Next() {
		if !h.NextArg() {
			return nil, h.ArgErr()
		}
		handler.ScriptPath = h.Val()
		if h.NextArg() {
			return nil, h.ArgErr()
		}
	}
	return &handler, nil
}

// ---------------------------------------------------------------------
//
//	Caddyfile unmarshalling
//
// ---------------------------------------------------------------------
func (h *LuaHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.NextArg() {
			return d.ArgErr()
		}
		h.LuaBlock = d.Val()

		if d.NextArg() {
			return d.ArgErr()
		}
	}
	return nil
}

func (h *LuaFileHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.NextArg() {
			return d.ArgErr()
		}
		h.ScriptPath = d.Val()

		if d.NextArg() {
			return d.ArgErr()
		}
	}
	return nil
}

var (
	_ caddy.Provisioner           = (*LuaHandler)(nil)
	_ caddyhttp.MiddlewareHandler = (*LuaHandler)(nil)
	_ caddyfile.Unmarshaler       = (*LuaHandler)(nil)

	_ caddy.Provisioner           = (*LuaFileHandler)(nil)
	_ caddyhttp.MiddlewareHandler = (*LuaFileHandler)(nil)
	_ caddyfile.Unmarshaler       = (*LuaFileHandler)(nil)
)

func luaServeHTTPScriptPath(w http.ResponseWriter, r *http.Request, scriptPath string) error {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(scriptPath); err != nil {
		http.Error(w, "Lua script error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	fn := L.GetGlobal("handle")
	if fn == lua.LNil {
		http.Error(w, "Lua script must define a global function `handle(req, resp)`", http.StatusInternalServerError)
		return nil
	}
	luaFn, ok := fn.(*lua.LFunction)
	if !ok {
		http.Error(w, "`handle` is not a function", http.StatusInternalServerError)
		return nil
	}

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

	if err := L.CallByParam(lua.P{
		Fn:      luaFn,
		NRet:    0,
		Protect: true,
	}, reqTable, respTable); err != nil {
		http.Error(w, "Lua runtime error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	status := int(L.GetField(respTable, "Status").(lua.LNumber))
	if status == 0 {
		status = http.StatusOK
	}

	respHeaders := http.Header{}
	luaHeaders := L.GetField(respTable, "Header")
	if luaHeaders != lua.LNil {
		L.ForEach(luaHeaders.(*lua.LTable), func(k, v lua.LValue) {
			key := string(k.(lua.LString))
			if arr, ok := v.(*lua.LTable); ok {
				arr.ForEach(func(_, val lua.LValue) {
					respHeaders.Add(key, string(val.(lua.LString)))
				})
			} else {
				respHeaders.Set(key, string(v.(lua.LString)))
			}
		})
	}

	body := L.GetField(respTable, "Body")
	var bodyBytes []byte
	switch b := body.(type) {
	case lua.LString:
		bodyBytes = []byte(b)
	case *lua.LTable:
		var parts []string
		b.ForEach(func(_, v lua.LValue) {
			parts = append(parts, string(v.(lua.LString)))
		})
		bodyBytes = []byte(strings.Join(parts, ""))
	default:
		bodyBytes = nil
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

func luaServeHTTPBlock(w http.ResponseWriter, r *http.Request, luaBlock string) error {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoString(luaBlock); err != nil {
		http.Error(w, "Lua script error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	fn := L.GetGlobal("handle")
	if fn == lua.LNil {
		http.Error(w, "Lua script must define a global function `handle(req, resp)`", http.StatusInternalServerError)
		return nil
	}
	luaFn, ok := fn.(*lua.LFunction)
	if !ok {
		http.Error(w, "`handle` is not a function", http.StatusInternalServerError)
		return nil
	}

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

	if err := L.CallByParam(lua.P{
		Fn:      luaFn,
		NRet:    0,
		Protect: true,
	}, reqTable, respTable); err != nil {
		http.Error(w, "Lua runtime error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	status := int(L.GetField(respTable, "Status").(lua.LNumber))
	if status == 0 {
		status = http.StatusOK
	}

	respHeaders := http.Header{}
	luaHeaders := L.GetField(respTable, "Header")
	if luaHeaders != lua.LNil {
		L.ForEach(luaHeaders.(*lua.LTable), func(k, v lua.LValue) {
			key := string(k.(lua.LString))
			if arr, ok := v.(*lua.LTable); ok {
				arr.ForEach(func(_, val lua.LValue) {
					respHeaders.Add(key, string(val.(lua.LString)))
				})
			} else {
				respHeaders.Set(key, string(v.(lua.LString)))
			}
		})
	}

	body := L.GetField(respTable, "Body")
	var bodyBytes []byte
	switch b := body.(type) {
	case lua.LString:
		bodyBytes = []byte(b)
	case *lua.LTable:
		var parts []string
		b.ForEach(func(_, v lua.LValue) {
			parts = append(parts, string(v.(lua.LString)))
		})
		bodyBytes = []byte(strings.Join(parts, ""))
	default:
		bodyBytes = nil
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
