package model

const (
	OpenAIAPIURL = "https://api.openai.com/v1/chat/completions"
)

// OpenAIClient 实现 OpenAI API 的客户端
type OpenAIClient struct {
	*BaseModelClient
}

// NewOpenAIClient 创建新的 OpenAI 客户端实例
func NewOpenAIClient(cfg *Config) *OpenAIClient {
	return &OpenAIClient{
		BaseModelClient: NewBaseModelClient(cfg),
	}
}

// Chat 发送聊天请求并获取响应
func (c *OpenAIClient) Chat(req *ChatRequest) (*ChatResponse, error) {
	// 应用基础配置
	c.ApplyConfig(req)

	// 发送请求并获取响应
	var resp ChatResponse
	err := c.httpClient.SendRequest(OpenAIAPIURL, req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
