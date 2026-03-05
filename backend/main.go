package main

import (
	"context"
	"llmaccountpool/config"
	"llmaccountpool/handlers"
	"llmaccountpool/middleware"
	"llmaccountpool/models"
	"llmaccountpool/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	models.InitDB(cfg)
	services.InitProxy()

	r := gin.Default()

	r.Use(middleware.CORSMiddleware(cfg))

	r.Static("/assets", "../frontend/dist/assets")
	r.StaticFile("/vite.svg", "../frontend/dist/vite.svg")

	r.GET("/", func(c *gin.Context) {
		c.File("../frontend/dist/index.html")
	})

	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path[:5] != "/api/" && c.Request.URL.Path[:3] != "/v1" {
			c.File("../frontend/dist/index.html")
		}
	})

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
		admin.POST("/models/import", handlers.ImportModelsFromExcel)
		admin.GET("/models/template", handlers.DownloadTemplate)

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

	r.POST("/v1/chat/completions", handlers.HandleChatCompletions)

	port := cfg.ServerPort

	if port == "" {
		port = "8080"
	}

	if _, err := os.Stat("../data"); os.IsNotExist(err) {
		os.Mkdir("../data", 0755)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
