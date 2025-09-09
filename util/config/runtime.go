package config

var (
	RUNTIME_TYPE           string
	DEBUG_DEVICES          string
	USE_NAME_IN_MQTT       bool
	VERBOSE_MODE           bool
	MQTT_CONFIGURABLE_PATH string = "data/"
	RUNTIME_LOCAL          bool
	DEMO_MODE              bool
)

const (
	DATA_ID_DEVICE_CONFIG      = "device-root.json"
	DATA_ID_DEVICE_TYPE_CONFIG = "device-types.json"
	DATA_ID_PROTOCOL_CONFIG    = "protocol-root.json"
	DATA_ID_DICT_CONFIG        = "dict-root.json"
)

var STARTUP_CONFIG_DATA_IDS = []string{
	DATA_ID_DEVICE_CONFIG,
	DATA_ID_DEVICE_TYPE_CONFIG,
	DATA_ID_PROTOCOL_CONFIG,
	DATA_ID_DICT_CONFIG,
}
