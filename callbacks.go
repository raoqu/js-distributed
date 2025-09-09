package main

import (
	"encoding/json"
	"log"
	"main/util"
	"main/util/config"
	"main/util/strings"
	gstrings "strings"
)

func onDeviceRemove(deviceName string) {
	log.Printf("Device removed: %s", deviceName)

	deviceType := getRealDeviceType(deviceName)
	util.RedisData.HDel("device_"+deviceType, deviceName)
}

func onDeviceUpdate(deviceConfig *config.DeviceConfig) {
	log.Printf("Device updated: %s", deviceConfig.Name)

	deviceType := getRealDeviceType(deviceConfig.Name)
	// DeviceConfig to json
	jsonStr, err := json.Marshal(deviceConfig)
	if err != nil {
		log.Printf("Failed to marshal device config: %v", err)
		return
	}
	err = util.RedisData.SetHValue("device_"+deviceType, deviceConfig.Name, string(jsonStr))
	if err != nil {
		log.Printf("Failed to set device config: %v", err)
	}
}

func onDeviceTypeUpdate(deviceTypeConfig *config.DeviceTypeConfig) {
	log.Printf("Device type updated: %s", deviceTypeConfig.TypeName)

	jsonStr, err := json.Marshal(deviceTypeConfig)
	if err != nil {
		log.Printf("Failed to marshal device type config: %v", err)
		return
	}
	err = util.RedisData.SetHValue("DEVICE_TYPE", deviceTypeConfig.TypeName, string(jsonStr))
	if err != nil {
		log.Printf("Failed to set device type config: %v", err)
	}
}

func onProtocolUpdate(csvName string, data string) {
	log.Printf("Protocol updated: %s", csvName)

	err := util.RedisData.SetHValue("DEVICE_PROTOCOL", csvName, data)
	if err != nil {
		log.Printf("Failed to set protocol config: %v", err)
	}
}

func onDictionaryUpdate(csvName string, data string) {
	log.Printf("Dictionary updated: %s", csvName)

	err := util.RedisData.SetHValue("DICT", csvName, data)
	if err != nil {
		log.Printf("Failed to set dictionary config: %v", err)
	}
}

// 获取设备类型
func getRealDeviceType(name string) string {
	return gstrings.Split(name, "_")[0]
}

// 获取设备协议配置key
func getDeviceProtocolConfigKey(deviceConfig *config.DeviceConfig) string {
	var deviceType = deviceConfig.Type
	var configKey = "unknown"
	var deviceTypeConfig = config.GetDeviceTypeConfig(deviceType)
	if deviceTypeConfig != nil {
		configKey = strings.GetPropertyValue(deviceTypeConfig.Config, "device_type", configKey)
	}
	return configKey
}
