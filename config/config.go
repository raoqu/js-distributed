package config

import (
	"encoding/json"
	"fmt"
	"log"
	"main/util/color"
	"main/util/config"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DEFAULT_SCRIPT_GROUP_NAME = "default_scripts"

type AppConfig struct {
	Title string `yaml:"title"`
}

type Config struct {
	Initialized  bool
	MySQLConfigs map[string]MySQLConfig `yaml:"-"` // Map of MySQL configs by name

	// Grouped structure for YAML
	App       AppConfig          `yaml:"app"`
	Nacos     config.NacosConfig `yaml:"nacos"`
	Database  DatabaseConfig     `yaml:"database"`
	Messaging MessagingConfig    `yaml:"messaging"`
	Web       WebConfig          `yaml:"web"`
	Script    ScriptConfig       `yaml:"script"`
}

// syncFlatAndGrouped synchronizes between flat and grouped structures
func (c *Config) syncFlatAndGrouped() {
	// Initialize MySQLConfigs map if needed
	if c.MySQLConfigs == nil {
		c.MySQLConfigs = make(map[string]MySQLConfig)
	}

	// If default config found, use it for backward compatibility
	if len(c.Database.MySQLList) > 0 {
		// If no default config but we have at least one config, use the first one
		// and mark it as default
		if c.Database.MySQLList[0].Name == "" {
			c.Database.MySQLList[0].Name = "default"
		}
	}

	// Process all MySQL configurations
	for i, mysqlConfig := range c.Database.MySQLList {
		if mysqlConfig.Name != "" {
			c.MySQLConfigs[mysqlConfig.Name] = mysqlConfig
		} else {
			// Generate a name for unnamed configs
			name := fmt.Sprintf("mysql_%d", i)
			c.Database.MySQLList[i].Name = name
			c.MySQLConfigs[name] = c.Database.MySQLList[i]
		}
	}
}

type DatabaseConfig struct {
	MySQLList []MySQLConfig `yaml:"mysql"` // Multiple MySQL configs
	Redis     RedisConfig   `yaml:"redis"`
}

type InfluxDBConfig struct {
	URL        string        `yaml:"url,omitempty"`
	Token      string        `yaml:"token,omitempty"`
	Org        string        `yaml:"org,omitempty"`
	Bucket     string        `yaml:"bucket,omitempty"`
	Timeout    string        `yaml:"timeout,omitempty"`
	TimeoutVal time.Duration `yaml:"-"`
}

// RedisConfig holds Redis connection details
type RedisConfig struct {
	Addr     string `yaml:"addr,omitempty"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db,omitempty"`
	DBConfig int    `yaml:"dbConfig,omitempty"`
	Enable   bool   `yaml:"enable,omitempty"`
}

// MessagingConfig groups all messaging-related configurations
type MessagingConfig struct {
	// MQTT configurations
	MQTT MQTTConfig `yaml:"mqtt"`
}

// MQTTConfig holds MQTT connection details
type MQTTConfig struct {
	Broker   string `yaml:"broker,omitempty"`
	ClientID string `yaml:"clientId,omitempty"`
}

// WebConfig holds web server configurations
type WebConfig struct {
	Port   int    `yaml:"port,omitempty"`
	Static string `yaml:"static,omitempty"`
	Enable bool   `yaml:"enable,omitempty"`
}

// ScriptConfig holds script-related configurations
type ScriptConfig struct {
	GroupName string `yaml:"groupName,omitempty"`
	Endpoint  string `yaml:"endpoint,omitempty"`
	Dir       string `yaml:"dir,omitempty"`
}

// DefaultConfig provides a default configuration.
func DefaultConfig() Config {
	// Default values, consider environment variables or flags for overrides
	config := Config{
		Initialized:  false,
		MySQLConfigs: make(map[string]MySQLConfig),
		Nacos: config.NacosConfig{
			ServerAddr: "",
			Port:       8848,
			Namespace:  "",
			Group:      "DEFAULT_GROUP",
			LogDir:     "./nacos",
		},
		Database: DatabaseConfig{
			MySQLList: []MySQLConfig{
				{
					Name:       "default",
					ConnString: "",
				},
			},
			Redis: RedisConfig{
				Addr:     "localhost:6379",
				Password: "", // No password by default
				DB:       0,  // Default Redis DB
				DBConfig: 10,
				Enable:   false, // Disabled by default
			},
		},
		Messaging: MessagingConfig{
			MQTT: MQTTConfig{
				Broker:   "",
				ClientID: "go_mqtt_client",
			},
		},
		Web: WebConfig{
			Port:   8080,
			Static: "static",
			Enable: true,
		},
		Script: ScriptConfig{
			GroupName: DEFAULT_SCRIPT_GROUP_NAME,
			Endpoint:  "",
			Dir:       "scripts",
		},
		App: AppConfig{
			Title: "Default",
		},
	}

	// Initialize the default MySQL config in the map
	config.MySQLConfigs["default"] = config.Database.MySQLList[0]

	return config
}

var CONFIG Config

// LoadConfig loads configuration from a file
func LoadConfig(configDir ...string) Config {
	if CONFIG.Initialized {
		return CONFIG
	}
	CONFIG = DefaultConfig()

	// Determine config directory
	dir := ""
	if len(configDir) > 0 && configDir[0] != "" {
		dir = configDir[0]
	}

	// Construct config file paths for both YAML and JSON
	yamlFileName := filepath.Join(dir, "config.yaml")

	// Try to load from YAML config file first, then fall back to JSON
	loaded := false

	// Try YAML first
	if _, err := os.Stat(yamlFileName); err == nil {
		file, err := os.Open(yamlFileName)
		if err == nil {
			defer file.Close()

			decoder := yaml.NewDecoder(file)
			if err := decoder.Decode(&CONFIG); err == nil {
				// After loading YAML, sync between flat and grouped structures
				CONFIG.syncFlatAndGrouped()
				log.Printf("Loaded configuration from %s", yamlFileName)
				loaded = true
			} else {
				log.Printf("%s: %v", "Warning: Could not parse YAML config file", err)
			}
		}
	}

	if !loaded {
		fmt.Printf("%s\n", color.Red("Warning: No valid config file found"))
		os.Exit(1)
	}

	// Process the loaded configuration
	processConfig()

	CONFIG.Initialized = true

	return CONFIG
}

// processConfig performs additional processing on the loaded configuration
func processConfig() {
	// Process MySQL configurations
	for i := range CONFIG.Database.MySQLList {
		if CONFIG.Database.MySQLList[i].ConnString != "" {
			// Parse connection string for each MySQL config
			CONFIG.Database.MySQLList[i] = CONFIG.Database.MySQLList[i].ParseMySQLConnString(CONFIG.Database.MySQLList[i].ConnString)
		}
	}

	// For backward compatibility, if MySQLConnString is set but not in MySQLList
	// if CONFIG.MySQLConnString != "" {
	// 	// Check if this connection string is already in the list
	// 	found := false
	// 	for _, cfg := range CONFIG.Database.MySQLList {
	// 		if cfg.ConnString == CONFIG.MySQLConnString {
	// 			found = true
	// 			break
	// 		}
	// 	}

	// 	// If not found, add it as a new config
	// 	if !found {
	// 		mysqlConfig := MySQLConfig{
	// 			Name:       "default",
	// 			ConnString: CONFIG.MySQLConnString,
	// 		}
	// 		mysqlConfig = mysqlConfig.ParseMySQLConnString(CONFIG.MySQLConnString)
	// 		CONFIG.Database.MySQLList = append(CONFIG.Database.MySQLList, mysqlConfig)
	// 	}
	// }

	// Sync between flat and grouped structures
	CONFIG.syncFlatAndGrouped()
}

func LoadConfigMap(configFile string) map[string]interface{} {
	configMap := map[string]interface{}{}

	if _, err := os.Stat(configFile); err != nil {
		fmt.Printf("%s %s\n", color.Red("Warning: Config file does not exist "), configFile)
		return configMap
	}

	// Read file content
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
		os.Exit(1)
	}

	// Check if it's a YAML or JSON file
	if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
		// Parse YAML
		if err := yaml.Unmarshal(data, &configMap); err != nil {
			log.Printf("%s: %v", color.Red("Warning: Could not parse YAML config file"), err)
			return configMap
		}
	} else {
		// Parse JSON
		if err := json.Unmarshal(data, &configMap); err != nil {
			log.Printf("%s: %v", color.Red("Warning: Could not parse JSON config file"), err)
			return configMap
		}
	}

	return configMap
}
