local methods = {"GET", "PATCH", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

function print_table(t, indent)
  indent = indent or ""
  for k, v in pairs(t) do
    if type(v) == "table" then
      print(indent .. tostring(k) .. "=")
      print_table(v, indent .. "  ")
    else
      print(indent .. tostring(k) .. "=", v)
    end
  end
end

function contains(haystack, needle)
  for _, v in ipairs(haystack) do
    if v == needle then
      return true
    end
  end
  return false
end

local function make_handler_alias_methods(router)
  for _, method in ipairs(methods) do
    router[string.lower(method)] = function(self, ...)
      local args = {...}
      local arg_count = #args
      local pattern = args[1]
      local handler = nil
      local middleware = nil
      
      if arg_count == 2 then
        handler = args[2]
      elseif arg_count == 3 then
        middleware = args[2]
        handler = args[3]
      else
        error("Invalid argument count for " .. lower_name)
      end

      self:add(method, pattern, middleware, handler)
    end
  end
end

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

Router = {}
Router.__index = Router

local function parse_pattern(pattern)
  local params = {}
  local parsed = pattern:gsub(":(%w+)", function(name)
    table.insert(params, name)
    return "([^/]+)"
  end)
  parsed = "^" .. parsed .. "$"
  return parsed, params
end

function Router:new(url)
  local self = setmetatable({}, Router)
  self.routes = {}
  make_handler_alias_methods(self)
  return self
end

function Router:add(...)
  local args = {...}
  local arg_count = #args
  local method = args[1] 
  local pattern = args[2]
  local handler = nil
  local middleware = LinkedList:new()
  
  if arg_count == 3 then
    handler = args[3]
  elseif arg_count == 4 then
    for _, middleware_method in ipairs(args[3]) do
      middleware:add(middleware_method)
    end
    handler = args[4]
  end
  
  local parsed_pattern, params = parse_pattern(pattern)
  table.insert(self.routes, {
    method=method,
    pattern=parsed_pattern,
    params=params,
    handler=handler,
    middleware=middleware
  })
end

function Router:match(path, method)
  for _, route in ipairs(self.routes) do
    if route.method == method then
      local captures = {path:match(route.pattern)}
      if #captures > 0 then
        local params = {}
        for i, name in ipairs(route.params) do
          params[name] = captures[i]
        end
        return route, params
      end
    end
  end
  return nil, nil
end

function Router:handle(route, params)
  local res = __LOOTBOX_RES or {}
  local req = __LOOTBOX_REQ or {}
  req.params = params

  local handlers = route.middleware
  handlers:add(route.handler)

  if handlers.count == 0 then return end

  local function run(node)
    if not node then return end
    node.value(req, res, function()
      run(node.next)
    end)
  end

  run(handlers.nodes[1])
end



-- tests

local router = Router:new()
router:get("/books/:id", {
function (req, res, next)
  req.params.id = req.params.id + 1
  next()
end,
function (req, res, next)
  req.params.id = req.params.id + 1
  next()
end,
function (req, res, next)
  req.params.id = req.params.id + 1
  next()
end
}, function (req, res, next)
  print(req.params.id)
end)
local route, params = router:match("/books/42", "GET")
router:handle(route, params)
