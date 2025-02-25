package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ReviewCache 代码评审缓存
type ReviewCache struct {
	// 缓存目录路径
	cacheDir string
}

// CacheItem 缓存项
type CacheItem struct {
	// 文件改动内容的哈希值
	ContentHash string `json:"content_hash"`
	// 评审结果
	ReviewResult string `json:"review_result"`
	// 缓存时间
	CachedAt time.Time `json:"cached_at"`
	// 过期时间（可选）
	ExpireAt *time.Time `json:"expire_at,omitempty"`
}

// NewReviewCache 创建新的评审缓存管理器
func NewReviewCache(cacheDir string) (*ReviewCache, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %v", err)
	}

	return &ReviewCache{cacheDir: cacheDir}, nil
}

// Get 获取缓存的评审结果
func (c *ReviewCache) Get(content string) (*CacheItem, error) {
	// 计算内容哈希
	contentHash := c.hashContent(content)

	// 构建缓存文件路径
	cacheFile := filepath.Join(c.cacheDir, contentHash+".json")

	// 读取缓存文件
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// 解析缓存项
	var item CacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	// 检查是否过期
	if item.ExpireAt != nil && time.Now().After(*item.ExpireAt) {
		// 删除过期缓存
		if err := os.Remove(cacheFile); err != nil {
			return nil, fmt.Errorf("删除过期缓存文件失败: %v", err)
		}
		return nil, nil
	}

	return &item, nil
}

// Set 设置评审结果缓存
func (c *ReviewCache) Set(content string, result string, expireAfter *time.Duration) error {
	// 创建缓存项
	item := CacheItem{
		ContentHash:  c.hashContent(content),
		ReviewResult: result,
		CachedAt:     time.Now(),
	}

	// 设置过期时间（如果指定）
	if expireAfter != nil {
		expireAt := item.CachedAt.Add(*expireAfter)
		item.ExpireAt = &expireAt
	}

	// 序列化缓存项
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	// 写入缓存文件
	cacheFile := filepath.Join(c.cacheDir, item.ContentHash+".json")
	return os.WriteFile(cacheFile, data, 0644)
}

// hashContent 计算内容的哈希值
func (c *ReviewCache) hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// Clear 清理过期的缓存文件
func (c *ReviewCache) Clear() error {
	// 遍历缓存目录
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(c.cacheDir, file.Name())

		// 读取缓存项
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var item CacheItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}

		// 删除过期的缓存文件
		if item.ExpireAt != nil && time.Now().After(*item.ExpireAt) {
			if err := os.Remove(filePath); err != nil {
				// 记录错误但继续处理其他文件
				fmt.Printf("删除过期缓存文件失败 %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}
