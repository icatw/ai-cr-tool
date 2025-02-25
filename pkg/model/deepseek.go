package model

const (
	DeepSeekAPIURL = "https://api.siliconflow.cn/v1/chat/completions"
)

// DeepSeekClient 实现DeepSeek API的客户端
type DeepSeekClient struct {
	*BaseModelClient
}

// NewDeepSeekClient 创建新的DeepSeek客户端实例
func NewDeepSeekClient(cfg *Config) *DeepSeekClient {
	return &DeepSeekClient{
		BaseModelClient: NewBaseModelClient(cfg),
	}
}

// Chat 发送聊天请求并获取响应
func (c *DeepSeekClient) Chat(req *ChatRequest) (*ChatResponse, error) {
	// 应用基础配置
	c.ApplyConfig(req)

	// 发送请求并获取响应
	var resp ChatResponse
	err := c.httpClient.SendRequest(DeepSeekAPIURL, req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
