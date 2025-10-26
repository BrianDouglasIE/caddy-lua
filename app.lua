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