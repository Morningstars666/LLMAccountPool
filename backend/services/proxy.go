package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"llmaccountpool/models"
	"llmaccountpool/utils"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProxyService struct {
	modelIndex map[uint]int
	mu         sync.RWMutex
	httpClient *http.Client
}

var Proxy *ProxyService

func InitProxy() {
	Proxy = &ProxyService{
		modelIndex: make(map[uint]int),
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
	Proxy.startUsageResetTicker()
}

func (p *ProxyService) startUsageResetTicker() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			p.resetExpiredSources()
		}
	}()
}

func (p *ProxyService) resetExpiredSources() {
	var sources []models.RequestSource
	if err := models.DB.Where("limit_reset_interval > 0 OR limit_reset_time != ''").Find(&sources).Error; err != nil {
		fmt.Printf("Failed to fetch sources for reset: %v\n", err)
		return
	}

	now := time.Now()
	for i := range sources {
		source := &sources[i]
		if source.LimitResetTime != "" {
			if p.shouldResetAtTime(source, now) {
				models.DB.Model(source).Updates(map[string]interface{}{
					"used_count":    0,
					"used_tokens":   0,
					"last_reset_at": now,
				})
				fmt.Printf("Reset usage for source %d (%s) at scheduled time\n", source.ID, source.Name)
			}
			continue
		}

		if source.LimitResetInterval <= 0 {
			continue
		}

		if source.LastResetAt.IsZero() {
			models.DB.Model(source).Updates(map[string]interface{}{
				"last_reset_at": now,
			})
			continue
		}

		elapsed := now.Sub(source.LastResetAt).Seconds()
		if elapsed >= float64(source.LimitResetInterval) {
			models.DB.Model(source).Updates(map[string]interface{}{
				"used_count":    0,
				"used_tokens":   0,
				"last_reset_at": now,
			})
			fmt.Printf("Reset usage for source %d (%s)\n", source.ID, source.Name)
		}
	}
}

func (p *ProxyService) shouldResetAtTime(source *models.RequestSource, now time.Time) bool {
	if source.LimitResetTime == "" {
		return false
	}

	resetTime, err := time.Parse("15:04", source.LimitResetTime)
	if err != nil {
		return false
	}

	currentTime := time.Date(2000, 1, 1, now.Hour(), now.Minute(), 0, 0, time.Local)
	targetTime := time.Date(2000, 1, 1, resetTime.Hour(), resetTime.Minute(), 0, 0, time.Local)

	if source.LastResetAt.IsZero() {
		return currentTime.Equal(targetTime) || currentTime.After(targetTime)
	}

	lastResetDate := source.LastResetAt.Format("2006-01-02")
	today := now.Format("2006-01-02")

	if lastResetDate == today {
		return false
	}

	return currentTime.Equal(targetTime) || currentTime.After(targetTime)
}

type ChatCompletionResult struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	IsStream   bool
	Err        error
}

func (p *ProxyService) HandleChatCompletion(c *gin.Context, apiKey string, reqBody []byte) {
	var chatReq models.ChatCompletionRequest
	if err := json.Unmarshal(reqBody, &chatReq); err != nil {
		utils.RespondWithInvalidRequestError(c, "Invalid request body: "+err.Error(), nil)
		return
	}

	if err := chatReq.Validate(); err != nil {
		if validationErr, ok := err.(*models.ValidationError); ok {
			param := validationErr.Param
			utils.RespondWithInvalidRequestError(c, validationErr.Message, &param)
		} else {
			utils.RespondWithInvalidRequestError(c, err.Error(), nil)
		}
		return
	}

	var apiKeyRecord models.APIKey
	if err := models.DB.Where("key = ?", apiKey).First(&apiKeyRecord).Error; err != nil {
		utils.RespondWithAuthenticationError(c, "Invalid API key provided")
		return
	}

	var externalModel models.ExternalModel
	if apiKeyRecord.ExternalModelID != 0 {
		if err := models.DB.Preload("Sources").First(&externalModel, apiKeyRecord.ExternalModelID).Error; err != nil {
			utils.RespondWithNotFoundError(c, "Model not found", utils.StringPtr("model"))
			return
		}
	} else {
		modelName := chatReq.Model
		if err := models.DB.Preload("Sources").Where("model = ?", modelName).First(&externalModel).Error; err != nil {
			utils.RespondWithNotFoundError(c, "Model not found", utils.StringPtr("model"))
			return
		}
	}

	sources := p.getActiveSources(&externalModel)
	if len(sources) == 0 {
		utils.RespondWithAPIError(c, http.StatusServiceUnavailable, "No available sources for this model")
		return
	}

	source, sourceIndex, err := p.selectSource(&externalModel, sources)
	if err != nil {
		utils.RespondWithAPIError(c, http.StatusServiceUnavailable, err.Error())
		return
	}

	modifiedReq, err := p.modifyRequest(reqBody, source.ModelName)
	if err != nil {
		utils.RespondWithInvalidRequestError(c, "Failed to modify request: "+err.Error(), nil)
		return
	}

	if chatReq.Stream {
		p.handleStreamRequest(c, apiKey, modifiedReq, source, &externalModel, sources, sourceIndex, &apiKeyRecord)
	} else {
		p.handleNonStreamRequest(c, apiKey, modifiedReq, source, &externalModel, sources, sourceIndex, &apiKeyRecord)
	}
}

func (p *ProxyService) handleNonStreamRequest(
	c *gin.Context,
	apiKey string,
	reqBody []byte,
	source models.RequestSource,
	externalModel *models.ExternalModel,
	sources []models.RequestSource,
	sourceIndex int,
	apiKeyRecord *models.APIKey,
) {
	respBody, statusCode, err := p.forwardRequestBytes(source.APIURL, source.APIKey, reqBody)

	if err != nil {
		p.recordUsageInTransaction(models.DB, apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, 0, 0, false)

		result := p.tryOtherSources(c, apiKey, reqBody, externalModel, sources, sourceIndex, apiKeyRecord)
		if result == nil {
			return
		}

		if statusCode >= 400 {
			if errResp, ok := utils.ParseUpstreamError(respBody, statusCode); ok {
				c.JSON(statusCode, errResp)
				return
			}
		}

		utils.RespondWithAPIError(c, http.StatusBadGateway, "All upstream sources failed: "+err.Error())
		return
	}

	var chatResp models.ChatCompletionResponse
	var inputTokens, outputTokens int64

	if err := json.Unmarshal(respBody, &chatResp); err == nil && chatResp.Usage != nil {
		inputTokens = int64(chatResp.Usage.PromptTokens)
		outputTokens = int64(chatResp.Usage.CompletionTokens)
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		return p.recordUsageInTransaction(tx, apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, inputTokens, outputTokens, true)
	})
	if err != nil {
		fmt.Printf("Failed to record usage: %v\n", err)
	}

	if externalModel.Strategy == "round_robin" {
		p.updateRoundRobinIndex(externalModel.ID, len(sources))
	}

	c.Data(statusCode, "application/json", respBody)
}

func (p *ProxyService) handleStreamRequest(
	c *gin.Context,
	apiKey string,
	reqBody []byte,
	source models.RequestSource,
	externalModel *models.ExternalModel,
	sources []models.RequestSource,
	sourceIndex int,
	apiKeyRecord *models.APIKey,
) {
	resp, err := p.forwardRequest(source.APIURL, source.APIKey, reqBody, true)

	if err != nil {
		p.recordUsageInTransaction(models.DB, apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, 0, 0, false)

		result := p.tryOtherSources(c, apiKey, reqBody, externalModel, sources, sourceIndex, apiKeyRecord)
		if result == nil {
			return
		}

		utils.RespondWithAPIError(c, http.StatusBadGateway, "All upstream sources failed: "+err.Error())
		return
	}

	defer resp.Body.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		utils.RespondWithAPIError(c, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	c.Status(http.StatusOK)
	flusher.Flush()

	var totalInputTokens, totalOutputTokens int64
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				c.Writer.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				break
			}

			var streamResp models.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err == nil {
				if streamResp.Usage != nil {
					totalInputTokens = int64(streamResp.Usage.PromptTokens)
					totalOutputTokens = int64(streamResp.Usage.CompletionTokens)
				}
			}

			c.Writer.Write([]byte(line + "\n\n"))
			flusher.Flush()
		}
	}

	if totalInputTokens == 0 && totalOutputTokens == 0 {
		totalOutputTokens = 1
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		return p.recordUsageInTransaction(tx, apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, totalInputTokens, totalOutputTokens, true)
	})
	if err != nil {
		fmt.Printf("Failed to record usage: %v\n", err)
	}

	if externalModel.Strategy == "round_robin" {
		p.updateRoundRobinIndex(externalModel.ID, len(sources))
	}
}

func (p *ProxyService) tryOtherSources(
	c *gin.Context,
	apiKey string,
	reqBody []byte,
	externalModel *models.ExternalModel,
	sources []models.RequestSource,
	failedIndex int,
	apiKeyRecord *models.APIKey,
) error {
	for i := 0; i < len(sources)-1; i++ {
		nextIndex := (failedIndex + 1 + i) % len(sources)
		nextSource := sources[nextIndex]

		if !p.isSourceAvailable(nextSource.ID) {
			continue
		}

		modifiedReq, modErr := p.modifyRequest(reqBody, nextSource.ModelName)
		if modErr != nil {
			continue
		}

		respBody, statusCode, err := p.forwardRequestBytes(nextSource.APIURL, nextSource.APIKey, modifiedReq)
		if err != nil {
			p.recordUsageInTransaction(models.DB, apiKeyRecord.ID, externalModel.ID, nextSource.ID, externalModel.Model, 0, 0, false)
			continue
		}

		var chatResp models.ChatCompletionResponse
		var inputTokens, outputTokens int64

		if err := json.Unmarshal(respBody, &chatResp); err == nil && chatResp.Usage != nil {
			inputTokens = int64(chatResp.Usage.PromptTokens)
			outputTokens = int64(chatResp.Usage.CompletionTokens)
		}

		err = models.DB.Transaction(func(tx *gorm.DB) error {
			return p.recordUsageInTransaction(tx, apiKeyRecord.ID, externalModel.ID, nextSource.ID, externalModel.Model, inputTokens, outputTokens, true)
		})
		if err != nil {
			fmt.Printf("Failed to record usage: %v\n", err)
		}

		if externalModel.Strategy == "round_robin" {
			p.updateRoundRobinIndex(externalModel.ID, len(sources))
		}

		c.Data(statusCode, "application/json", respBody)
		return nil
	}

	return fmt.Errorf("all upstream sources failed")
}

func (p *ProxyService) selectSource(externalModel *models.ExternalModel, sources []models.RequestSource) (models.RequestSource, int, error) {
	var source models.RequestSource
	var sourceIndex int

	if externalModel.Strategy == "round_robin" {
		p.mu.RLock()
		currentIndex := p.modelIndex[externalModel.ID]
		p.mu.RUnlock()
		sourceIndex = currentIndex % len(sources)
		source = sources[sourceIndex]
	} else {
		p.resetRoundRobinIndex(externalModel.ID)
		for i, s := range sources {
			if p.isSourceAvailable(s.ID) {
				source = s
				sourceIndex = i
				break
			}
		}
		if source.ID == 0 {
			return models.RequestSource{}, 0, fmt.Errorf("all sources have exceeded their limits")
		}
	}

	return source, sourceIndex, nil
}

func (p *ProxyService) updateRoundRobinIndex(modelID uint, sourceCount int) {
	p.mu.Lock()
	p.modelIndex[modelID] = (p.modelIndex[modelID] + 1) % sourceCount
	p.mu.Unlock()
}

func (p *ProxyService) resetRoundRobinIndex(modelID uint) {
	p.mu.Lock()
	delete(p.modelIndex, modelID)
	p.mu.Unlock()
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

	p.checkAndResetUsage(&source)

	if source.BillingMode == "count" {
		return source.LimitCount == 0 || source.UsedCount < source.LimitCount
	}
	return source.LimitTokens == 0 || source.UsedTokens < source.LimitTokens
}

func (p *ProxyService) checkAndResetUsage(source *models.RequestSource) {
	now := time.Now()

	if source.LimitResetTime != "" {
		if p.shouldResetAtTime(source, now) {
			models.DB.Model(source).Updates(map[string]interface{}{
				"used_count":    0,
				"used_tokens":   0,
				"last_reset_at": now,
			})
			source.UsedCount = 0
			source.UsedTokens = 0
			source.LastResetAt = now
		}
		return
	}

	if source.LimitResetInterval <= 0 {
		return
	}

	if source.LastResetAt.IsZero() {
		models.DB.Model(source).Updates(map[string]interface{}{
			"last_reset_at": now,
		})
		source.LastResetAt = now
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

func (p *ProxyService) forwardRequest(apiURL, apiKey string, reqBody []byte, isStream bool) (*http.Response, error) {
	fullURL := apiURL
	apiURL = strings.TrimRight(apiURL, "/")
	if !strings.HasSuffix(apiURL, "/v1/chat/completions") && !strings.HasSuffix(apiURL, "/v1") {
		fullURL = apiURL + "/v1/chat/completions"
	} else if strings.HasSuffix(apiURL, "/v1") {
		fullURL = apiURL + "/chat/completions"
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	if isStream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return &http.Response{
			StatusCode: resp.StatusCode,
			Body:       io.NopCloser(bytes.NewBuffer(body)),
			Header:     resp.Header,
		}, fmt.Errorf("upstream error: %d", resp.StatusCode)
	}

	return resp, nil
}

func (p *ProxyService) forwardRequestBytes(apiURL, apiKey string, reqBody []byte) ([]byte, int, error) {
	resp, err := p.forwardRequest(apiURL, apiKey, reqBody, false)
	if err != nil {
		if resp != nil && resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return body, resp.StatusCode, err
		}
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

func (p *ProxyService) updateUsage(sourceID uint, tokens int64) {
	models.DB.Model(&models.RequestSource{}).Where("id = ?", sourceID).Updates(map[string]interface{}{
		"used_count":  gorm.Expr("used_count + ?", 1),
		"used_tokens": gorm.Expr("used_tokens + ?", tokens),
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

func (p *ProxyService) recordUsageInTransaction(tx *gorm.DB, apiKeyID, modelID, sourceID uint, modelName string, inputTokens, outputTokens int64, success bool) error {
	record := models.UsageRecord{
		APIKeyID:        apiKeyID,
		ExternalModelID: modelID,
		SourceID:        sourceID,
		Model:           modelName,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		Success:         success,
	}

	if err := tx.Create(&record).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.RequestSource{}).Where("id = ?", sourceID).Updates(map[string]interface{}{
		"used_count":  gorm.Expr("used_count + ?", 1),
		"used_tokens": gorm.Expr("used_tokens + ?", inputTokens+outputTokens),
	}).Error; err != nil {
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
}

func (p *ProxyService) modifyRequest(reqBody []byte, modelName string) ([]byte, error) {
	if len(reqBody) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	var req map[string]interface{}
	if err := json.Unmarshal(reqBody, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request body: %v", err)
	}

	if req == nil {
		return nil, fmt.Errorf("request body is nil after parsing")
	}

	req["model"] = modelName

	result, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	return result, nil
}

func estimateTokens(messages []map[string]interface{}) int64 {
	total := int64(0)
	for _, msg := range messages {
		if content, ok := msg["content"].(string); ok {
			total += int64(len(content)) / 4
		}
	}
	return total
}
