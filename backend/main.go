package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"llmaccountpool/config"
	"llmaccountpool/handlers"
	"llmaccountpool/middleware"
	"llmaccountpool/models"
	"llmaccountpool/services"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

var fileHashes map[string]string
var hashOnce sync.Once

func getFileHash(filePath string) string {
	hashOnce.Do(func() {
		fileHashes = make(map[string]string)
	})

	if hash, ok := fileHashes[filePath]; ok {
		return hash
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])[:8]
	fileHashes[filePath] = hashStr
	return hashStr
}

func getStaticFileVersion() map[string]string {
	wd, _ := os.Getwd()
	cssPath := filepath.Join(wd, "../frontend/css/style.css")
	jsPath := filepath.Join(wd, "../frontend/js/app.js")

	return map[string]string{
		"css": getFileHash(cssPath),
		"js":  getFileHash(jsPath),
	}
}

func main() {
	cfg := config.LoadConfig()

	models.InitDB(cfg)
	services.InitProxy()

	r := gin.Default()

	r.Use(middleware.CORSMiddleware(cfg))

	r.LoadHTMLGlob("../frontend/*.html")

	r.GET("/", func(c *gin.Context) {
		versions := getStaticFileVersion()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"cssVersion": versions["css"],
			"jsVersion":  versions["js"],
		})
	})

	r.GET("/api/static-versions", func(c *gin.Context) {
		c.JSON(http.StatusOK, getStaticFileVersion())
	})

	r.Static("/static/css", "../frontend/css")
	r.Static("/static/js", "../frontend/js")

	r.POST("/api/login", func(c *gin.Context) {
		handlers.Login(c, cfg)
	})

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(cfg))
	{
		admin.GET("/profile", handlers.GetProfile)
		admin.POST("/refresh-token", func(c *gin.Context) {
			handlers.RefreshToken(c, cfg)
		})
		admin.POST("/change-password", handlers.ChangePassword)
		admin.POST("/change-username", handlers.ChangeUsername)

		admin.GET("/models", handlers.GetExternalModels)
		admin.GET("/models/:id", handlers.GetExternalModel)
		admin.POST("/models", handlers.CreateExternalModel)
		admin.PUT("/models/:id", handlers.UpdateExternalModel)
		admin.DELETE("/models/:id", handlers.DeleteExternalModel)

		admin.GET("/sources", handlers.GetRequestSources)
		admin.GET("/sources/:id", handlers.GetRequestSource)
		admin.POST("/sources", handlers.CreateRequestSource)
		admin.PUT("/sources/:id", handlers.UpdateRequestSource)
		admin.DELETE("/sources/:id", handlers.DeleteRequestSource)
		admin.POST("/sources/:id/reset", handlers.ResetSourceUsage)
		admin.PATCH("/sources/:id/name", handlers.UpdateRequestSourceName)

		admin.GET("/keys", handlers.GetAPIKeys)
		admin.POST("/keys", handlers.CreateAPIKey)
		admin.DELETE("/keys/:id", handlers.DeleteAPIKey)
		admin.POST("/keys/:id/reset", handlers.ResetAPIKeyUsage)

		admin.GET("/usage", handlers.GetUsageStats)
		admin.GET("/usage/records", handlers.GetUsageRecords)
		admin.GET("/server-info", func(c *gin.Context) {
			handlers.GetServerInfo(c, cfg)
		})
	}

	r.POST("/v1/chat/completions", func(c *gin.Context) {
		apiKey := c.GetHeader("Authorization")
		if apiKey == "" {
			apiKey = c.Query("key")
		}
		apiKey = apiKey[len("Bearer "):]

		body, _ := io.ReadAll(c.Request.Body)

		statusCode, respBody, err := services.Proxy.HandleChatCompletion(apiKey, body)
		if err != nil {
			c.JSON(statusCode, gin.H{"error": err.Error()})
			return
		}

		c.Data(statusCode, "application/json", respBody)
	})

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	if _, err := os.Stat("../data"); os.IsNotExist(err) {
		os.Mkdir("../data", 0755)
	}

	r.Run(":" + port)
}
