package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"time"

	cfg "main/config"
	"main/util/script"

	"github.com/gin-gonic/gin"
)

// ScriptItem represents a task script with its metadata
type ScriptItem struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// ScriptManager manages HTTP requests for script operations
type ScriptManager struct {
	config     *cfg.Config
	ScriptPool *script.ScriptPool
}

// NewScriptManager creates a new ScriptManager instance
func NewScriptManager(config *cfg.Config, scPool *script.ScriptPool) *ScriptManager {
	return &ScriptManager{
		config:     config,
		ScriptPool: scPool,
	}
}

func (h *ScriptManager) Execute(c *gin.Context) {
	taskName := c.Param("taskname")

	if taskName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task name is required",
		})
		return
	}

	h.executeInner(c, taskName, false)
}

func (h *ScriptManager) ExecutePost(c *gin.Context) {
	taskName := c.Param("taskname")

	if taskName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task name is required",
		})
		return
	}

	h.executeInner(c, taskName, true)
}

func (h *ScriptManager) executeInner(c *gin.Context, taskName string, isPost bool) {
	scriptCache := h.ScriptPool.Cache
	exists, err := scriptCache.ScriptExists(taskName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Error checking script existence: %v", err),
		})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Task script '%s' not found", taskName),
		})
		return
	}

	var params map[string]interface{}
	if isPost {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Invalid request body: %v", err),
			})
			return
		}
		params = make(map[string]interface{})
		params["request"] = string(bodyBytes)
	} else {
		query := c.Request.URL.Query()
		if query != nil {
			params = make(map[string]interface{})
			for key, value := range query {
				params[key] = value[0]
			}
		}
	}

	// Execute script
	startTime := time.Now()

	result, err := executeJavaScript(taskName, params)

	elapsedTime := time.Since(startTime)
	elapsedMs := float64(elapsedTime.Nanoseconds()) / 1e6

	if err != nil {
		log.Printf("Error executing task '%s': %v, elapsed %.2f ms", taskName, err, elapsedMs)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	} else {
		log.Printf("Task '%s' executed successfully， %.2f ms", taskName, elapsedMs)

		if result != nil {
			resultType := reflect.TypeOf(result)
			if resultType.Kind() == reflect.Map && resultType.Key().Kind() == reflect.String {
				m := result.(map[string]interface{})
				if _, ok := m["data"]; ok {
					// return m["data"] as the final result
					c.JSON(http.StatusOK, m["data"])
					return
				}
			}
		}
		// Return the script output in the response
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"elapsed_ms": elapsedMs,
			"data":       result,
		})
	}
}

func (h *ScriptManager) GetScript(c *gin.Context) {
	taskName := c.Param("taskname")

	if taskName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task name is required",
		})
		return
	}

	// Get script cache
	scriptCache := h.ScriptPool.Cache

	// Check if task script exists in cache or Redis
	exists, err := scriptCache.ScriptExists(taskName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error checking script existence: %v", err),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Task script '%s' not found", taskName),
		})
		return
	}

	// Read script content from cache
	content, err := scriptCache.GetScript(taskName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read task script: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, ScriptItem{
		Name: taskName,
		Code: content,
	})
}

// SaveTaskScript handles POST /task/:taskname/script endpoint
func (h *ScriptManager) SaveScript(c *gin.Context) {
	taskName := c.Param("taskname")

	if taskName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task name is required",
		})
		return
	}

	// Parse request body
	var script ScriptItem
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Ensure script.Name matches taskName from URL
	if script.Name != taskName {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task name in URL does not match task name in request body",
		})
		return
	}

	// Get script cache
	scriptCache := h.ScriptPool.Cache

	// Store script in cache (which will also store in Redis)
	if err := scriptCache.StoreScript(taskName, script.Code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to save task script: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Task script '%s' saved successfully", taskName),
	})
}

func (h *ScriptManager) DeleteScript(c *gin.Context) {
	taskName := c.Param("taskname")

	if taskName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task name is required",
		})
		return
	}

	// Get script cache
	scriptCache := h.ScriptPool.Cache

	// Check if task script exists in cache or Redis
	exists, err := scriptCache.ScriptExists(taskName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error checking script existence: %v", err),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Task script '%s' not found", taskName),
		})
		return
	}

	// Delete script from cache (which will also delete from Redis)
	if err := scriptCache.DeleteScript(taskName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete task script: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Task script '%s' deleted successfully", taskName),
	})
}

func (h *ScriptManager) ListTaskScripts(c *gin.Context) {
	// Get script cache
	scriptCache := h.ScriptPool.Cache

	// Get all script names from cache
	tasks, err := scriptCache.ListScripts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list task scripts: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func SetupScriptsRoutes(router *gin.Engine, config *cfg.Config, manager *ScriptManager) {

	// Use the configured endpoint or default to "scripts"
	endpoint := config.Script.Endpoint

	// Create a tasks endpoint for listing all tasks
	router.GET("/scripts", manager.ListTaskScripts)

	// Create a task group
	if endpoint != "" {
		taskGroup := router.Group("/" + endpoint)
		{
			taskGroup.GET("/:taskname", manager.Execute)
			taskGroup.POST("/:taskname", manager.ExecutePost)
		}
	}

	scriptsGroup := router.Group("/scripts")
	{
		scriptsGroup.GET("/:taskname", manager.GetScript)
		scriptsGroup.POST("/:taskname", manager.SaveScript)
		scriptsGroup.DELETE("/:taskname", manager.DeleteScript)
	}

	manageGroup := router.Group("/manage")
	{
		manageGroup.GET("/export", manager.ExportScripts)
		manageGroup.POST("/import", manager.ImportScripts)
	}
}
