package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient 封装基础的 HTTP 客户端功能
type HTTPClient struct {
	client *http.Client
	config *Config
}

// NewHTTPClient 创建新的 HTTP 客户端实例
func NewHTTPClient(cfg *Config) *HTTPClient {
	return &HTTPClient{
		config: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// SendRequest 发送 HTTP 请求并处理响应
func (c *HTTPClient) SendRequest(url string, req interface{}, resp interface{}) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %v", err)
	}

	// 添加重试机制
	var httpResp *http.Response
	var lastErr error
	for retries := 0; retries < 3; retries++ {
		httpReq, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("create request failed: %v", err)
		}

		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
		httpReq.Header.Set("Content-Type", "application/json")

		httpResp, err = c.client.Do(httpReq)
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(time.Duration(retries+1) * time.Second)
	}

	if httpResp == nil {
		return fmt.Errorf("all retries failed: %v", lastErr)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", httpResp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("unmarshal response failed: %v", err)
	}

	return nil
}
