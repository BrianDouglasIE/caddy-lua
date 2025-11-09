Response = {}
Response.__index = Response

function Response:new(status, headers, body)
  local self = setmetatable({}, Response)
  self.status = status or 404
  self.headers = headers or {}
  self.body = body or ""
  return self
end

function Response:ok(body)
  self.status = 200
  self.body = body
end

return Response
