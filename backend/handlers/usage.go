package handlers

import (
	"llmaccountpool/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UsageResponse struct {
	APIKeyStats []APIKeyUsageStat `json:"api_key_stats"`
	SourceStats []SourceUsageStat `json:"source_stats"`
	ModelStats  []ModelUsageStat  `json:"model_stats"`
}

type APIKeyUsageStat struct {
	ID              uint   `json:"id"`
	Key             string `json:"key"`
	Note            string `json:"note"`
	ExternalModelID uint   `json:"external_model_id"`
	UsedCount       int64  `json:"used_count"`
	UsedTokens      int64  `json:"used_tokens"`
	InputTokens     int64  `json:"input_tokens"`
	OutputTokens    int64  `json:"output_tokens"`
}

type SourceUsageStat struct {
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	ExternalModelID    uint   `json:"external_model_id"`
	ModelName          string `json:"model_name"`
	BillingMode        string `json:"billing_mode"`
	UsedCount          int64  `json:"used_count"`
	UsedTokens         int64  `json:"used_tokens"`
	LimitCount         int64  `json:"limit_count"`
	LimitTokens        int64  `json:"limit_tokens"`
	LimitResetInterval int64  `json:"limit_reset_interval"`
	LastResetAt        string `json:"last_reset_at"`
	IsActive           bool   `json:"is_active"`
}

type ModelUsageStat struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	Strategy    string `json:"strategy"`
	SourceCount int    `json:"source_count"`
}

func GetUsageStats(c *gin.Context) {
	var apiKeys []models.APIKey
	models.DB.Find(&apiKeys)

	var apiKeyStats []APIKeyUsageStat
	for _, key := range apiKeys {
		apiKeyStats = append(apiKeyStats, APIKeyUsageStat{
			ID:              key.ID,
			Key:             key.Key[:12] + "...",
			Note:            key.Note,
			ExternalModelID: key.ExternalModelID,
			UsedCount:       key.UsedCount,
			UsedTokens:      key.UsedTokens,
			InputTokens:     key.InputTokens,
			OutputTokens:    key.OutputTokens,
		})
	}

	var sources []models.RequestSource
	models.DB.Find(&sources)

	var sourceStats []SourceUsageStat
	for _, source := range sources {
		lastResetAt := ""
		if !source.LastResetAt.IsZero() {
			lastResetAt = source.LastResetAt.Format("2006-01-02 15:04:05")
		}
		sourceStats = append(sourceStats, SourceUsageStat{
			ID:                 source.ID,
			Name:               source.Name,
			ExternalModelID:    source.ExternalModelID,
			ModelName:          source.ModelName,
			BillingMode:        source.BillingMode,
			UsedCount:          source.UsedCount,
			UsedTokens:         source.UsedTokens,
			LimitCount:         source.LimitCount,
			LimitTokens:        source.LimitTokens,
			LimitResetInterval: source.LimitResetInterval,
			LastResetAt:        lastResetAt,
			IsActive:           source.IsActive,
		})
	}

	var externalModels []models.ExternalModel
	models.DB.Preload("Sources").Find(&externalModels)

	var modelStats []ModelUsageStat
	for _, m := range externalModels {
		modelStats = append(modelStats, ModelUsageStat{
			ID:          m.ID,
			Name:        m.Name,
			Model:       m.Model,
			Strategy:    m.Strategy,
			SourceCount: len(m.Sources),
		})
	}

	response := UsageResponse{
		APIKeyStats: apiKeyStats,
		SourceStats: sourceStats,
		ModelStats:  modelStats,
	}

	c.JSON(http.StatusOK, response)
}

func GetUsageRecords(c *gin.Context) {
	keyID := c.Query("api_key_id")
	modelID := c.Query("model_id")

	var records []models.UsageRecord
	query := models.DB
	if keyID != "" {
		query = query.Where("api_key_id = ?", keyID)
	}
	if modelID != "" {
		query = query.Where("external_model_id = ?", modelID)
	}
	query.Order("created_at desc").Limit(100).Find(&records)

	c.JSON(http.StatusOK, records)
}
