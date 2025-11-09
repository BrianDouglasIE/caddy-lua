local dump = require("loot/dump")
local config = require("loot/config")
local utils = require("loot/utils")

local methods = { "GET", "PATCH", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE" }

LinkedListNode = {}
LinkedListNode.__index = LinkedListNode

function LinkedListNode:new(value)
  local self = setmetatable({}, LinkedListNode)
  self.value = value
  self.next = nil
  return self
end

LinkedList = {}
LinkedList.__index = LinkedList

function LinkedList:new()
  local self = setmetatable({}, LinkedList)
  self.nodes = {}
  self.count = 0
  return self
end

function LinkedList:add(value)
  self.count = self.count + 1
  self.nodes[self.count] = LinkedListNode:new(value)
  if self.count > 1 then
    self.nodes[self.count - 1].next = self.nodes[self.count]
  end
end

function LinkedList:from(tbl)
  local list = LinkedList:new()
  for _, value in ipairs(tbl) do
    list:add(value)
  end
  return list
end

function LinkedList:merge_table(list, tbl)
  local new_list = LinkedList:new()
  local head = list.nodes[1]
  while head do
    new_list:add(head.value)
    head = head.next
  end
  for _, value in ipairs(tbl) do
    new_list:add(value)
  end
  return new_list
end

Router = {}
Router.__index = Router

function Router:new(base_path)
  local self = setmetatable({}, Router)
  self.base_path = config.base_path .. base_path or config.base_path
  self.routes = {}
  self.middleware = LinkedList:new()

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
    handlers = LinkedList:merge_table(self.middleware, handlers)
  })
end

function Router:use(global_middleware)
  self.middleware:add(global_middleware)
end

function Router:match(path, method)
  for _, route in ipairs(self.routes) do
    if route.method == method then
      -- https://pkg.go.dev/github.com/julienschmidt/httprouter#Router.Lookup
      local is_match, params = __loot_ext.match_route(path, route.pattern)
      if is_match then
        return route, params
      end
    end
  end
  return nil, nil
end

function Router:handle(route, params)
  local res = __loot_res or {}
  local req = __loot_req or {}
  req.params = params
  req.url = __loot_url or {}

  local handlers = route.handlers
  if handlers.count == 0 then return end

  local function run(node)
    if not node then return end
    node.value(req, res, function()
      run(node.next)
    end)
  end

  run(handlers.nodes[1])
end

return Router
