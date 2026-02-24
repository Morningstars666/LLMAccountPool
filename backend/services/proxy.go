package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"llmaccountpool/models"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"
)

type ProxyService struct {
	modelIndex map[uint]int
	mu         sync.RWMutex
}

var Proxy *ProxyService

func InitProxy() {
	Proxy = &ProxyService{
		modelIndex: make(map[uint]int),
	}
}

type ChatCompletionRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Stream      bool                     `json:"stream,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      interface{} `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (p *ProxyService) HandleChatCompletion(apiKey string, reqBody []byte) (int, []byte, error) {
	var chatReq ChatCompletionRequest
	if err := json.Unmarshal(reqBody, &chatReq); err != nil {
		return http.StatusBadRequest, nil, fmt.Errorf("invalid request body: %v", err)
	}

	var apiKeyRecord models.APIKey
	if err := models.DB.Where("key = ?", apiKey).First(&apiKeyRecord).Error; err != nil {
		return http.StatusUnauthorized, nil, fmt.Errorf("invalid API key")
	}

	var model models.ExternalModel
	if err := models.DB.Preload("Sources").First(&model, apiKeyRecord.ExternalModelID).Error; err != nil {
		return http.StatusNotFound, nil, fmt.Errorf("model not found")
	}

	sources := p.getActiveSources(&model)
	if len(sources) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("no available sources for this model")
	}

	var source models.RequestSource
	var sourceIndex int

	p.mu.RLock()
	currentIndex := p.modelIndex[model.ID]
	p.mu.RUnlock()

	if model.Strategy == "round_robin" {
		sourceIndex = currentIndex % len(sources)
		source = sources[sourceIndex]
	} else {
		for i, s := range sources {
			if p.isSourceAvailable(s.ID) {
				source = s
				sourceIndex = i
				break
			}
		}
		if source.ID == 0 {
			return http.StatusBadRequest, nil, fmt.Errorf("all sources have exceeded their limits")
		}
	}

	modifiedReq := modifyRequest(reqBody, source.ModelName)

	resp, err := p.forwardRequest(source.APIURL, source.APIKey, modifiedReq)
	if err != nil {
		p.recordUsage(apiKeyRecord.ID, model.ID, source.ID, model.Model, 0, 0, false)

		if model.Strategy == "round_robin" {
			p.mu.Lock()
			p.modelIndex[model.ID] = (currentIndex + 1) % len(sources)
			p.mu.Unlock()
		}

		for i := 0; i < len(sources); i++ {
			nextIndex := (sourceIndex + 1 + i) % len(sources)
			nextSource := sources[nextIndex]
			if p.isSourceAvailable(nextSource.ID) {
				modifiedReq = modifyRequest(reqBody, nextSource.ModelName)
				resp, err = p.forwardRequest(nextSource.APIURL, nextSource.APIKey, modifiedReq)
				if err == nil {
					p.updateUsage(nextSource.ID, &chatReq, resp)
					p.recordUsage(apiKeyRecord.ID, model.ID, nextSource.ID, model.Model, 0, 0, true)
					return http.StatusOK, resp, nil
				}
			}
		}
		return http.StatusBadGateway, nil, fmt.Errorf("all upstream sources failed")
	}

	p.updateUsage(source.ID, &chatReq, resp)
	p.recordUsage(apiKeyRecord.ID, model.ID, source.ID, model.Model, 0, 0, true)

	if model.Strategy == "round_robin" {
		p.mu.Lock()
		p.modelIndex[model.ID] = (currentIndex + 1) % len(sources)
		p.mu.Unlock()
	}

	return http.StatusOK, resp, nil
}

func (p *ProxyService) getActiveSources(model *models.ExternalModel) []models.RequestSource {
	var sources []models.RequestSource
	models.DB.Where("external_model_id = ? AND is_active = ?", model.ID, true).Find(&sources)
	return sources
}

func (p *ProxyService) isSourceAvailable(sourceID uint) bool {
	var source models.RequestSource
	if err := models.DB.First(&source, sourceID).Error; err != nil {
		return false
	}

	p.checkAndResetUsageByID(sourceID)

	if source.BillingMode == "count" {
		return source.LimitCount == 0 || source.UsedCount < source.LimitCount
	}
	return source.LimitTokens == 0 || source.UsedTokens < source.LimitTokens
}

func (p *ProxyService) checkAndResetUsageByID(sourceID uint) {
	var source models.RequestSource
	if err := models.DB.First(&source, sourceID).Error; err != nil {
		return
	}

	if source.LimitResetInterval <= 0 {
		return
	}

	now := time.Now()
	if source.LastResetAt.IsZero() {
		return
	}

	elapsed := now.Sub(source.LastResetAt).Seconds()
	if elapsed >= float64(source.LimitResetInterval) {
		models.DB.Model(&source).Updates(map[string]interface{}{
			"used_count":    0,
			"used_tokens":   0,
			"last_reset_at": now,
		})
	}
}

func (p *ProxyService) checkAndResetUsage(source *models.RequestSource) {
	if source.LimitResetInterval <= 0 {
		return
	}

	now := time.Now()
	if source.LastResetAt.IsZero() {
		return
	}

	elapsed := now.Sub(source.LastResetAt).Seconds()
	if elapsed >= float64(source.LimitResetInterval) {
		models.DB.Model(source).Updates(map[string]interface{}{
			"used_count":    0,
			"used_tokens":   0,
			"last_reset_at": now,
		})
		source.UsedCount = 0
		source.UsedTokens = 0
		source.LastResetAt = now
	}
}

func (p *ProxyService) forwardRequest(apiURL, apiKey string, reqBody []byte) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("POST", apiURL+"/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream error: %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (p *ProxyService) updateUsage(sourceID uint, req *ChatCompletionRequest, respBody []byte) {
	var resp ChatCompletionResponse
	json.Unmarshal(respBody, &resp)

	totalTokens := resp.Usage.TotalTokens

	models.DB.Model(&models.RequestSource{}).Where("id = ?", sourceID).Updates(map[string]interface{}{
		"used_count":  gorm.Expr("used_count + ?", 1),
		"used_tokens": gorm.Expr("used_tokens + ?", totalTokens),
	})
}

func (p *ProxyService) recordUsage(apiKeyID, modelID, sourceID uint, modelName string, inputTokens, outputTokens int64, success bool) {
	record := models.UsageRecord{
		APIKeyID:        apiKeyID,
		ExternalModelID: modelID,
		SourceID:        sourceID,
		Model:           modelName,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		Success:         success,
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.APIKey{}).Where("id = ?", apiKeyID).Updates(map[string]interface{}{
			"used_count":    gorm.Expr("used_count + 1"),
			"used_tokens":   gorm.Expr("used_tokens + ?", inputTokens+outputTokens),
			"input_tokens":  gorm.Expr("input_tokens + ?", inputTokens),
			"output_tokens": gorm.Expr("output_tokens + ?", outputTokens),
		}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Failed to record usage: %v\n", err)
	}
}

func modifyRequest(reqBody []byte, modelName string) []byte {
	var req map[string]interface{}
	json.Unmarshal(reqBody, &req)
	req["model"] = modelName
	result, _ := json.Marshal(req)
	return result
}

func estimateTokens(messages []map[string]interface{}) int64 {
	total := 0
	for _, msg := range messages {
		if content, ok := msg["content"].(string); ok {
			total += len(content) / 4
		}
	}
	return int64(total)
}
