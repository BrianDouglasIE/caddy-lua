local req = __LOOTBOX_REQ
local res = __LOOTBOX_RES
local next = __LOOTBOX_NXT
local env = __LOOTBOX_ENV
local util = __LOOTBOX_UTL
local server_info = __LOOTBOX_SRV

if req.Method == "GET" and req.URL == "/app" then
    res.Status = 200
    res.Header = { ["Content-Type"] = { "application/json" } }
    res.Body = util.json_encode({["message"] = "Hello from " .. server_info.Hostname})
else
    res.Status = 404
    res.Body = '{"error": "Not Found"}'
end
