local utils = require("loot/utils")

Request = {}
Request.__index = Request

function Request:new(params)
  local self = setmetatable({}, Request)
  self.params = params
  self.url = __loot_url
  utils.table_merge(self, __loot_req)
  return self
end

return Request

