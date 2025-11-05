# loot

A tool for making websites with Lua, using the Caddy web server.

## Example Usage

```
:8080 {
    route {
        handle /app* {
            lua_file ./logger.middleware.lua
            lua_file ./app.lua
        }

        handle / {
            lua `
                local req = __loot_req
                local res = __loot_res
                res.Status = 200
                res.Body = "Hi from loot"
            `
        }
    }
}
```

```lua
-- app.lua
local req = __loot_req
local res = __loot_res
local next = __loot_next
local env = __loot_env
local util = __loot_ext
local server_info = __loot_SVR

if req.Method == "GET" and req.URL == "/app" then
    res.Status = 200
    res.Header = { ["Content-Type"] = { "application/json" } }
    res.Body = util.json_encode({["message"] = "Hello from " .. server_info.Hostname})
else
    res.Status = 404
    res.Body = '{"error": "Not Found"}'
end
```

## The Global Vars

loot makes the following global variables available to your Lua code.

### __loot_req

This table holds the following `*http.Request` values.

 - Method
 - URL
 - Proto
 - Host
 - RemoteAddr
 - Header

### __loot_res

This table will be used to build Caddy's response. It contains the following values.

 - Status
 - Header
 - Body

### __loot_next

This method allows lua code to be used as Caddy middleware. It is a reference to the
`next caddyhttp.Handler` from the `ServeHttp` method.

#### Example Middleware

```
handle /app* {
    lua `
        print("Route is : " .. __loot_req.URL)
        __loot_next() -- Must be called to continue to next directive
    `
    lua_file ./app.lua
}
```


### __loot_SVR

This table holds info about the Caddy server. It contains the following fields.

 - Version
 - Module
 - Hostname
 - TLS

### __loot_env

This table contains the `os.Environ()` values that are available.

### __loot_ext

This table holds certain Golang methods that are useful to the Lua script.

For example `json_encode`, and `json_decode`.
