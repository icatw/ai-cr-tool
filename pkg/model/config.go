package model

// DefaultModelConfig 默认的全局模型配置
var DefaultModelConfig = ModelConfig{
	DefaultModel: "qwen",
	Models: map[string]*Config{
		"deepseek": {
			Type:        "deepseek",
			Model:       "deepseek-ai/DeepSeek-R1",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
		"openai": {
			Type:        "openai",
			Model:       "gpt-3.5-turbo",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
		"chatglm": {
			Type:        "chatglm",
			Model:       "glm-4",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
		"qwen": {
			Type:        "qwen",
			Model:       "qwen-coder-plus",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
	},
}

// NewModelConfigWithKeys 创建带有API密钥的模型配置
func NewModelConfigWithKeys(deepseekKey, openaiKey, chatglmKey, qwenKey string) *ModelConfig {
	// 创建新的配置
	cfg := ModelConfig{
		DefaultModel: "qwen", // 设置默认模型为qwen
		Models:       make(map[string]*Config),
	}

	// 只添加有API密钥的模型配置
	if qwenKey != "" {
		cfg.Models["qwen"] = &Config{
			Type:        "qwen",
			Model:       "qwen-coder-plus",
			MaxTokens:   2000,
			Temperature: 0.7,
			APIKey:      qwenKey,
			ExtraParams: make(map[string]interface{}),
		}
	}

	if deepseekKey != "" {
		cfg.Models["deepseek"] = &Config{
			Type:        "deepseek",
			Model:       "deepseek-ai/DeepSeek-R1",
			MaxTokens:   2000,
			Temperature: 0.7,
			APIKey:      deepseekKey,
			ExtraParams: make(map[string]interface{}),
		}
	}

	if openaiKey != "" {
		cfg.Models["openai"] = &Config{
			Type:        "openai",
			Model:       "gpt-3.5-turbo",
			MaxTokens:   2000,
			Temperature: 0.7,
			APIKey:      openaiKey,
			ExtraParams: make(map[string]interface{}),
		}
	}

	if chatglmKey != "" {
		cfg.Models["chatglm"] = &Config{
			Type:        "chatglm",
			Model:       "glm-4",
			MaxTokens:   2000,
			Temperature: 0.7,
			APIKey:      chatglmKey,
			ExtraParams: make(map[string]interface{}),
		}
	}

	return &cfg
}

// NewConfig 创建新的配置实例（保留用于兼容性）
func NewConfig(apiKey string) *Config {
	return &Config{
		Type:        "deepseek",
		Model:       "deepseek-ai/DeepSeek-R1",
		MaxTokens:   2000,
		Temperature: 0.7,
		APIKey:      apiKey,
	}
}
