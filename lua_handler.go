package lute

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type LuaHandler struct {
	LuaBlock string `json:"lua_block,omitempty"`
}

func (LuaHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.lua",
		New: func() caddy.Module { return new(LuaHandler) },
	}
}

func (h *LuaHandler) Provision(ctx caddy.Context) error {
	if h.LuaBlock == "" {
		return fmt.Errorf("lua: lua_block is required")
	}

	return nil
}

func (h *LuaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return luaServeHTTPBlock(w, r, h.LuaBlock)
}
