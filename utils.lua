local utils = {}

function utils.table_merge(a, b)
    local result = {}
    for i = 1, #a do
        table.insert(result, a[i])
    end
    for i = 1, #b do
        table.insert(result, b[i])
    end
    return result
end

function utils.table_equal(a, b)
    if #a ~= #b then return false end
    for i = 1, #a do
        if a[i] ~= b[i] then return false end
    end
    return true
end

-- =======================
-- Tests
-- =======================
local function run_tests()
    -- Test table_merge
    local merged = utils.table_merge({1}, {2, 3})
    assert(utils.table_equal(merged, {1, 2, 3}), "Test 1 failed")

    merged = utils.table_merge({10, 20}, {30})
    assert(utils.table_equal(merged, {10, 20, 30}), "Test 2 failed")

    merged = utils.table_merge({}, {1})
    assert(utils.table_equal(merged, {1}), "Test 3 failed")

    merged = utils.table_merge({}, {})
    assert(utils.table_equal(merged, {}), "Test 4 failed")

    -- Test table_equal
    assert(utils.table_equal({1,2,3}, {1,2,3}), "Test 5 failed")
    assert(not utils.table_equal({1,2,3}, {1,2}), "Test 6 failed")
    assert(not utils.table_equal({1,2,3}, {3,2,1}), "Test 7 failed")
    assert(utils.table_equal({}, {}), "Test 8 failed")

    print("All tests passed!")
end

run_tests()

utils.contains = function (haystack, needle)
    for _, v in ipairs(haystack) do
        if v == needle then
            return true
        end
    end
    return false
end

assert(utils.contains({ 1, 2, 3 }, 2))
assert(utils.contains({ 1, 2, 3 }, 4) == false)

utils.startswith = function (haystack, needle)
    if #needle > #haystack then return false end

    return string.sub(haystack, 0, #needle) == needle
end

assert(utils.startswith("test", "tes"))
assert(utils.startswith("test", "s") == false)

utils.endswith = function (haystack, needle)
    if #needle > #haystack then return false end

    return string.sub(haystack, #haystack - (#needle - 1), #haystack) == needle
end

assert(utils.endswith("test", "st"))
assert(utils.endswith("test", "s") == false)

utils.join_path = function (...)
    local sep = package.config:sub(1, 1)
    local parts = { ... }
    local cleaned = {}

    for i, part in ipairs(parts) do
        if i == 1 then
            part = part:gsub(sep .. "+$", "")
        else
            part = part:gsub("^" .. sep .. "+", "")
            part = part:gsub(sep .. "+$", "")
        end
        if part ~= "" then
            table.insert(cleaned, part)
        end
    end

    local result = table.concat(cleaned, sep)
    return result
end

assert(utils.join_path("a", "b", "c") == "a/b/c")
assert(utils.join_path("a/", "/b", "c") == "a/b/c")
assert(utils.join_path("/a/", "/b/", "/c/") == "/a/b/c")
assert(utils.join_path("a//", "//b", "///c") == "a/b/c")

return utils