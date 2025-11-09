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

function utils.table_contains(haystack, needle)
    for _, v in ipairs(haystack) do
        if v == needle then
            return true
        end
    end
    return false
end

function utils.string_startswith (haystack, needle)
    if #needle > #haystack then return false end

    return string.sub(haystack, 0, #needle) == needle
end

function utils.string_endswith (haystack, needle)
    if #needle > #haystack then return false end

    return string.sub(haystack, #haystack - (#needle - 1), #haystack) == needle
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

    assert(utils.table_contains({ 1, 2, 3 }, 2))
    assert(utils.table_contains({ 1, 2, 3 }, 4) == false)

    assert(utils.string_endswith("test", "st"))
    assert(utils.string_endswith("test", "s") == false)

    assert(utils.string_startswith("test", "tes"))
    assert(utils.string_startswith("test", "s") == false)

    print("All tests passed!")
end

run_tests()

return utils