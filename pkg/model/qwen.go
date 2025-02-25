package model

const (
	QWENAPIURL = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
)

// QWENClient 实现DeepSeek API的客户端
type QWENClient struct {
	*BaseModelClient
}

// NewQWENClient 创建新的DeepSeek客户端实例
func NewQWENClient(cfg *Config) *QWENClient {
	return &QWENClient{
		BaseModelClient: NewBaseModelClient(cfg),
	}
}

// Chat 发送聊天请求并获取响应
func (c *QWENClient) Chat(req *ChatRequest) (*ChatResponse, error) {
	// 应用基础配置
	c.ApplyConfig(req)

	// 发送请求并获取响应
	var resp ChatResponse
	err := c.httpClient.SendRequest(QWENAPIURL, req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
