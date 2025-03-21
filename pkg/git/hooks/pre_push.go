package hooks

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/icatw/ai-cr-tool/pkg/cache"

	"github.com/icatw/ai-cr-tool/pkg/git"
	"github.com/icatw/ai-cr-tool/pkg/model"
	"github.com/icatw/ai-cr-tool/pkg/review"
	"github.com/icatw/ai-cr-tool/pkg/types"
)

// PrePushHook 处理pre-push钩子的逻辑
type PrePushHook struct {
	Options map[string]string
}

// NewPrePushHook 创建新的pre-push钩子处理器
func NewPrePushHook(options map[string]string) *PrePushHook {
	return &PrePushHook{
		Options: options,
	}
}

// Execute 执行pre-push钩子逻辑
func (h *PrePushHook) Execute() error {
	// 获取标准输入中的引用信息
	refInfo, err := h.readRefInfo()
	if err != nil {
		return fmt.Errorf("failed to read ref info: %v", err)
	}

	// 解析引用信息
	refs := h.parseRefInfo(refInfo)
	if len(refs) == 0 {
		return fmt.Errorf("no refs to push")
	}

	// 对每个要推送的引用进行代码评审
	for _, ref := range refs {
		if err := h.reviewRef(ref); err != nil {
			return fmt.Errorf("review failed for ref %s: %v", ref.Name, err)
		}
	}

	return nil
}

// RefInfo 存储引用信息
type RefInfo struct {
	Name    string
	OldHash string
	NewHash string
	Remote  string
}

// readRefInfo 从标准输入读取引用信息
func (h *PrePushHook) readRefInfo() (string, error) {
	info, err := os.ReadFile(os.Stdin.Name())
	if err != nil {
		return "", err
	}
	return string(info), nil
}

// parseRefInfo 解析引用信息
func (h *PrePushHook) parseRefInfo(info string) []RefInfo {
	var refs []RefInfo
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 4 {
			continue
		}

		refs = append(refs, RefInfo{
			Name:    parts[2],
			OldHash: parts[0],
			NewHash: parts[1],
			Remote:  parts[3],
		})
	}

	return refs
}

// reviewRef 对指定引用进行代码评审
func (h *PrePushHook) reviewRef(ref RefInfo) error {
	// 如果是删除分支操作，则跳过评审
	if ref.NewHash == "0000000000000000000000000000000000000000" {
		return nil
	}

	// 创建Git客户端
	gitClient := git.NewGitClient(h.Options["repo_path"])

	// 创建代码分析器
	analyzer := review.NewAnalyzer(gitClient)

	// 获取代码改动
	changes, err := analyzer.AnalyzeChanges(ref.OldHash, ref.NewHash)
	if err != nil {
		return fmt.Errorf("分析代码改动失败: %v", err)
	}

	// 如果没有改动，直接返回
	if len(changes) == 0 {
		return nil
	}

	// 初始化缓存
	cacheManager, err := cache.NewReviewCache(h.Options["cache_dir"])
	if err != nil {
		return fmt.Errorf("初始化缓存失败: %v", err)
	}

	// 初始化AI模型客户端
	modelCfg := model.NewModelConfigWithKeys(h.Options["api_key"], "", "", "")

	// 创建模型管理器
	modelManager, err := model.NewModelManager(modelCfg)
	if err != nil {
		return fmt.Errorf("初始化模型管理器失败: %v", err)
	}

	// 获取默认模型客户端
	modelClient, err := modelManager.GetClient("")
	if err != nil {
		return fmt.Errorf("获取模型客户端失败: %v", err)
	}

	// 创建评审提示模板
	prompt := model.DefaultReviewPrompt()

	// 创建评审报告生成器
	reporter := review.NewReporter("ai-cr-tool", ref.NewHash)

	// 分析代码问题
	var issues []types.Issue
	for _, change := range changes {
		// 检查缓存
		if cached, err := cacheManager.Get(change.DiffContent); err == nil && cached != nil {
			issues = append(issues, types.Issue{
				Title:       "AI代码评审结果",
				FilePath:    change.FilePath,
				Severity:    types.SeverityInfo,
				Description: cached.ReviewResult,
				Suggestion:  "请根据AI评审建议进行相应修改",
			})
			continue
		}

		// 生成评审提示
		messages := prompt.GeneratePrompt(change.FilePath, change.ChangeType, change.DiffContent)

		// 调用AI进行评审
		req := &model.ChatRequest{
			Model:       modelCfg.Models[modelCfg.DefaultModel].Model,
			Messages:    messages,
			MaxTokens:   modelCfg.Models[modelCfg.DefaultModel].MaxTokens,
			Temperature: modelCfg.Models[modelCfg.DefaultModel].Temperature,
		}

		resp, err := modelClient.Chat(req)
		if err != nil {
			return fmt.Errorf("评审失败 - %s: %v", change.FilePath, err)
		}

		reviewResult := resp.Choices[0].Message.Content

		// 缓存评审结果
		expireAfter := 24 * time.Hour
		if err := cacheManager.Set(change.DiffContent, reviewResult, &expireAfter); err != nil {
			return fmt.Errorf("缓存评审结果失败: %v", err)
		}

		// 添加评审结果
		issues = append(issues, types.Issue{
			Title:       "AI代码评审结果",
			FilePath:    change.FilePath,
			Severity:    types.SeverityInfo,
			Description: reviewResult,
			Suggestion:  "请根据AI评审建议进行相应修改",
		})
	}

	// 生成评审报告
	reportContent, err := reporter.Generate(issues, review.MarkdownFormat)
	if err != nil {
		return fmt.Errorf("生成评审报告失败: %v", err)
	}

	// 如果评审发现问题，返回错误
	if len(issues) > 0 {
		return fmt.Errorf("代码评审发现问题:\n%s", reportContent)
	}

	return nil
}
