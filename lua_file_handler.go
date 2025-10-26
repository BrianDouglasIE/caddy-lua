package lute

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type LuaFileHandler struct {
	ScriptPath string `json:"script_path,omitempty"`
	scriptAbs  string
}

func (LuaFileHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.lua_file",
		New: func() caddy.Module { return new(LuaFileHandler) },
	}
}

func (h *LuaFileHandler) Provision(ctx caddy.Context) error {
	if h.ScriptPath == "" {
		return fmt.Errorf("lua: script_path is required")
	}

	abs, err := filepath.Abs(h.ScriptPath)
	if err != nil {
		return fmt.Errorf("lua: cannot resolve script path: %w", err)
	}
	if _, err = os.Stat(abs); os.IsNotExist(err) {
		return fmt.Errorf("lua: script file does not exist: %s", abs)
	}
	h.scriptAbs = abs
	return nil
}

func (h *LuaFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return luaServeHTTP(w, r, h.scriptAbs, true)
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
