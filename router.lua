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

function startswith(haystack, needle)
  if #needle > #haystack then return false end

  return string.sub(haystack, 0, #needle) == needle
end

assert(startswith("test", "tes"))
assert(startswith("test", "s") == false)

function endswith(haystack, needle)
  if #needle > #haystack then return false end

  return string.sub(haystack, #haystack - (#needle - 1), #haystack) == needle
end

assert(endswith("test", "st"))
assert(endswith("test", "s") == false)

function join_path(...)
  local result = ""

  local segments = { ... }
  for _, segment in ipairs(segments) do
    if endswith(segment, "/") then
      result = result .. segment
    else
      result = result .. "/" .. segment
    end
  end

  return result
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

function LinkedList:from(tbl)
  local list = LinkedList:new()
  for _, value in ipairs(tbl) do
    list:add(value)
  end
  return list
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

function Router:new(base_path)
  local self = setmetatable({}, Router)
  self.base_path = base_path or ""
  self.routes = {}

  for _, method in ipairs(methods) do
    self[string.lower(method)] = function(self, pattern, handlers)
      self:add(method, pattern, handlers)
    end
  end

  return self
end

function Router:add(method, pattern, handlers)
  local parsed_pattern, params = parse_pattern(pattern)
  table.insert(self.routes, {
    method=method,
    pattern=parsed_pattern,
    params=params,
    handlers=LinkedList:from(handlers)
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



-- tests

local router = Router:new()

router:get("/", { function () print("home") end })

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
  end,
  function (req, res, next)
    print(req.params.id)
  end
})

local home, home_params = router:match("/", "GET")
router:handle(home, home_params)

local route, params = router:match("/books/42", "GET")
router:handle(route, params)
