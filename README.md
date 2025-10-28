# LOOT

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
                local req = __LOOT_REQ
                local res = __LOOT_RES
                res.Status = 200
                res.Body = "Hi from LOOT"
            `
        }
    }
}
```

```lua
-- app.lua
local req = __LOOT_REQ
local res = __LOOT_RES
local next = __LOOT_NXT
local env = __LOOT_ENV
local util = __LOOT_UTL
local server_info = __LOOT_SVR

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

### __LOOT_REQ

This table holds the following `*http.Request` values.

 - Method
 - URL
 - Proto
 - Host
 - RemoteAddr
 - Header

### __LOOT_RES

This table will be used to build Caddy's response. It contains the following values.

 - Status
 - Header
 - Body

### __LOOT_NXT

This method allows lua code to be used as Caddy middleware. It is a reference to the
`next caddyhttp.Handler` from the `ServeHttp` method.

#### Example Middleware

```
handle /app* {
    lua `
        print("Route is : " .. __LOOT_REQ.URL)
        __LOOT_NXT() -- Must be called to continue to next directive
    `
    lua_file ./app.lua
}
```


### __LOOT_SVR

This table holds info about the Caddy server. It contains the following fields.

 - Version
 - Module
 - Hostname
 - TLS

### __LOOT_ENV

This table contains the `os.Environ()` values that are available.

### __LOOT_UTL

This table holds certain Golang methods that are useful to the Lua script.

For example `json_encode`, and `json_decode`.
