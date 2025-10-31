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

Router = {}
Router.__index = Router

local methods = {"GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"}

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

  return self
end

function Router:add(method, pattern, handler)
  local parsed_pattern, params = parse_pattern(pattern)
  table.insert(self.routes, {
    method=method,
    pattern=parsed_pattern,
    params=params,
    handler=handler
  })
end

function Router:get(pattern, handler)
  self:add("GET", pattern, handler)
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
        return route.handler, params
      end
    end
  end
  return nil, nil
end


-- tests

local router = Router:new()
router:get("/books/:id", function (req, res, next)
  print(req.params[1])
end)
local handler, params = router:match("/books/42", "GET")
print(params.id)
print_table(params)
