package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v4"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

const (
	ROOT_CONFIG_SUFFIX = "-root.json"
)

type ConfigChangeCallback func(dataId, data, parentId string) error

type NacosClient struct {
	Namespace            string
	Group                string
	ConfigClient         config_client.IConfigClient
	Callback             ConfigChangeCallback
	SubscribedSubDataIds xsync.Map[string, *xsync.Map[string, struct{}]]
	Initialized          atomic.Value
	Mu                   sync.Mutex
}

var nacosClient *NacosClient = NewNacosClient()

func NewNacosClient() *NacosClient {
	client := &NacosClient{
		SubscribedSubDataIds: *xsync.NewMap[string, *xsync.Map[string, struct{}]](),
		Mu:                   sync.Mutex{},
	}
	client.Initialized.Store(false)
	return client
}

func (nc *NacosClient) CreateClient(callback ConfigChangeCallback, nacosConfig *NacosConfig) error {
	nc.Callback = callback
	nc.Namespace = nacosConfig.Namespace
	nc.Group = nacosConfig.Group

	clientConfig := constant.ClientConfig{
		NamespaceId:         nc.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              nacosConfig.LogDir + "/log",
		CacheDir:            nacosConfig.LogDir + "/cache",
		LogLevel:            "debug",
	}
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: nacosConfig.ServerAddr,
			Port:   nacosConfig.Port,
		},
	}

	var err error
	nc.ConfigClient, err = clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		msg := fmt.Sprintf("Failed to create Nacos client: %v", err)
		return errors.New(msg)
	}
	return nil
}

func (nc *NacosClient) SubscribeToDataIds(dataIds []string, userCallback ConfigChangeCallback, parentId string) {
	for _, subId := range dataIds {
		content, err := nc.ConfigClient.GetConfig(vo.ConfigParam{DataId: subId, Group: nc.Group})
		if err != nil {
			log.Printf("Failed to get initial config for sub-dataId %s: %v", subId, err)
			continue
		}
		userCallback(subId, content, parentId)

		subCallback := func(namespace, group, dataId, data string) {
			userCallback(dataId, data, parentId)
		}

		if err := nc.ConfigClient.ListenConfig(vo.ConfigParam{DataId: subId, Group: nc.Group, OnChange: subCallback}); err != nil {
			log.Printf("Failed to listen for config changes for sub-dataId %s: %v", subId, err)
		} else {
			var subscribedSubDataIds *xsync.Map[string, struct{}]
			var ok bool
			if subscribedSubDataIds, ok = nc.SubscribedSubDataIds.Load(parentId); !ok {
				subscribedSubDataIds = xsync.NewMap[string, struct{}]()
				nc.SubscribedSubDataIds.Store(parentId, subscribedSubDataIds)
			}
			subscribedSubDataIds.Store(subId, struct{}{})
		}
	}
}

func (nc *NacosClient) UnsubscribeFromDataIds(parentId string, dataIds []string) {
	for _, subId := range dataIds {
		if err := nc.ConfigClient.CancelListenConfig(vo.ConfigParam{DataId: subId, Group: nc.Group}); err != nil {
			log.Printf("Failed to cancel listener for %s: %v", subId, err)
		} else {
			log.Printf("Cancelled listener for %s", subId)

			var subscribedSubIds *xsync.Map[string, struct{}]
			var ok bool
			if subscribedSubIds, ok = nc.SubscribedSubDataIds.Load(parentId); !ok {
				subscribedSubIds = xsync.NewMap[string, struct{}]()
				nc.SubscribedSubDataIds.Store(parentId, subscribedSubIds)
			}
			subscribedSubIds.Delete(subId)
		}
	}
}

func (nc *NacosClient) ProcessSubConfig(dataId, data string, userCallback ConfigChangeCallback) {
	nc.Mu.Lock()
	defer nc.Mu.Unlock()

	var subDataIds []string
	if data != "" {
		if err := json.Unmarshal([]byte(data), &subDataIds); err != nil {
			log.Printf("Failed to unmarshal sub dataId list from %s: %v", dataId, err)
			return
		}
	} else {
		log.Printf("Warning: configuration for %s is empty. No sub-data-ids to process.", dataId)
	}

	if len(subDataIds) == 0 {
		log.Printf("root configuration '%s' does not contain any sub-data-ids", dataId)
	}

	newSubDataIdsSet := make(map[string]struct{})
	for _, subId := range subDataIds {
		newSubDataIdsSet[subId] = struct{}{}
	}

	var toUnsubscribe, toSubscribe []string

	var subscribedSubDataIds *xsync.Map[string, struct{}]
	var ok bool
	if subscribedSubDataIds, ok = nc.SubscribedSubDataIds.Load(dataId); !ok {
		subscribedSubDataIds = xsync.NewMap[string, struct{}]()
		nc.SubscribedSubDataIds.Store(dataId, subscribedSubDataIds)
	}

	subscribed := xsync.ToPlainMap(subscribedSubDataIds)
	for oldSubId := range subscribed {
		if _, exists := newSubDataIdsSet[oldSubId]; !exists {
			toUnsubscribe = append(toUnsubscribe, oldSubId)
		}
	}

	for subId := range newSubDataIdsSet {
		if _, exists := subscribedSubDataIds.Load(subId); !exists {
			toSubscribe = append(toSubscribe, subId)
		}
	}

	if len(toUnsubscribe) > 0 {
		nc.UnsubscribeFromDataIds(dataId, toUnsubscribe)
	}
	if len(toSubscribe) > 0 {
		nc.SubscribeToDataIds(toSubscribe, userCallback, dataId)
	}
}

func InitNacos(callback ConfigChangeCallback, nacosConfig *NacosConfig) error {
	err := nacosClient.CreateClient(callback, nacosConfig)
	if err != nil {
		return err
	}

	// 没有子级配置项的回调处理
	noneRootConfigCallback := func(namespace, group, dataId, data string) {
		log.Println("non root config ", dataId)
		callback(dataId, data, "")
		log.Println("end for non root config ", dataId)
	}

	// 有子级配置项，需要进一步订阅子主题变更
	rootConfigCallback := func(namespace, group, dataId, data string) {
		nacosClient.ProcessSubConfig(dataId, data, callback)
	}

	for _, dataId := range STARTUP_CONFIG_DATA_IDS {
		content, err := nacosClient.ConfigClient.GetConfig(vo.ConfigParam{DataId: dataId, Group: nacosClient.Group})
		if err != nil {
			return fmt.Errorf("failed to get initial root config for %s: %v", dataId, err)
		}

		listenCallback := noneRootConfigCallback
		if isRootDataId(dataId) {
			listenCallback = rootConfigCallback
			log.Println("-- Processing root config for", dataId)
			rootConfigCallback(nacosConfig.Namespace, nacosConfig.Group, dataId, content)
			log.Println("-- Processed root config for", dataId)
		} else {
			log.Println("-- Processing non-root config for", dataId)
			callback(dataId, content, "")
			log.Println("-- Processed non-root config for", dataId)
		}

		err = nacosClient.ConfigClient.ListenConfig(vo.ConfigParam{
			DataId:   dataId,
			Group:    nacosConfig.Group,
			OnChange: listenCallback,
		})
		if err != nil {
			return fmt.Errorf("failed to listen for root config changes for %s: %v", dataId, err)
		}
	}

	log.Println("Nacos configuration intialized")
	nacosClient.Initialized.Store(true)
	return nil
}

func isRootDataId(dataId string) bool {
	return strings.HasSuffix(dataId, ROOT_CONFIG_SUFFIX)
}

func CheckConfigReady() bool {
	return nacosClient.Initialized.Load().(bool)
}
