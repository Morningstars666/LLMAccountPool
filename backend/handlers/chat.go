package handlers

import (
	"io"
	"llmaccountpool/models"
	"llmaccountpool/services"
	"llmaccountpool/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleChatCompletions 处理 /v1/chat/completions 端点请求
// 支持与 OpenAI API 完全兼容的流式和非流式调用
func HandleChatCompletions(c *gin.Context) {
	// 提取 API Key
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		utils.RespondWithAuthenticationError(c, "You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth (i.e. Authorization: Bearer YOUR_KEY)")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.RespondWithInvalidRequestError(c, "Failed to read request body", nil)
		return
	}

	// 检查请求体是否为空
	if len(body) == 0 {
		utils.RespondWithInvalidRequestError(c, "Request body cannot be empty", nil)
		return
	}

	// 处理聊天完成请求
	services.Proxy.HandleChatCompletion(c, apiKey, body)
}

// extractAPIKey 从请求中提取 API Key
// 支持 Authorization 头部和 URL 查询参数
func extractAPIKey(c *gin.Context) string {
	// 首先检查 Authorization 头部
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 支持 Bearer token 格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 也支持直接传递 token（某些客户端可能不使用 Bearer 前缀）
		return strings.TrimSpace(authHeader)
	}

	// 然后检查 URL 查询参数（用于测试或特定场景）
	if key := c.Query("api_key"); key != "" {
		return key
	}

	// 最后检查 x-api-key 头部（某些代理配置使用）
	if key := c.GetHeader("x-api-key"); key != "" {
		return key
	}

	return ""
}

// HandleModels 处理 /v1/models 端点请求
// 返回可用的模型列表（OpenAI 兼容格式）
func HandleModels(c *gin.Context) {
	var externalModels []models.ExternalModel
	if err := models.DB.Find(&externalModels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models"})
		return
	}

	modelEntries := make([]models.ModelEntry, len(externalModels))
	for i, m := range externalModels {
		modelEntries[i] = models.ModelEntry{
			ID:      m.Model,
			Object:  "model",
			Created: m.CreatedAt.Unix(),
			OwnedBy: "system",
		}
	}

	c.JSON(http.StatusOK, models.ModelListResponse{
		Object: "list",
		Data:   modelEntries,
	})
}
