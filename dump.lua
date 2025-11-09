local colors = {
    reset  = "\27[0m",
    key    = "\27[90m",  -- gray
    string = "\27[32m",  -- green
    number = "\27[33m",  -- yellow
    bool   = "\27[36m",  -- cyan
    nilval = "\27[31m",  -- red
    table  = "\27[35m",  -- magenta
    func   = "\27[95m",  -- purple
}

function dump(...)
    local seen = {}

    local MAX_DEPTH = 6 
    local MAX_ITEMS = 50

    local function dump(value, indent, depth)
        indent = indent or 0
        depth  = depth or 0
        local pad = string.rep(" ", indent)

        if depth > MAX_DEPTH then
            return colors.table .. "..." .. colors.reset
        end

        local t = type(value)
        if t == "table" then
            if seen[value] then
                return colors.table .. "<circular reference>" .. colors.reset
            end
            seen[value] = true

            local lines = { colors.table .. "{" .. colors.reset }

            local keys = {}
            for k in pairs(value) do table.insert(keys, k) end
            table.sort(keys, function(a, b)
                if type(a) == "number" and type(b) == "number" then
                    return a < b
                end
                return tostring(a) < tostring(b)
            end)

            local count = 0
            for _, k in ipairs(keys) do
                count = count + 1
                if count > MAX_ITEMS then
                    table.insert(lines, pad .. "  " .. colors.table .. "... (" .. #keys - MAX_ITEMS .. " more)" .. colors.reset)
                    break
                end

                local v = value[k]
                local keyStr = colors.key .. "[" .. tostring(k) .. "]" .. colors.reset .. " = "
                table.insert(lines, pad .. "  " .. keyStr .. dump(v, indent + 2, depth + 1))
            end

            table.insert(lines, colors.table .. pad .. "}" .. colors.reset)
            return table.concat(lines, "\n")

        elseif t == "string" then
            return colors.string .. string.format("%q", value) .. colors.reset
        elseif t == "number" then
            return colors.number .. tostring(value) .. colors.reset
        elseif t == "boolean" then
            return colors.bool .. tostring(value) .. colors.reset
        elseif t == "nil" then
            return colors.nilval .. "nil" .. colors.reset
        elseif t == "function" then
            return colors.func .. "<function: " .. tostring(value) .. ">" .. colors.reset
        elseif t == "userdata" then
            return colors.func .. "<userdata: " .. tostring(value) .. ">" .. colors.reset
        elseif t == "thread" then
            return colors.func .. "<thread: " .. tostring(value) .. ">" .. colors.reset
        else
            return pad .. tostring(value)
        end
    end

    local args = {...}
    for i, v in ipairs(args) do
        print(colors.key .. "▶ d(" .. i .. "):" .. colors.reset, dump(v, 0, 0))
    end
end

function dd(...)
    dump(...)
    os.exit()
end

function ddd(...)
    dump(...)
    print("\nEntering interactive debug mode (type 'cont' or 'quit' to exit)...\n")
    if debug and debug.debug then
        debug.debug()
    else
        print("debug library not available.")
    end
    os.exit()
end

return dump