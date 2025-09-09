package script

import (
	"context"
	"main/util"

	"github.com/dop251/goja"
)

// Redis_set (group, key, value)
func Redis_set(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	var group string
	var key string
	var value string
	// var outputBuffer strings.Builder
	if len(call.Arguments) < 2 {
		return goja.Undefined(), nil
	} else if len(call.Arguments) == 2 {
		key = call.Arguments[0].String()
		value = call.Arguments[1].String()
		util.RedisData.Set(key, value)
	} else {
		group = call.Arguments[0].String()
		key = call.Arguments[1].String()
		value = call.Arguments[2].String()
		util.RedisData.SetHValue(group, key, value)
	}

	return goja.Undefined(), nil
}

// Redis_get (group, key) -> string
func Redis_get(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	var group string
	var key string
	var value string
	var err error
	if len(call.Arguments) < 2 {
		return goja.Undefined(), nil
	} else if len(call.Arguments) == 2 {
		group = call.Arguments[0].String()
		key = call.Arguments[1].String()
		value, err = util.RedisData.GetHValue(group, key)
		if err != nil {
			return goja.Undefined(), nil
		}
	} else {
		key = call.Arguments[0].String()
		value, err = util.RedisData.Get(key)
		if err != nil {
			return goja.Undefined(), nil
		}
	}

	return rt.ToValue(value), nil
}

func Redis_keys(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return goja.Undefined(), nil
	}
	group := call.Arguments[0].Export().(string)
	keys, err := util.RedisData.HKeys(group)
	if err != nil {
		return rt.ToValue(""), nil
	}
	return rt.ToValue(keys), nil
}

func Redis_hgetall(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return goja.Undefined(), nil
	}
	group := call.Arguments[0].Export().(string)
	cmd := util.RedisData.Client.HGetAll(context.Background(), group)
	if cmd.Err() != nil {
		return rt.ToValue(""), nil
	}
	return rt.ToValue(cmd.Val()), nil
}
