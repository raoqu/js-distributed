package mysql

import (
	"main/config"
	"sync"
)

var (
	MYSQL_CLIENT  *MySQLClient
	MYSQL_CLIENTS map[string]*MySQLClient
	once          sync.Once
)

func GetConfig(name string) (config.MySQLConfig, bool) {
	if name == "" {
		name = "default"
	}

	if !config.CONFIG.Initialized {
		config.LoadConfig("") // Load default config if not already loaded
	}

	config, exists := config.CONFIG.MySQLConfigs[name]
	return config, exists
}

func GetClient(name string) *MySQLClient {
	if name == "" || name == "default" {
		return MYSQL_CLIENT
	}

	return MYSQL_CLIENTS[name]
}
