package config

type ConfigChangeHandler func(id string)

type ConfigChangeHandlers struct {
	DeviceRemoveHandler     func(string)
	DeviceUpdateHandler     func(*DeviceConfig)
	DeviceTypeUpdateHandler func(*DeviceTypeConfig)
	ProtocolUpdateHandler   func(string, string)
	DictionaryUpdateHandler func(string, string)
}

var CONFIG_CHANGE_HANDLERS = ConfigChangeHandlers{}

func SetConfigChangeHandlers(handlers ConfigChangeHandlers) {
	CONFIG_CHANGE_HANDLERS = handlers
}
