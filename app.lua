function handle(req, resp)
    if req.Method == "GET" and req.URL == "/app" then
        resp.Status = 200
        resp.Header = { ["Content-Type"] = { "application/json" } }
        resp.Body = [[{"message":"Hello from FrankenLua + LuaJIT!"}]]
        return
    end

    if req.Method == "POST" and req.URL == "/app/echo" then
        resp.Status = 200
        resp.Header = { ["Content-Type"] = { "application/json" } }
        resp.Body = req.Body or [[{"message":"(no body)"}]]
        return
    end

    resp.Status = 404
    resp.Header = { ["Content-Type"] = { "application/json" } }
    resp.Body = [[{"error":"Not Found"}]]
end
