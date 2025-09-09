package config

import (
	"github.com/puzpuzpuz/xsync/v4"
)

type ProtocolConfiguration struct {
	ProtocolMap       xsync.Map[string, string]
	ModbusRegisterMap xsync.Map[string, []ModbusRegister]
}

var (
	PROTOCOL_CONFIG *ProtocolConfiguration = NewProtocolConfiguration()
)

func NewProtocolConfiguration() *ProtocolConfiguration {
	configuration := &ProtocolConfiguration{
		ProtocolMap:       *xsync.NewMap[string, string](),
		ModbusRegisterMap: *xsync.NewMap[string, []ModbusRegister](),
	}
	return configuration
}

func UpdateProtocolConfig(csvName, data string) error {
	if modbusRegisters, err := ParseModbusProtocol(data); err != nil {
		return err
	} else {
		PROTOCOL_CONFIG.ProtocolMap.Store(csvName, data)
		PROTOCOL_CONFIG.ModbusRegisterMap.Store(csvName, modbusRegisters)
		// 协议配置更新
		if CONFIG_CHANGE_HANDLERS.ProtocolUpdateHandler != nil {
			CONFIG_CHANGE_HANDLERS.ProtocolUpdateHandler(csvName, data)
		}
		return nil
	}
}
