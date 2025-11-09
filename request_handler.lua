local dump = require("loot/dump")
local config = require("loot/config")

return function ()
  for _, router in ipairs(config.routers) do
    local route, params = router:match(__loot_req.url, __loot_req.method)
    if route then 
      router:handle(route, params)
      break
    end
  end
end
