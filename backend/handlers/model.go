package handlers

import (
	"llmaccountpool/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExternalModelRequest struct {
	Name     string `json:"name" binding:"required"`
	Model    string `json:"model" binding:"required"`
	Strategy string `json:"strategy"`
}

func GetExternalModels(c *gin.Context) {
	var externalModels []models.ExternalModel
	if err := models.DB.Preload("Sources").Find(&externalModels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models"})
		return
	}
	c.JSON(http.StatusOK, externalModels)
}

func GetExternalModel(c *gin.Context) {
	id := c.Param("id")
	var model models.ExternalModel
	if err := models.DB.Preload("Sources").First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	c.JSON(http.StatusOK, model)
}

func CreateExternalModel(c *gin.Context) {
	var req ExternalModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = "round_robin"
	}

	model := models.ExternalModel{
		Name:     req.Name,
		Model:    req.Model,
		Strategy: strategy,
	}

	if err := models.DB.Create(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model"})
		return
	}

	c.JSON(http.StatusCreated, model)
}

func UpdateExternalModel(c *gin.Context) {
	id := c.Param("id")
	var req ExternalModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var model models.ExternalModel
	if err := models.DB.First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	model.Name = req.Name
	model.Model = req.Model
	if req.Strategy != "" {
		model.Strategy = req.Strategy
	}

	models.DB.Save(&model)
	c.JSON(http.StatusOK, model)
}

func DeleteExternalModel(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Delete(&models.ExternalModel{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}
