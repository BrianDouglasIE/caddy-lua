local req = __LOOT_REQ
local res = __LOOT_RES
local next = __LOOT_NXT
local env = __LOOT_ENV
local util = __LOOT_UTL
local server_info = __LOOT_SRV

if req.Method == "GET" and req.URL == "/app" then
    res.Status = 200
    res.Header = { ["Content-Type"] = { "application/json" } }
    res.Body = util.json_encode({["message"] = "Hello from " .. server_info.Hostname})
else
    res.Status = 404
    res.Body = '{"error": "Not Found"}'
end