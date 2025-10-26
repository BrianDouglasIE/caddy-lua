package lute

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

var (
	_ caddy.Provisioner           = (*LuaHandler)(nil)
	_ caddyhttp.MiddlewareHandler = (*LuaHandler)(nil)
	_ caddyfile.Unmarshaler       = (*LuaHandler)(nil)

	_ caddy.Provisioner           = (*LuaFileHandler)(nil)
	_ caddyhttp.MiddlewareHandler = (*LuaFileHandler)(nil)
	_ caddyfile.Unmarshaler       = (*LuaFileHandler)(nil)
)

func init() {
	caddy.RegisterModule(LuaHandler{})
	httpcaddyfile.RegisterHandlerDirective("lua", func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		return parseLuaDirective(h, "lua")
	})
	httpcaddyfile.RegisterDirectiveOrder("lua", httpcaddyfile.After, "header")

	caddy.RegisterModule(LuaFileHandler{})
	httpcaddyfile.RegisterHandlerDirective("lua_file", func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		return parseLuaDirective(h, "lua_file")
	})
	httpcaddyfile.RegisterDirectiveOrder("lua_file", httpcaddyfile.After, "header")
}

func parseLuaDirective(h httpcaddyfile.Helper, kind string) (caddyhttp.MiddlewareHandler, error) {
	switch kind {
	case "lua":
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

	case "lua_file":
		var handler LuaFileHandler
		for h.Next() {
			if !h.NextArg() {
				return nil, h.ArgErr()
			}
			handler.FilePath = h.Val()
			if h.NextArg() {
				return nil, h.ArgErr()
			}
		}
		return &handler, nil

	default:
		return nil, h.Errf("unsupported lua directive: %s", kind)
	}
}
