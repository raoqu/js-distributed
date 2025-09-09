package config

import (
	"github.com/puzpuzpuz/xsync/v4"
)

type DictionaryConfiguration struct {
	DictionaryMap xsync.Map[string, string]
}

var (
	DICTIONARY_CONFIG *DictionaryConfiguration = NewDictionaryConfiguration()
)

func NewDictionaryConfiguration() *DictionaryConfiguration {
	configuration := &DictionaryConfiguration{
		DictionaryMap: *xsync.NewMap[string, string](),
	}
	return configuration
}

func UpdateDictionaryConfig(csvName, data string) error {
	DICTIONARY_CONFIG.DictionaryMap.Store(csvName, data)
	// 协议配置更新
	if CONFIG_CHANGE_HANDLERS.DictionaryUpdateHandler != nil {
		CONFIG_CHANGE_HANDLERS.DictionaryUpdateHandler(csvName, data)
	}
	return nil
}
