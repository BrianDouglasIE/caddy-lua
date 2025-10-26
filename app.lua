local req = __CADDY_REQUEST
local res = __CADDY_RESPONSE

if req.Method == "GET" and req.URL == "/app" then
    res.Status = 200
    res.Header = { ["Content-Type"] = { "application/json" } }
    res.Body = '{"message": "Hello from __CADDY globals!"}'
else
    res.Status = 404
    res.Body = '{"error": "Not Found"}'
end