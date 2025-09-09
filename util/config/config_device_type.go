package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
)

// DeviceTypeConfig represents configuration for a specific device type
type DeviceTypeConfig struct {
	TypeName    string `json:"-"`
	Interval    int    `json:"interval"`
	Timeout     int    `json:"timeout"`
	Retries     int    `json:"retries"`
	Description string `json:"description"`
	Bucket      string `json:"bucket,omitempty"`
	Config      string `json:"config,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Params      string `json:"params,omitempty"`
}

// DeviceTypesConfig represents the structure of the device_types.json file
type DeviceTypesConfig struct {
	DeviceTypes map[string]*DeviceTypeConfig `json:"device_types"`
}

type DeviceTypeConfiguration struct {
	DeviceTypes atomic.Value
}

var (
	DEVICE_TYPES_CONFIG *DeviceTypeConfiguration = NewDeviceTypeConfiguration()
)

func NewDeviceTypeConfiguration() *DeviceTypeConfiguration {
	configuration := &DeviceTypeConfiguration{
		DeviceTypes: atomic.Value{},
	}
	return configuration
}

func (d *DeviceTypeConfiguration) parseDeviceTypeConfig(dataId, configData string) (*DeviceTypesConfig, error) {
	var cfg DeviceTypesConfig
	if err := json.Unmarshal([]byte(configData), &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling device type config from %s, %v", dataId, err)
	}
	for k, v := range cfg.DeviceTypes {
		v.TypeName = k
		cfg.DeviceTypes[k] = v
	}

	return &cfg, nil
}

// LoadDeviceTypeConfigs loads device type configurations from the config/device_types.json file
func LoadDeviceTypeConfigs() error {
	configPath := filepath.Join("config", "device_types.json")

	// Read the JSON file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading device type config file: %v", err)
	}

	if cfg, err := DEVICE_TYPES_CONFIG.parseDeviceTypeConfig("", string(data)); err != nil {
		return err
	} else {
		DEVICE_TYPES_CONFIG.DeviceTypes.Store(cfg)
		return nil
	}
}

// GetDeviceTypeConfig returns the configuration for a specific device type
// If the device type is not found, returns nil
func GetDeviceTypeConfig(deviceType string) *DeviceTypeConfig {
	cfg := DEVICE_TYPES_CONFIG.DeviceTypes.Load().(*DeviceTypesConfig)
	if cfg == nil {
		log.Printf("Device type config not initialized")
		return nil
	}
	if config, exists := cfg.DeviceTypes[deviceType]; exists {
		return config
	}
	return nil
}

// ApplyDeviceTypeConfig applies device type configuration to a device config
// This applies the type settings to the device config, overriding any existing settings
func ApplyDeviceTypeConfig(deviceConfig *DeviceConfig) {
	typeConfig := GetDeviceTypeConfig(deviceConfig.Type)

	// Always apply the interval from device type configuration
	deviceConfig.Interval = typeConfig.Interval
}

// GetDeviceTypeConfigValue parses the config string in format "key1=val1,key2=val2,..."
// and returns the value for the specified key
// If the key is not found or config is empty, returns the default value
func GetDeviceTypeConfigValue(deviceType string, key string, defaultValue string) string {
	typeConfig := GetDeviceTypeConfig(deviceType)
	return GetPropertyValue(typeConfig.Config, key, defaultValue)
}

// GetDeviceTypeTagValue parses the tags string in format "key1=val1,key2=val2,..."
// and returns the value for the specified key
// If the key is not found or tags is empty, returns the default value
func GetDeviceTypeTagValue(deviceType string, key string, defaultValue string) string {
	typeConfig := GetDeviceTypeConfig(deviceType)
	return GetPropertyValue(typeConfig.Tags, key, defaultValue)
}

func UpdateDeviceTypeConfig(dataId, configData string) error {
	if cfg, err := DEVICE_TYPES_CONFIG.parseDeviceTypeConfig(dataId, configData); err != nil {
		return err
	} else {
		DEVICE_TYPES_CONFIG.DeviceTypes.Store(cfg)
		// 设备类型配置更新
		if CONFIG_CHANGE_HANDLERS.DeviceTypeUpdateHandler != nil {
			for _, typeConfig := range cfg.DeviceTypes {
				CONFIG_CHANGE_HANDLERS.DeviceTypeUpdateHandler(typeConfig)
			}
		}
		return nil
	}
}

func GetAllDeviceTypeConfig() []*DeviceTypeConfig {
	deviceTypeMap := DEVICE_TYPES_CONFIG.DeviceTypes.Load().(*DeviceTypesConfig).DeviceTypes
	// to []*DeviceTypeConfig
	configs := make([]*DeviceTypeConfig, 0, len(deviceTypeMap))
	for _, config := range deviceTypeMap {
		configs = append(configs, config)
	}
	return configs
}
