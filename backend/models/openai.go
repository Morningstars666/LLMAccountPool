package models

import (
	"encoding/json"
)

// ChatCompletionRequest 表示 OpenAI Chat Completions API 的请求参数
type ChatCompletionRequest struct {
	Model             string                 `json:"model" binding:"required"`
	Messages          []ChatCompletionMessage `json:"messages" binding:"required,min=1"`
	MaxTokens         *int                   `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                 `json:"max_completion_tokens,omitempty"`
	Temperature       *float64               `json:"temperature,omitempty"`
	TopP              *float64               `json:"top_p,omitempty"`
	N                 *int                   `json:"n,omitempty"`
	Stream            bool                   `json:"stream,omitempty"`
	StreamOptions     *StreamOptions         `json:"stream_options,omitempty"`
	Stop              interface{}            `json:"stop,omitempty"`
	PresencePenalty   *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64               `json:"frequency_penalty,omitempty"`
	LogitBias         map[string]int         `json:"logit_bias,omitempty"`
	User              string                 `json:"user,omitempty"`
	Seed              *int                   `json:"seed,omitempty"`
	Tools             []Tool                 `json:"tools,omitempty"`
	ToolChoice        interface{}            `json:"tool_choice,omitempty"`
	ResponseFormat    *ResponseFormat        `json:"response_format,omitempty"`
}

// ChatCompletionMessage 表示聊天消息
type ChatCompletionMessage struct {
	Role       string     `json:"role" binding:"required,oneof=system user assistant tool function"`
	Content    interface{} `json:"content"` // string 或 []MessageContent
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// MessageContent 用于多模态消息内容
type MessageContent struct {
	Type     string          `json:"type" binding:"required,oneof=text image_url"`
	Text     string          `json:"text,omitempty"`
	ImageURL *ImageURL       `json:"image_url,omitempty"`
}

// ImageURL 表示图片URL
type ImageURL struct {
	URL    string `json:"url" binding:"required"`
	Detail string `json:"detail,omitempty"`
}

// StreamOptions 流式选项
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// Tool 表示工具定义
type Tool struct {
	Type     string   `json:"type" binding:"required"`
	Function Function `json:"function" binding:"required"`
}

// Function 表示函数定义
type Function struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolCall 表示工具调用
type ToolCall struct {
	ID       string       `json:"id" binding:"required"`
	Type     string       `json:"type" binding:"required"`
	Function ToolFunction `json:"function" binding:"required"`
}

// ToolFunction 表示工具函数调用
type ToolFunction struct {
	Name      string `json:"name" binding:"required"`
	Arguments string `json:"arguments" binding:"required"`
}

// ResponseFormat 响应格式
type ResponseFormat struct {
	Type       string                 `json:"type" binding:"required,oneof=text json_object json_schema"`
	JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
}

// ChatCompletionResponse 表示 OpenAI Chat Completions API 的非流式响应
type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             *Usage                 `json:"usage,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
}

// ChatCompletionChoice 表示响应选择
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason,omitempty"`
	Logprobs     *Logprobs             `json:"logprobs,omitempty"`
}

// ChatCompletionStreamResponse 表示流式响应
type ChatCompletionStreamResponse struct {
	ID                string                      `json:"id"`
	Object            string                      `json:"object"`
	Created           int64                       `json:"created"`
	Model             string                      `json:"model"`
	Choices           []ChatCompletionStreamChoice `json:"choices"`
	Usage             *Usage                      `json:"usage,omitempty"`
	SystemFingerprint string                      `json:"system_fingerprint,omitempty"`
}

// ChatCompletionStreamChoice 表示流式响应选择
type ChatCompletionStreamChoice struct {
	Index        int                   `json:"index"`
	Delta        ChatCompletionDelta   `json:"delta"`
	FinishReason string                `json:"finish_reason,omitempty"`
	Logprobs     *Logprobs             `json:"logprobs,omitempty"`
}

// ChatCompletionDelta 表示流式增量
type ChatCompletionDelta struct {
	Role       string     `json:"role,omitempty"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// Usage 表示令牌使用情况
type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	PromptTokensDetails     *PromptTokensDetails    `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

// PromptTokensDetails 提示令牌详情
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

// CompletionTokensDetails 完成令牌详情
type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`
	AudioTokens              int `json:"audio_tokens,omitempty"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"`
}

// Logprobs 对数概率
type Logprobs struct {
	Content []LogprobContent `json:"content,omitempty"`
}

// LogprobContent 对数概率内容
type LogprobContent struct {
	Token       string             `json:"token"`
	Logprob     float64            `json:"logprob"`
	Bytes       []int              `json:"bytes,omitempty"`
	TopLogprobs []TopLogprobDetail `json:"top_logprobs,omitempty"`
}

// TopLogprobDetail 顶部对数概率详情
type TopLogprobDetail struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// OpenAIErrorResponse OpenAI 错误响应格式
type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

// OpenAIError OpenAI 错误结构
type OpenAIError struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// Validate 验证请求参数
func (r *ChatCompletionRequest) Validate() error {
	// 验证 temperature
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return NewValidationError("temperature", "must be between 0 and 2")
	}

	// 验证 top_p
	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return NewValidationError("top_p", "must be between 0 and 1")
	}

	// 验证 presence_penalty
	if r.PresencePenalty != nil && (*r.PresencePenalty < -2 || *r.PresencePenalty > 2) {
		return NewValidationError("presence_penalty", "must be between -2 and 2")
	}

	// 验证 frequency_penalty
	if r.FrequencyPenalty != nil && (*r.FrequencyPenalty < -2 || *r.FrequencyPenalty > 2) {
		return NewValidationError("frequency_penalty", "must be between -2 and 2")
	}

	// 验证 n
	if r.N != nil && *r.N < 1 {
		return NewValidationError("n", "must be at least 1")
	}

	// 验证 stop
	if r.Stop != nil {
		switch v := r.Stop.(type) {
		case string:
			if len(v) > 4 {
				return NewValidationError("stop", "string version must not exceed 4 characters")
			}
		case []interface{}:
			if len(v) > 4 {
				return NewValidationError("stop", "array version must not exceed 4 elements")
			}
			for _, item := range v {
				if str, ok := item.(string); ok && len(str) > 4 {
					return NewValidationError("stop", "each string in array must not exceed 4 characters")
				}
			}
		}
	}

	// 验证 logit_bias
	if r.LogitBias != nil {
		for tokenID, bias := range r.LogitBias {
			if bias < -100 || bias > 100 {
				return NewValidationError("logit_bias", "bias for token "+tokenID+" must be between -100 and 100")
			}
		}
	}

	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Param   string `json:"param"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Param + ": " + e.Message
}

// NewValidationError 创建验证错误
func NewValidationError(param, message string) *ValidationError {
	return &ValidationError{
		Param:   param,
		Message: message,
	}
}

// UnmarshalContent 解析消息内容
func (m *ChatCompletionMessage) UnmarshalContent() (string, error) {
	switch v := m.Content.(type) {
	case string:
		return v, nil
	case []interface{}:
		// 多模态内容，返回序列化后的JSON
		data, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case nil:
		return "", nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}
