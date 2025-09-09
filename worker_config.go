package main

import (
	"log"
	"main/util/config"
)

func onConfigChange(dataId, data, parentId string) error {
	log.Printf("Config update: %s, %s", dataId, parentId)

	if parentId == config.DATA_ID_DEVICE_CONFIG {
		if err := config.UpdateDeviceConfig(dataId, data); err != nil {
			log.Printf("Failed to update device config: %v", err)
		} else {
			ApplyDeviceConfiguration()
		}
		return nil
	} else if dataId == config.DATA_ID_DEVICE_TYPE_CONFIG {
		if err := config.UpdateDeviceTypeConfig(dataId, data); err != nil {
			log.Printf("Failed to update device type config: %v", err)
		} else {
			ApplyDeviceConfiguration()
		}
	} else if parentId == config.DATA_ID_PROTOCOL_CONFIG {
		if err := config.UpdateProtocolConfig(dataId, data); err != nil {
			log.Printf("Failed to update protocol config: %v", err)
		}
	} else if parentId == config.DATA_ID_DICT_CONFIG {
		if err := config.UpdateDictionaryConfig(dataId, data); err != nil {
			log.Printf("Failed to update dictionary config: %v", err)
		}
	}
	return nil
}

func ApplyDeviceConfiguration() {
	if !config.CheckConfigReady() {
		return
	}
	log.Printf("Apply device configuration")

	deviceConfigs := config.GetAllDeviceConfig()
	deviceConfigs = config.FilterDeviceConfigs(deviceConfigs)

	//TODO
}
