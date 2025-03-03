package cache

import (
	"container/list"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ReviewCache 代码评审缓存
type ReviewCache struct {
	// 缓存目录路径
	cacheDir string
	// 内存缓存
	memCache    map[string]*list.Element
	lru         *list.List
	maxItems    int
	mutex       sync.RWMutex
	maxFileSize int64
	// 清理间隔
	cleanupInterval time.Duration
	// 停止清理信号
	stopCleanup chan struct{}
	// 缓存容量限制（字节）
	maxCacheSize int64
	currentSize  int64
	// 子目录数量
	subDirCount int
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
	// 最后访问时间
	LastAccessed time.Time `json:"last_accessed"`
}

// NewReviewCache 创建新的评审缓存管理器
func NewReviewCache(cacheDir string) (*ReviewCache, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %v", err)
	}

	cache := &ReviewCache{
		cacheDir:        cacheDir,
		memCache:        make(map[string]*list.Element),
		lru:             list.New(),
		maxItems:        100,             // 减少内存缓存项数量
		maxFileSize:     5 * 1024 * 1024, // 限制单文件大小为5MB
		cleanupInterval: 1 * time.Hour,   // 更频繁地清理缓存
		stopCleanup:     make(chan struct{}),
		maxCacheSize:    100 * 1024 * 1024, // 限制总缓存大小为100MB
		subDirCount:     16,                // 减少子目录数量
	}

	go cache.startCleanupRoutine()
	return cache, nil
}

// hashContent 计算内容的哈希值
func (c *ReviewCache) hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// Get 获取缓存的评审结果
func (c *ReviewCache) Get(content string) (*CacheItem, error) {
	contentHash := c.hashContent(content)

	// 先从内存缓存中查找
	c.mutex.RLock()
	if elem, ok := c.memCache[contentHash]; ok {
		item := elem.Value.(*CacheItem)
		// 更新访问时间
		item.LastAccessed = time.Now()
		// 移动到LRU链表头部
		c.lru.MoveToFront(elem)
		c.mutex.RUnlock()
		return item, nil
	}
	c.mutex.RUnlock()

	// 从文件中读取
	cacheFile := filepath.Join(c.cacheDir, contentHash+".json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

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

	// 添加到内存缓存
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果缓存已满，移除最久未使用的项
	if c.lru.Len() >= c.maxItems {
		c.removeOldest()
	}

	// 更新访问时间
	item.LastAccessed = time.Now()
	elem := c.lru.PushFront(&item)
	c.memCache[contentHash] = elem

	return &item, nil
}

// Set 设置评审结果缓存
func (c *ReviewCache) Set(content string, result string, expireAfter *time.Duration) error {
	// 检查内容大小
	contentSize := int64(len(content))
	if contentSize > c.maxFileSize {
		return fmt.Errorf("内容大小超过限制")
	}

	// 检查缓存总容量
	if c.currentSize+contentSize > c.maxCacheSize {
		// 触发清理
		if err := c.Clear(); err != nil {
			return fmt.Errorf("清理缓存失败: %v", err)
		}
		// 如果清理后仍然超出容量，返回错误
		if c.currentSize+contentSize > c.maxCacheSize {
			return fmt.Errorf("缓存容量已满")
		}
	}

	// 创建缓存项
	item := CacheItem{
		ContentHash:  c.hashContent(content),
		ReviewResult: result,
		CachedAt:     time.Now(),
		LastAccessed: time.Now(),
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

	// 计算子目录
	subDir := fmt.Sprintf("%02x", item.ContentHash[0])
	subDirPath := filepath.Join(c.cacheDir, subDir)

	// 确保子目录存在
	if err := os.MkdirAll(subDirPath, 0755); err != nil {
		return fmt.Errorf("创建子目录失败: %v", err)
	}

	// 写入文件缓存
	cacheFile := filepath.Join(subDirPath, item.ContentHash+".json")
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return err
	}

	// 更新内存缓存
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果缓存已满，移除最久未使用的项
	if c.lru.Len() >= c.maxItems {
		c.removeOldest()
	}

	// 添加到LRU缓存
	elem := c.lru.PushFront(&item)
	c.memCache[item.ContentHash] = elem

	return nil
}

// removeOldest 移除最久未使用的缓存项
func (c *ReviewCache) removeOldest() {
	elem := c.lru.Back()
	if elem != nil {
		c.lru.Remove(elem)
		item := elem.Value.(*CacheItem)
		delete(c.memCache, item.ContentHash)
	}
}

// Clear 清理过期的缓存文件
// startCleanupRoutine 启动定期清理routine
func (c *ReviewCache) startCleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.Clear(); err != nil {
				fmt.Printf("清理缓存失败: %v\n", err)
			}
		case <-c.stopCleanup:
			return
		}
	}
}

// Stop 停止缓存清理routine
func (c *ReviewCache) Stop() {
	close(c.stopCleanup)
}

func (c *ReviewCache) Clear() error {
	c.mutex.Lock()
	c.memCache = make(map[string]*list.Element)
	c.lru.Init()
	c.currentSize = 0
	c.mutex.Unlock()

	for i := 0; i < c.subDirCount; i++ {
		subDir := fmt.Sprintf("%02x", i)
		subDirPath := filepath.Join(c.cacheDir, subDir)

		if _, err := os.Stat(subDirPath); os.IsNotExist(err) {
			continue
		}

		files, err := os.ReadDir(subDirPath)
		if err != nil {
			fmt.Printf("读取子目录失败 %s: %v\n", subDirPath, err)
			continue
		}

		for _, file := range files {
			if filepath.Ext(file.Name()) != ".json" {
				continue
			}

			filePath := filepath.Join(subDirPath, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var item CacheItem
			if err := json.Unmarshal(data, &item); err != nil {
				continue
			}

			// 清理过期或超过1天未访问的缓存
			if (item.ExpireAt != nil && time.Now().After(*item.ExpireAt)) ||
				time.Since(item.LastAccessed) > 24*time.Hour {
				os.Remove(filePath)
				c.currentSize -= info.Size()
			}
		}
	}

	return nil
}

// BatchGet 批量获取缓存项
func (c *ReviewCache) BatchGet(contents []string) (map[string]*CacheItem, error) {
	result := make(map[string]*CacheItem)
	var wg sync.WaitGroup

	// 创建工作池
	workers := make(chan struct{}, 10) // 最多10个并发
	var mu sync.Mutex

	for _, content := range contents {
		wg.Add(1)
		workers <- struct{}{}

		go func(content string) {
			defer wg.Done()
			defer func() { <-workers }()

			item, err := c.Get(content)
			if err == nil && item != nil {
				mu.Lock()
				result[content] = item
				mu.Unlock()
			}
		}(content)
	}

	wg.Wait()
	return result, nil
}

// BatchSet 批量设置缓存项
func (c *ReviewCache) BatchSet(items map[string]string, expireAfter *time.Duration) error {
	var wg sync.WaitGroup
	workers := make(chan struct{}, 10)
	errors := make(chan error, len(items))

	for content, result := range items {
		wg.Add(1)
		workers <- struct{}{}

		go func(content, result string) {
			defer wg.Done()
			defer func() { <-workers }()

			if err := c.Set(content, result, expireAfter); err != nil {
				errors <- fmt.Errorf("设置缓存失败 [%s]: %v", content, err)
			}
		}(content, result)
	}

	wg.Wait()
	close(errors)

	// 收集所有错误
	var errMsgs []string
	for err := range errors {
		errMsgs = append(errMsgs, err.Error())
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf("批量设置缓存出现错误:\n%s", strings.Join(errMsgs, "\n"))
	}

	return nil
}
