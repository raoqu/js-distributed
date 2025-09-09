package script

import (
	"main/util"
	"main/util/strings"

	"github.com/dop251/goja"
)

// Redis_sadd adds one or more members to a set stored at key
// Usage: redis.sadd(key, member1, [member2, ...])
func Redis_sadd(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 2 {
		return goja.Undefined(), nil
	}

	key := call.Arguments[0].String()

	// Convert remaining arguments to interface{} slice
	members := make([]interface{}, len(call.Arguments)-1)
	for i := 1; i < len(call.Arguments); i++ {
		members[i-1] = strings.ToString(call.Arguments[i].Export())
	}

	count, err := util.RedisData.SAdd(key, members...)
	if err != nil {
		return rt.ToValue(0), nil
	}

	return rt.ToValue(count), nil
}

// Redis_srem removes one or more members from a set stored at key
// Usage: redis.srem(key, member1, [member2, ...])
func Redis_srem(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 2 {
		return goja.Undefined(), nil
	}

	key := call.Arguments[0].String()

	// Convert remaining arguments to interface{} slice
	members := make([]interface{}, len(call.Arguments)-1)
	for i := 1; i < len(call.Arguments); i++ {
		members[i-1] = strings.ToString(call.Arguments[i].Export())
	}

	count, err := util.RedisData.SRem(key, members...)
	if err != nil {
		return rt.ToValue(0), nil
	}

	return rt.ToValue(count), nil
}

// Redis_scard returns the number of elements in the set stored at key
// Usage: redis.scard(key)
func Redis_scard(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return goja.Undefined(), nil
	}

	key := call.Arguments[0].String()

	count, err := util.RedisData.SCard(key)
	if err != nil {
		return rt.ToValue(0), nil
	}

	return rt.ToValue(count), nil
}

// Redis_smembers returns all the members of the set value stored at key
// Usage: redis.smembers(key)
func Redis_smembers(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return goja.Undefined(), nil
	}

	key := call.Arguments[0].String()

	members, err := util.RedisData.SMembers(key)
	if err != nil {
		return rt.ToValue([]string{}), nil
	}

	return rt.ToValue(members), nil
}
