package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"main/config"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ExportTaskScripts handles GET /export/scripts endpoint
// It exports all task scripts as a zip file
func (h *ScriptManager) ExportScripts(c *gin.Context) {
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

	// Create a buffer to write our zip to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add files to the zip
	for _, taskName := range tasks {
		// Get script content
		content, err := scriptCache.GetScript(taskName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to read task script '%s': %v", taskName, err),
			})
			return
		}

		// Create a file in the zip
		fileName := fmt.Sprintf("%s.js", taskName)
		f, err := zipWriter.Create(fileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to create zip file entry: %v", err),
			})
			return
		}

		// Write content to the file
		_, err = f.Write([]byte(content))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to write to zip file: %v", err),
			})
			return
		}
	}

	// Close the zip writer
	err = zipWriter.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to finalize zip file: %v", err),
		})
		return
	}

	// Set headers for file download
	timestamp := time.Now().Format("20060102")
	filename := fmt.Sprintf("Scripts-%s_%s.zip", config.CONFIG.Script.Endpoint, timestamp)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))

	// Write the zip file to the response
	c.Writer.Write(buf.Bytes())
}

// ImportTaskScripts handles POST /import/scripts endpoint
// It imports task scripts from an uploaded zip file
func (h *ScriptManager) ImportScripts(c *gin.Context) {
	// Get script cache
	scriptCache := h.ScriptPool.Cache

	// Get the uploaded file
	file, err := c.FormFile("zipfile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to get uploaded file: %v", err),
		})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to open uploaded file: %v", err),
		})
		return
	}
	defer src.Close()

	// Create a buffer to read the zip file
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read uploaded file: %v", err),
		})
		return
	}

	// Create a zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid zip file: %v", err),
		})
		return
	}

	// Process each file in the zip
	importedCount := 0
	skippedCount := 0
	for _, zipFile := range zipReader.File {
		// Skip directories
		if zipFile.FileInfo().IsDir() {
			continue
		}

		// Get the file name without extension
		fileName := filepath.Base(zipFile.Name)
		ext := filepath.Ext(fileName)
		taskName := strings.TrimSuffix(fileName, ext)

		// Skip non-js files
		if ext != ".js" {
			skippedCount++
			continue
		}

		// Open the file in the zip
		f, err := zipFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to open file in zip: %v", err),
			})
			return
		}
		defer f.Close()

		// Read the file content
		content := new(bytes.Buffer)
		_, err = io.Copy(content, f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to read file in zip: %v", err),
			})
			return
		}

		// Store the script in cache
		err = scriptCache.StoreScript(taskName, content.String())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to store script '%s': %v", taskName, err),
			})
			return
		}

		importedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"message":        fmt.Sprintf("Successfully imported %d scripts, skipped %d non-JS files", importedCount, skippedCount),
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
	})
}
