package model

import "fmt"

// ModelClient 定义通用的AI模型客户端接口
type ModelClient interface {
	// Chat 发送聊天请求并获取响应
	Chat(req *ChatRequest) (*ChatResponse, error)
}

// ModelConfig 定义全局模型配置
type ModelConfig struct {
	// 默认使用的模型类型
	DefaultModel string `json:"default_model"`
	// 模型配置映射
	Models map[string]*Config `json:"models"`
}

// Config 定义单个模型配置
type Config struct {
	// 模型类型（如 "deepseek", "openai" 等）
	Type string `json:"type"`
	// API密钥
	APIKey string `json:"api_key"`
	// 模型名称
	Model string `json:"model"`
	// 其他通用配置参数
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	// 模型特定的配置参数
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
}

// ChatRequest 定义聊天请求的参数结构
type ChatRequest struct {
	Model            string            `json:"model"`
	Messages         []Message         `json:"messages"`
	Stream           bool              `json:"stream"`
	MaxTokens        int               `json:"max_tokens"`
	Stop             []string          `json:"stop,omitempty"`
	Temperature      float64           `json:"temperature"`
	TopP             float64           `json:"top_p"`
	FrequencyPenalty float64           `json:"frequency_penalty"`
	N                int               `json:"n"`
	ResponseFormat   map[string]string `json:"response_format"`
	Tools            []Tool            `json:"tools,omitempty"`
}

// Message 定义聊天消息的结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool 定义工具的结构
type Tool struct {
	Type     string         `json:"type"`
	Function FunctionConfig `json:"function"`
}

// FunctionConfig 定义函数配置的结构
type FunctionConfig struct {
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Parameters  map[string]interface{} `json:"parameters"`
	Strict      bool                   `json:"strict"`
}

// ChatResponse 定义API响应的结构
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 定义响应选项的结构
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 定义token使用情况的结构
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolCall 定义工具调用的结构
type ToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function FunctionResult `json:"function"`
}

// FunctionResult 定义函数调用结果的结构
type FunctionResult struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// BaseModelClient 提供基础的模型客户端实现
type BaseModelClient struct {
	httpClient *HTTPClient
	config     *Config
}

// NewBaseModelClient 创建基础模型客户端
func NewBaseModelClient(cfg *Config) *BaseModelClient {
	return &BaseModelClient{
		httpClient: NewHTTPClient(cfg),
		config:     cfg,
	}
}

// ApplyConfig 应用配置到请求
func (c *BaseModelClient) ApplyConfig(req *ChatRequest) {
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = c.config.Temperature
	}
}

// NewModelClient 根据配置创建对应的模型客户端
func NewModelClient(cfg *Config) (ModelClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	switch cfg.Type {
	case "deepseek":
		return NewDeepSeekClient(cfg), nil
	case "openai":
		return NewOpenAIClient(cfg), nil
	case "chatglm":
		return NewChatGLMClient(cfg), nil
	case "qwen":
		return NewQWENClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported model type: %s", cfg.Type)
	}
}

// ModelManager 管理多个模型客户端
type ModelManager struct {
	config  *ModelConfig
	clients map[string]ModelClient
}

// NewModelManager 创建模型管理器
func NewModelManager(config *ModelConfig) (*ModelManager, error) {
	if config == nil {
		return nil, fmt.Errorf("model config cannot be nil")
	}
	if config.DefaultModel == "" {
		return nil, fmt.Errorf("default model must be specified")
	}
	if len(config.Models) == 0 {
		return nil, fmt.Errorf("at least one model configuration is required")
	}

	return &ModelManager{
		config:  config,
		clients: make(map[string]ModelClient),
	}, nil
}

// GetClient 获取指定模型的客户端
func (m *ModelManager) GetClient(modelType string) (ModelClient, error) {
	// 如果未指定模型类型，使用默认模型
	if modelType == "" {
		modelType = m.config.DefaultModel
	}

	// 检查客户端是否已经创建
	if client, exists := m.clients[modelType]; exists {
		fmt.Printf("使用已创建的模型客户端: %s (模型: %s)\n", modelType, m.config.Models[modelType].Model)
		return client, nil
	}

	// 获取模型配置
	config, exists := m.config.Models[modelType]
	if !exists {
		return nil, fmt.Errorf("model config not found for type: %s", modelType)
	}

	// 创建新的客户端
	client, err := NewModelClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create model client: %v", err)
	}

	// 缓存客户端
	m.clients[modelType] = client
	fmt.Printf("创建新的模型客户端: %s (模型: %s)\n", modelType, config.Model)
	return client, nil
}
