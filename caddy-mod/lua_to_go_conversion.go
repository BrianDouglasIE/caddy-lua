package LOOT

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func luaToGo(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(v)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(v))
	case lua.LTString:
		return lua.LVAsString(v)
	case lua.LTTable:
		tbl := v.(*lua.LTable)
		isArray := true
		var maxKey int
		tbl.ForEach(func(k, _ lua.LValue) {
			if key, ok := k.(lua.LNumber); ok {
				if int(key) > maxKey {
					maxKey = int(key)
				}
			} else {
				isArray = false
			}
		})
		if isArray {
			arr := make([]interface{}, maxKey)
			tbl.ForEach(func(k, v lua.LValue) {
				idx := int(k.(lua.LNumber))
				if idx > 0 && idx <= len(arr) {
					arr[idx-1] = luaToGo(v)
				}
			})
			return arr
		}
		obj := map[string]interface{}{}
		tbl.ForEach(func(k, v lua.LValue) {
			obj[k.String()] = luaToGo(v)
		})
		return obj
	default:
		return v.String()
	}
}

func goToLua(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case float64:
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case []interface{}:
		tbl := L.NewTable()
		for i, elem := range val {
			tbl.RawSetInt(i+1, goToLua(L, elem))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, elem := range val {
			tbl.RawSetString(k, goToLua(L, elem))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}
