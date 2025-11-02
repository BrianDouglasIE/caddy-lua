require("./dump")
local Router = require("./router")

local req = __LOOTBOX_REQ
local res = __LOOTBOX_RES
local next = __LOOTBOX_NXT
local url = __LOOTBOX_URL
local env = __LOOTBOX_ENV
local util = __LOOTBOX_UTL
local server_info = __LOOTBOX_SRV

res.status = 404
res.body = '404 Not Found'

local BookRouter = Router:new("/app/books")

BookRouter:use(function (req, res, next) 
    print("HI FROM ROUTE LEVEL MIDDLEWARE")
    next()
end)

BookRouter:get("/", { function() print("home") end })

BookRouter:get("/:id", {
    function(req, res, next)
        print("middleware called")
        next()
    end,
    function(req, res, next)
        res.status = 200
        res.header = { ["Content-Type"] = "application/json" }
        res.body = util.json_encode({ ["message"] = "Book with id: " .. req.params.id })
    end
})

local routers = { BookRouter }

dump(BookRouter)

for index, value in ipairs(routers) do
    dump(url.pathname)
    local route, params = value:match(url.pathname, req.method)
    if route == nil then
        goto continue
    end
    value:handle(route, params)
    ::continue::
end
