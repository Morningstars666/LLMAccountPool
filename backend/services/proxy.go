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

// ProxyService 代理服务
type ProxyService struct {
	modelIndex map[uint]int
	mu         sync.RWMutex
	httpClient *http.Client
}

// Proxy 全局代理实例
var Proxy *ProxyService

// InitProxy 初始化代理服务
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
}

// ChatCompletionResult 聊天完成结果
type ChatCompletionResult struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	IsStream   bool
	Err        error
}

// HandleChatCompletion 处理聊天完成请求（支持流式和非流式）
func (p *ProxyService) HandleChatCompletion(c *gin.Context, apiKey string, reqBody []byte) {
	// 解析请求
	var chatReq models.ChatCompletionRequest
	if err := json.Unmarshal(reqBody, &chatReq); err != nil {
		utils.RespondWithInvalidRequestError(c, "Invalid request body: "+err.Error(), nil)
		return
	}

	// 验证请求参数
	if err := chatReq.Validate(); err != nil {
		validationErr := err.(*models.ValidationError)
		param := validationErr.Param
		utils.RespondWithInvalidRequestError(c, validationErr.Message, &param)
		return
	}

	// 验证 API Key
	var apiKeyRecord models.APIKey
	if err := models.DB.Where("key = ?", apiKey).First(&apiKeyRecord).Error; err != nil {
		utils.RespondWithAuthenticationError(c, "Invalid API key provided")
		return
	}

	// 获取关联的模型配置
	var externalModel models.ExternalModel
	if err := models.DB.Preload("Sources").First(&externalModel, apiKeyRecord.ExternalModelID).Error; err != nil {
		utils.RespondWithNotFoundError(c, "Model not found", utils.StringPtr("model"))
		return
	}

	// 获取活跃的数据源
	sources := p.getActiveSources(&externalModel)
	if len(sources) == 0 {
		utils.RespondWithAPIError(c, http.StatusServiceUnavailable, "No available sources for this model")
		return
	}

	// 选择数据源
	source, sourceIndex, err := p.selectSource(&externalModel, sources)
	if err != nil {
		utils.RespondWithAPIError(c, http.StatusServiceUnavailable, err.Error())
		return
	}

	// 修改请求体（替换模型名称）
	modifiedReq, err := p.modifyRequest(reqBody, source.ModelName)
	if err != nil {
		utils.RespondWithInvalidRequestError(c, "Failed to modify request: "+err.Error(), nil)
		return
	}

	// 转发请求
	if chatReq.Stream {
		p.handleStreamRequest(c, apiKey, modifiedReq, source, &externalModel, sources, sourceIndex, &apiKeyRecord)
	} else {
		p.handleNonStreamRequest(c, apiKey, modifiedReq, source, &externalModel, sources, sourceIndex, &apiKeyRecord)
	}
}

// handleNonStreamRequest 处理非流式请求
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
		// 记录失败
		p.recordUsage(apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, 0, 0, false)

		// 尝试其他数据源
		result := p.tryOtherSources(c, apiKey, reqBody, externalModel, sources, sourceIndex, apiKeyRecord)
		if result != nil {
			return
		}

		// 所有数据源都失败
		if statusCode >= 400 {
			// 尝试解析上游错误
			if errResp, ok := utils.ParseUpstreamError(respBody, statusCode); ok {
				c.JSON(statusCode, errResp)
				return
			}
		}

		utils.RespondWithAPIError(c, http.StatusBadGateway, "All upstream sources failed: "+err.Error())
		return
	}

	// 解析响应以获取使用统计
	var chatResp models.ChatCompletionResponse
	var inputTokens, outputTokens int64

	if err := json.Unmarshal(respBody, &chatResp); err == nil && chatResp.Usage != nil {
		inputTokens = int64(chatResp.Usage.PromptTokens)
		outputTokens = int64(chatResp.Usage.CompletionTokens)
	}

	// 更新使用量
	p.updateUsage(source.ID, inputTokens+outputTokens)
	p.recordUsage(apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, inputTokens, outputTokens, true)

	// 更新轮询索引
	if externalModel.Strategy == "round_robin" {
		p.updateRoundRobinIndex(externalModel.ID, len(sources))
	}

	// 返回响应
	c.Data(statusCode, "application/json", respBody)
}

// handleStreamRequest 处理流式请求
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
		// 记录失败
		p.recordUsage(apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, 0, 0, false)

		// 尝试其他数据源（非流式降级）
		result := p.tryOtherSources(c, apiKey, reqBody, externalModel, sources, sourceIndex, apiKeyRecord)
		if result != nil {
			return
		}

		utils.RespondWithAPIError(c, http.StatusBadGateway, "All upstream sources failed: "+err.Error())
		return
	}

	defer resp.Body.Close()

	// 更新轮询索引
	if externalModel.Strategy == "round_robin" {
		p.updateRoundRobinIndex(externalModel.ID, len(sources))
	}

	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 流式传输数据
	var totalInputTokens, totalOutputTokens int64
	scanner := bufio.NewScanner(resp.Body)

	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			return false
		}

		line := scanner.Text()

		// 跳过空行
		if line == "" {
			return true
		}

		// 处理 SSE 格式
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否结束
			if data == "[DONE]" {
				w.Write([]byte("data: [DONE]\n\n"))
				return false
			}

			// 解析流式响应以获取使用统计
			var streamResp models.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err == nil {
				if streamResp.Usage != nil {
					totalInputTokens = int64(streamResp.Usage.PromptTokens)
					totalOutputTokens = int64(streamResp.Usage.CompletionTokens)
				}
			}

			// 转发数据
			w.Write([]byte(line + "\n\n"))
		}

		return true
	})

	// 记录使用量（流式请求结束后）
	if totalInputTokens == 0 && totalOutputTokens == 0 {
		// 如果没有获取到使用量，进行估算
		totalOutputTokens = 1 // 流式请求至少消耗1个token
	}

	p.updateUsage(source.ID, totalInputTokens+totalOutputTokens)
	p.recordUsage(apiKeyRecord.ID, externalModel.ID, source.ID, externalModel.Model, totalInputTokens, totalOutputTokens, true)
}

// tryOtherSources 尝试其他数据源
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

		// 修改请求体
		modifiedReq, modErr := p.modifyRequest(reqBody, nextSource.ModelName)
		if modErr != nil {
			continue
		}

		// 尝试转发请求
		respBody, statusCode, err := p.forwardRequestBytes(nextSource.APIURL, nextSource.APIKey, modifiedReq)
		if err != nil {
			p.recordUsage(apiKeyRecord.ID, externalModel.ID, nextSource.ID, externalModel.Model, 0, 0, false)
			continue
		}

		// 解析响应
		var chatResp models.ChatCompletionResponse
		var inputTokens, outputTokens int64

		if err := json.Unmarshal(respBody, &chatResp); err == nil && chatResp.Usage != nil {
			inputTokens = int64(chatResp.Usage.PromptTokens)
			outputTokens = int64(chatResp.Usage.CompletionTokens)
		}

		// 更新使用量
		p.updateUsage(nextSource.ID, inputTokens+outputTokens)
		p.recordUsage(apiKeyRecord.ID, externalModel.ID, nextSource.ID, externalModel.Model, inputTokens, outputTokens, true)

		// 返回响应
		c.Data(statusCode, "application/json", respBody)
		return nil
	}

	return fmt.Errorf("all upstream sources failed")
}

// selectSource 选择数据源
func (p *ProxyService) selectSource(externalModel *models.ExternalModel, sources []models.RequestSource) (models.RequestSource, int, error) {
	p.mu.RLock()
	currentIndex := p.modelIndex[externalModel.ID]
	p.mu.RUnlock()

	var source models.RequestSource
	var sourceIndex int

	if externalModel.Strategy == "round_robin" {
		sourceIndex = currentIndex % len(sources)
		source = sources[sourceIndex]
	} else {
		// 优先策略：找到第一个可用的数据源
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

// updateRoundRobinIndex 更新轮询索引
func (p *ProxyService) updateRoundRobinIndex(modelID uint, sourceCount int) {
	p.mu.Lock()
	p.modelIndex[modelID] = (p.modelIndex[modelID] + 1) % sourceCount
	p.mu.Unlock()
}

// getActiveSources 获取活跃的数据源
func (p *ProxyService) getActiveSources(model *models.ExternalModel) []models.RequestSource {
	var sources []models.RequestSource
	models.DB.Where("external_model_id = ? AND is_active = ?", model.ID, true).Find(&sources)
	return sources
}

// isSourceAvailable 检查数据源是否可用
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

// checkAndResetUsage 检查并重置使用量
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

// forwardRequest 转发请求到上游（非流式）
func (p *ProxyService) forwardRequest(apiURL, apiKey string, reqBody []byte, isStream bool) (*http.Response, error) {
	fullURL := apiURL
	if !strings.HasSuffix(apiURL, "/v1/chat/completions") {
		fullURL = strings.TrimRight(apiURL, "/") + "/v1/chat/completions"
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

	if resp.StatusCode != http.StatusOK {
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

// forwardRequestBytes 转发请求并返回字节（非流式）
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

// updateUsage 更新数据源使用量
func (p *ProxyService) updateUsage(sourceID uint, tokens int64) {
	models.DB.Model(&models.RequestSource{}).Where("id = ?", sourceID).Updates(map[string]interface{}{
		"used_count":  gorm.Expr("used_count + ?", 1),
		"used_tokens": gorm.Expr("used_tokens + ?", tokens),
	})
}

// recordUsage 记录使用记录
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

// modifyRequest 修改请求体（替换模型名称）
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

// estimateTokens 估算令牌数
func estimateTokens(messages []map[string]interface{}) int64 {
	total := int64(0)
	for _, msg := range messages {
		if content, ok := msg["content"].(string); ok {
			total += int64(len(content)) / 4
		}
	}
	return total
}
