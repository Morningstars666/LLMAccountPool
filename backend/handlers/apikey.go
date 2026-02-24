package handlers

import (
	"llmaccountpool/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIKeyRequest struct {
	ExternalModelID uint   `json:"external_model_id" binding:"required"`
	Note            string `json:"note"`
}

func GetAPIKeys(c *gin.Context) {
	var keys []models.APIKey
	if err := models.DB.Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func CreateAPIKey(c *gin.Context) {
	var req APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	key := models.APIKey{
		Key:             "sk-" + uuid.New().String(),
		Note:            req.Note,
		ExternalModelID: req.ExternalModelID,
	}

	if err := models.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, key)
}

func DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Delete(&models.APIKey{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

func ResetAPIKeyUsage(c *gin.Context) {
	id := c.Param("id")
	var apiKey models.APIKey
	if err := models.DB.First(&apiKey, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	models.DB.Model(&models.APIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"used_count":    0,
		"used_tokens":   0,
		"input_tokens":  0,
		"output_tokens": 0,
	})

	models.DB.First(&apiKey, id)
	c.JSON(http.StatusOK, gin.H{"message": "API key usage reset successfully", "api_key": apiKey})
}
