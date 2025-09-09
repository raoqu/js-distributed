package main

import (
	"fmt"
	"log"
	"main/util"
	"main/util/mysql"
	"main/util/script"
	"sync"
)

// Global script pool for caching compiled scripts
var scriptPool *script.ScriptPool
var scriptInitOnce sync.Once

// initScriptPool initializes the global script pool if not already initialized
func initScriptPool(once *sync.Once, poolName string) {
	once.Do(func() {
		scriptPool = script.NewScriptPool(poolName, util.RedisConfig)

		// Inject console functions
		scriptPool.Inject("console.log", script.Console_log)
		scriptPool.Inject("console.error", script.Console_error)

		// Inject Redis functions
		scriptPool.Inject("redis.set", script.Redis_set)
		scriptPool.Inject("redis.get", script.Redis_get)
		scriptPool.Inject("redis.keys", script.Redis_keys)
		scriptPool.Inject("redis.hgetall", script.Redis_hgetall)

		// Inject Redis set operations
		scriptPool.Inject("redis.sadd", script.Redis_sadd)
		scriptPool.Inject("redis.srem", script.Redis_srem)
		scriptPool.Inject("redis.scard", script.Redis_scard)
		scriptPool.Inject("redis.smembers", script.Redis_smembers)

		// Inject MySQL functions
		if mysql.MYSQL_CLIENT != nil {
			scriptPool.Inject("mysql.query", script.MySQL_query)
			scriptPool.Inject("mysql.exec", script.MySQL_exec)
			scriptPool.Inject("mysql.queryRow", script.MySQL_queryRow)
			scriptPool.Inject("mysql.transaction", script.MySQL_transaction)
		}

		// Inject Net functions
		scriptPool.Inject("net.fetch", script.Net_fetch)

		// Inject Sys functions
		scriptPool.Inject("sys.command", script.Sys_command)

		log.Println("Script pool initialized with injected functions")
	})
}

var EnableScript = true

// executeJavaScript runs a JavaScript code using goja with ScriptPool for caching
func executeJavaScript(name string, ctx map[string]interface{}) (interface{}, error) {
	if !EnableScript {
		return "", fmt.Errorf("script disabled")
	}

	code, err := scriptPool.Cache.GetScript(name)
	if err != nil {
		return "", fmt.Errorf("failed to get script: %s %v", name, err)
	}

	// Set the script in the pool (compiles and caches it)
	err = scriptPool.SetScript(name, code)
	if err != nil {
		return "", fmt.Errorf("failed to compile script: %v", err)
	}

	// Run the script from the pool
	result, err := scriptPool.RunScript(name, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to run script: %v", err)
	}

	return result.Value, nil
}
