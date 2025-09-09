package main

import (
	"fmt"
	"log"
	cfg "main/config"
	"main/util"
	"main/util/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

var SCRIPT_MANAGER *ScriptManager

func main() {

	fileConfig := cfg.LoadConfig(".")

	config.SetConfigChangeHandlers(config.ConfigChangeHandlers{
		DeviceRemoveHandler:     onDeviceRemove,
		DeviceUpdateHandler:     onDeviceUpdate,
		DeviceTypeUpdateHandler: onDeviceTypeUpdate,
		ProtocolUpdateHandler:   onProtocolUpdate,
		DictionaryUpdateHandler: onDictionaryUpdate,
	})

	err := util.InitRedisClient()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		return
	}

	if _, err := initializeDeviceConfigs(); err != nil {
		fmt.Printf("Error loading device configs: %v\n", err)
		return
	}

	initScriptPool(&scriptInitOnce, cfg.CONFIG.Script.GroupName)

	// Initialize web server if enabled
	var httpServer *http.Server
	var webConfig = cfg.CONFIG.Web
	var scriptConfig = cfg.CONFIG.Script
	if webConfig.Enable {
		log.Printf("Starting web server on port %d...", webConfig.Port)

		// Create Gin router
		router := gin.Default()
		// Setup HTML template rendering
		router.LoadHTMLGlob(webConfig.Static + "/*.html")

		router.Static("/static", webConfig.Static)
		router.GET("/favicon.ico", func(c *gin.Context) {
			c.File(webConfig.Static + "/favicon.ico")
		})

		SCRIPT_MANAGER = NewScriptManager(&fileConfig, scriptPool)
		if scriptConfig.Endpoint != "" {
			SetupScriptsRoutes(router, &fileConfig, SCRIPT_MANAGER)
		}

		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"ScriptEndpoint": scriptConfig.Endpoint,
				"AppTitle":       cfg.CONFIG.App.Title,
			})
		})

		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", webConfig.Port),
			Handler: router,
		}

		// Start HTTP server in a goroutine
		func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server error: %v", err)
			}
		}()

		log.Printf("Web server started: http://127.0.0.1:%d", webConfig.Port)
	} else {
		log.Printf("Web server is disabled")
	}

}

func initializeDeviceConfigs() ([]*config.DeviceConfig, error) {
	var deviceConfigs []*config.DeviceConfig
	// Initialize Nacos client and listener
	if err := config.InitNacos(onConfigChange, &cfg.CONFIG.Nacos); err != nil {
		log.Println("Failed to initialize Nacos client. Exiting.", err)
		return nil, err
	}

	// Load device type configurations
	// err := config.LoadDeviceTypeConfigs()
	if !config.CheckConfigReady() {
		fmt.Printf("Warning: Failed to load device type configurations")
		return nil, fmt.Errorf("failed to load device type configurations")
	}

	deviceConfigs = config.GetAllDeviceConfig()

	return deviceConfigs, nil
}
