package main

import (
	"io"
	"llmaccountpool/config"
	"llmaccountpool/handlers"
	"llmaccountpool/middleware"
	"llmaccountpool/models"
	"llmaccountpool/services"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	models.InitDB(cfg)
	services.InitProxy()

	r := gin.Default()

	r.Use(middleware.CORSMiddleware(cfg))

	r.LoadHTMLGlob("../frontend/*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
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
