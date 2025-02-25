package model

const (
	ChatGLMAPIURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
)

// ChatGLMClient 实现智普AI ChatGLM API的客户端
type ChatGLMClient struct {
	*BaseModelClient
}

// NewChatGLMClient 创建新的ChatGLM客户端实例
func NewChatGLMClient(cfg *Config) *ChatGLMClient {
	return &ChatGLMClient{
		BaseModelClient: NewBaseModelClient(cfg),
	}
}

// Chat 发送聊天请求并获取响应
func (c *ChatGLMClient) Chat(req *ChatRequest) (*ChatResponse, error) {
	// 应用基础配置
	c.ApplyConfig(req)

	// 发送请求并获取响应
	var resp ChatResponse
	err := c.httpClient.SendRequest(ChatGLMAPIURL, req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
