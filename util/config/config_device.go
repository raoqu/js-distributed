package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/puzpuzpuz/xsync/v4"
)

// DeviceConfig represents a device configuration
type DeviceConfig struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
	Type string `json:"type"`
	// Interval int    `json:"interval"` // in seconds
	SlaveID  int    `json:"slave_id,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Interval int    `json:"interval,omitempty"`
	Config   string `json:"config,omitempty"`
}

/**
 * device_name 是唯一的设备名称标识
 */
var DEVICES_CONFIG = NewDeviceConfiguration()

type DeviceConfiguration struct {
	DataIdConfigMap *xsync.Map[string, string]
	DataIdDeviceMap *xsync.Map[string, *DeviceConfig]
}

func NewDeviceConfiguration() *DeviceConfiguration {
	return &DeviceConfiguration{
		DataIdConfigMap: xsync.NewMap[string, string](),
		DataIdDeviceMap: xsync.NewMap[string, *DeviceConfig](),
	}
}

func (d *DeviceConfiguration) parseDeviceConfigs(dataId, data string) (*[]*DeviceConfig, error) {
	var configs []*DeviceConfig
	if strings.TrimSpace(data) == "" {
		return &[]*DeviceConfig{}, nil
	}
	err := json.Unmarshal([]byte(data), &configs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON from %s, %v", dataId, err)
	}
	return &configs, nil
}

func (d *DeviceConfiguration) UpdateDeviceConfig(dataId, configData string) error {
	var deviceConfigs *[]*DeviceConfig
	var err error
	if configData == "" {
		deviceConfigs = &[]*DeviceConfig{}
		log.Printf("config data is empty, %s, %s", dataId, configData)
	} else {
		if deviceConfigs, err = d.parseDeviceConfigs(dataId, configData); err != nil {
			log.Printf("Ignore invalid device config %s.", dataId)
			return err
		}
	}

	for _, config := range *deviceConfigs {
		d.DataIdDeviceMap.Store(config.Name, config)
		// 设备配置更新
		if CONFIG_CHANGE_HANDLERS.DeviceUpdateHandler != nil {
			CONFIG_CHANGE_HANDLERS.DeviceUpdateHandler(config)
		}
	}

	deviceConfigMap := buildDeviceConfigMap(deviceConfigs)

	if previousConfig, ok := d.DataIdConfigMap.Load(dataId); ok {
		if previousDeviceConfig, err := d.parseDeviceConfigs(dataId, previousConfig); err == nil && previousDeviceConfig != nil {
			for _, cfg := range *previousDeviceConfig {
				if _, ok := deviceConfigMap[cfg.Name]; !ok {
					d.DataIdDeviceMap.Delete(cfg.Name)
					// 设备配置删除
					if CONFIG_CHANGE_HANDLERS.DeviceRemoveHandler != nil {
						CONFIG_CHANGE_HANDLERS.DeviceRemoveHandler(cfg.Name)
					}
				}
			}
		}
	}
	d.DataIdConfigMap.Store(dataId, configData)

	return nil
}

func buildDeviceConfigMap(deviceConfigs *[]*DeviceConfig) map[string]*DeviceConfig {
	deviceConfigMap := make(map[string]*DeviceConfig)
	if deviceConfigs != nil {
		for _, cfg := range *deviceConfigs {
			deviceConfigMap[cfg.Name] = cfg
		}
	}
	return deviceConfigMap
}

func UpdateDeviceConfig(dataId, configData string) error {
	return DEVICES_CONFIG.UpdateDeviceConfig(dataId, configData)
}

// Filter by debug device if specified
func FilterDeviceConfigs(configs []*DeviceConfig) []*DeviceConfig {
	for _, cfg := range configs {
		typeConfig := GetDeviceTypeConfig(cfg.Type)
		if typeConfig == nil {
			log.Printf("Ignore \"%s\" for type not found", cfg.Name)
			continue
		}
	}

	for i := range configs {
		ApplyDeviceTypeConfig(configs[i])
	}
	return configs
}

func GetAllDeviceConfig() []*DeviceConfig {
	configMap := xsync.ToPlainMap(DEVICES_CONFIG.DataIdDeviceMap)
	configs := make([]*DeviceConfig, 0, len(configMap))
	for _, config := range configMap {
		configs = append(configs, config)
	}

	return configs
}

func GetDeviceConfig(deviceId string) *DeviceConfig {
	val, ok := DEVICES_CONFIG.DataIdDeviceMap.Load(deviceId)
	if !ok {
		return nil
	}
	return val
}

func (c *DeviceConfig) Compare(other *DeviceConfig) bool {
	return c.Name == other.Name && c.Type == other.Type && c.Interval == other.Interval && c.SlaveID == other.SlaveID && c.Tags == other.Tags && c.IP == other.IP && c.Port == other.Port
}
