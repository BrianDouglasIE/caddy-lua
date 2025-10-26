# Lute

Use Lua within Caddy to create web apps and script middleware.

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
                local req = __CADDY_REQUEST
                local res = __CADDY_RESPONSE
                res.Status = 200
                res.Body = "Hi from global __CADDY vars!"
            `
        }
    }
}
```

```lua
-- app.lua
local req = __CADDY_REQUEST
local res = __CADDY_RESPONSE
local next = __CADDY_NEXT
local env = __CADDY_ENV
local util = __CADDY_UTIL
local server_info = __CADDY_SERVER_INFO

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

Lute makes the following global variables available to your Lua code.

### __CADDY_REQUEST

This table holds the following `*http.Request` values.

 - Method
 - URL
 - Proto
 - Host
 - RemoteAddr
 - Header

### __CADDY_RESPONSE

This table will be used to build Caddy's response. It contains the following values.

 - Status
 - Header
 - Body

### __CADDY_NEXT

This method allows lua code to be used as Caddy middleware. It is a reference to the
`next caddyhttp.Handler` from the `ServeHttp` method.

#### Example Middleware

```
handle /app* {
    lua `
        print("Route is : " .. __CADDY_REQUEST.URL)
        __CADDY_NEXT() -- Must be called to continue to next directive
    `
    lua_file ./app.lua
}
```


### __CADDY_SERVER_INFO

This table holds info about the Caddy server. It contains the following fields.

 - Version
 - Module
 - Hostname
 - TLS

### __CADDY_ENV

This table contains the `os.Environ()` values that are available.

### __CADDY_UTIL

This table holds certain Golang methods that are useful to the Lua script.

For example `json_encode`, and `json_decode`.
