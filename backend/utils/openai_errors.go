package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAIErrorType 错误类型常量
const (
	OpenAIErrorTypeInvalidRequest = "invalid_request_error"
	OpenAIErrorTypeAuthentication = "authentication_error"
	OpenAIErrorTypePermission     = "permission_error"
	OpenAIErrorTypeNotFound       = "not_found_error"
	OpenAIErrorTypeRateLimit      = "rate_limit_error"
	OpenAIErrorTypeAPIError       = "api_error"
	OpenAIErrorTypeOverloaded     = "overloaded_error"
)

// OpenAIErrorCode 错误代码常量
const (
	OpenAIErrorCodeContextLengthExceeded = "context_length_exceeded"
	OpenAIErrorCodeInsufficientQuota     = "insufficient_quota"
	OpenAIErrorCodeInvalidAPIKey         = "invalid_api_key"
	OpenAIErrorCodeModelNotFound         = "model_not_found"
	OpenAIErrorCodeMaxTokensTooLarge     = "max_tokens_too_large"
	OpenAIErrorCodeInvalidTemperature    = "invalid_temperature"
	OpenAIErrorCodeInvalidTopP           = "invalid_top_p"
	OpenAIErrorCodeInvalidN              = "invalid_n"
	OpenAIErrorCodeInvalidStop           = "invalid_stop"
	OpenAIErrorCodeTimeout               = "timeout"
	OpenAIErrorCodeRateLimit             = "rate_limit"
	OpenAIErrorCodeServerError           = "server_error"
	OpenAIErrorCodeBadGateway            = "bad_gateway"
	OpenAIErrorCodeServiceUnavailable    = "service_unavailable"
)

// OpenAIError OpenAI 标准错误结构
type OpenAIError struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// OpenAIErrorResponse OpenAI 错误响应
type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

// NewOpenAIError 创建 OpenAI 错误
func NewOpenAIError(message, errType string, param, code *string) OpenAIError {
	return OpenAIError{
		Message: message,
		Type:    errType,
		Param:   param,
		Code:    code,
	}
}

// NewOpenAIErrorResponse 创建 OpenAI 错误响应
func NewOpenAIErrorResponse(message, errType string, param, code *string) OpenAIErrorResponse {
	return OpenAIErrorResponse{
		Error: NewOpenAIError(message, errType, param, code),
	}
}

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}

// RespondWithOpenAIError 使用 OpenAI 格式返回错误
func RespondWithOpenAIError(c *gin.Context, statusCode int, message, errType string, param, code *string) {
	c.JSON(statusCode, NewOpenAIErrorResponse(message, errType, param, code))
}

// RespondWithInvalidRequestError 返回无效请求错误
func RespondWithInvalidRequestError(c *gin.Context, message string, param *string) {
	code := OpenAIErrorCodeInvalidAPIKey
	if param != nil {
		code = "invalid_" + *param
	}
	RespondWithOpenAIError(c, http.StatusBadRequest, message, OpenAIErrorTypeInvalidRequest, param, &code)
}

// RespondWithAuthenticationError 返回认证错误
func RespondWithAuthenticationError(c *gin.Context, message string) {
	code := OpenAIErrorCodeInvalidAPIKey
	RespondWithOpenAIError(c, http.StatusUnauthorized, message, OpenAIErrorTypeAuthentication, nil, &code)
}

// RespondWithPermissionError 返回权限错误
func RespondWithPermissionError(c *gin.Context, message string) {
	RespondWithOpenAIError(c, http.StatusForbidden, message, OpenAIErrorTypePermission, nil, nil)
}

// RespondWithNotFoundError 返回未找到错误
func RespondWithNotFoundError(c *gin.Context, message string, param *string) {
	code := OpenAIErrorCodeModelNotFound
	RespondWithOpenAIError(c, http.StatusNotFound, message, OpenAIErrorTypeNotFound, param, &code)
}

// RespondWithRateLimitError 返回速率限制错误
func RespondWithRateLimitError(c *gin.Context, message string) {
	code := OpenAIErrorCodeRateLimit
	RespondWithOpenAIError(c, http.StatusTooManyRequests, message, OpenAIErrorTypeRateLimit, nil, &code)
}

// RespondWithAPIError 返回 API 错误
func RespondWithAPIError(c *gin.Context, statusCode int, message string) {
	code := OpenAIErrorCodeServerError
	if statusCode == http.StatusBadGateway {
		code = OpenAIErrorCodeBadGateway
	} else if statusCode == http.StatusServiceUnavailable {
		code = OpenAIErrorCodeServiceUnavailable
	}
	RespondWithOpenAIError(c, statusCode, message, OpenAIErrorTypeAPIError, nil, &code)
}

// ParseUpstreamError 解析上游错误响应
func ParseUpstreamError(body []byte, statusCode int) (OpenAIErrorResponse, bool) {
	var errResp OpenAIErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return OpenAIErrorResponse{}, false
	}

	// 如果解析成功但错误消息为空，返回 false
	if errResp.Error.Message == "" {
		return OpenAIErrorResponse{}, false
	}

	return errResp, true
}
