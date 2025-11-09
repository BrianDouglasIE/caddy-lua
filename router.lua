local Request = require("loot/request")
local Response = require("loot/response")
local dump = require("loot/dump")
local config = require("loot/config")
local utils = require("loot/utils")

local methods = { "GET", "PATCH", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE" }

Router = {}
Router.__index = Router

function Router:new(base_path)
  local self = setmetatable({}, Router)
  self.base_path = config.base_path .. base_path or config.base_path
  self.routes = {}
  self.middleware = {}

  for _, method in ipairs(methods) do
    self[string.lower(method)] = function(self, pattern, handlers)
      self:add(method, pattern, handlers)
    end
  end

  return self
end

function Router:add(method, pattern, handlers)
  table.insert(self.routes, {
    method = method,
    pattern = self.base_path .. pattern,
    handlers = utils.table_merge(self.middleware, handlers)
  })
  return self
end

function Router:use(global_middleware)
  table.insert(self.middleware, global_middleware)
  return self
end

function Router:match(path, method)
  for _, route in ipairs(self.routes) do
    if route.method == method then
      local is_match, params = __loot_ext.match_route(path, route.pattern)
      if is_match then
        return route, params
      end
    end
  end
  return nil, nil
end

function Router:handle(route, params)
  local req = Request:new(params)
  local res = Response:new(__loot_res.status, __loot_res.headers, __loot_res.body)

  local handlers = route.handlers
  if not #handlers then return end

  local co = coroutine.create(function()
    for i = 1, #handlers do
      handlers[i](req, res, function() 
        coroutine.yield()
      end)
    end
  end)

  while coroutine.status(co) ~= "dead" do
    coroutine.resume(co)
  end

  __loot_res.status = res.status
  __loot_res.headers = res.headers
  __loot_res.body = res.body
end

return Router
