package handlers

import (
	"llmaccountpool/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestSourceRequest struct {
	ExternalModelID    uint   `json:"external_model_id" binding:"required"`
	Name               string `json:"name" binding:"required"`
	APIURL             string `json:"api_url" binding:"required"`
	APIKey             string `json:"api_key" binding:"required"`
	ModelName          string `json:"model_name" binding:"required"`
	BillingMode        string `json:"billing_mode"`
	LimitCount         int64  `json:"limit_count"`
	LimitTokens        int64  `json:"limit_tokens"`
	LimitResetInterval int64  `json:"limit_reset_interval"`
	LimitResetTime     string `json:"limit_reset_time"`
}

type UpdateSourceNameRequest struct {
	Name string `json:"name" binding:"required"`
}

func GetRequestSources(c *gin.Context) {
	modelID := c.Query("external_model_id")
	var sources []models.RequestSource
	query := models.DB
	if modelID != "" {
		query = query.Where("external_model_id = ?", modelID)
	}
	if err := query.Find(&sources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sources"})
		return
	}
	c.JSON(http.StatusOK, sources)
}

func GetRequestSource(c *gin.Context) {
	id := c.Param("id")
	var source models.RequestSource
	if err := models.DB.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}
	c.JSON(http.StatusOK, source)
}

func CreateRequestSource(c *gin.Context) {
	var req RequestSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	billingMode := req.BillingMode
	if billingMode == "" {
		billingMode = "count"
	}

	source := models.RequestSource{
		ExternalModelID:    req.ExternalModelID,
		Name:               req.Name,
		APIURL:             req.APIURL,
		APIKey:             req.APIKey,
		ModelName:          req.ModelName,
		BillingMode:        billingMode,
		LimitCount:         req.LimitCount,
		LimitTokens:        req.LimitTokens,
		LimitResetInterval: req.LimitResetInterval,
		LimitResetTime:     req.LimitResetTime,
		IsActive:           true,
	}

	if err := models.DB.Create(&source).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create source"})
		return
	}

	c.JSON(http.StatusCreated, source)
}

func UpdateRequestSource(c *gin.Context) {
	id := c.Param("id")
	var req RequestSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var source models.RequestSource
	if err := models.DB.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	source.Name = req.Name
	source.APIURL = req.APIURL
	source.APIKey = req.APIKey
	source.ModelName = req.ModelName
	if req.BillingMode != "" {
		source.BillingMode = req.BillingMode
	}
	source.LimitCount = req.LimitCount
	source.LimitTokens = req.LimitTokens
	source.LimitResetInterval = req.LimitResetInterval
	source.LimitResetTime = req.LimitResetTime

	models.DB.Save(&source)
	c.JSON(http.StatusOK, source)
}

func DeleteRequestSource(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Delete(&models.RequestSource{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete source"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Source deleted successfully"})
}

func ResetSourceUsage(c *gin.Context) {
	id := c.Param("id")
	var source models.RequestSource
	if err := models.DB.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	now := time.Now()
	models.DB.Model(&models.RequestSource{}).Where("id = ?", id).Updates(map[string]interface{}{
		"used_count":    0,
		"used_tokens":   0,
		"last_reset_at": now,
	})

	models.DB.First(&source, id)
	c.JSON(http.StatusOK, gin.H{"message": "Usage reset successfully", "source": source})
}

func UpdateRequestSourceName(c *gin.Context) {
	id := c.Param("id")
	var req UpdateSourceNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var source models.RequestSource
	if err := models.DB.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	source.Name = req.Name
	models.DB.Save(&source)
	c.JSON(http.StatusOK, gin.H{"message": "Name updated successfully", "source": source})
}
