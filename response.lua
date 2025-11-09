Response = {}
Response.__index = Response

function Response:new(status, headers, body)
  local self = setmetatable({}, Response)
  self.status = status or 200
  self.headers = headers or {}
  self.body = body or ""
  return self
end

function Response:add_header(name, content)
  table.insert(self.headers, {[name] = content})
end

function Response:set_content_type(content_type)
  self.headers["Content-Type"] = content_type
end

function Response:send(body, status)
  self.status = status or 200
  self.body = body
end

function Response:text(body, status)
  self:set_content_type("text/plain")
  self:send(body, status)
end

function Response:html(body, status)
  self:set_content_type("text/html")
  self:send(body, status)
end

function Response:json(body, status)
  self:set_content_type("application/json")
  self:send(body, status)
end

return Response
