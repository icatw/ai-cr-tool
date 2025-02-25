package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ReviewPrompt 定义代码评审的提示模板
type ReviewPrompt struct {
	// 基础提示信息
	BasePrompt string
	// 评审重点
	FocusAreas []string
	// 输出格式
	OutputFormat string
	// 语言相关的最佳实践
	LanguageBestPractices map[string][]string
}

// DefaultReviewPrompt 创建默认的代码评审提示模板
func DefaultReviewPrompt() *ReviewPrompt {
	return &ReviewPrompt{
		BasePrompt: "你是一个专业的代码评审助手，请基于以下几个方面进行评审：\n" +
			"1. 代码质量和可维护性\n" +
			"2. 性能优化建议\n" +
			"3. 安全性考虑\n" +
			"4. 最佳实践遵循情况",
		FocusAreas: []string{
			"代码结构和组织",
			"错误处理",
			"命名规范",
			"注释完整性",
			"测试覆盖",
		},
		OutputFormat: "markdown",
		LanguageBestPractices: map[string][]string{
			"go": {
				"使用 defer 释放资源",
				"错误处理遵循 Go 风格",
				"避免使用 panic",
				"使用 context 控制超时",
			},
		},
	}
}

// GeneratePrompt 根据代码差异生成完整的评审提示
func (p *ReviewPrompt) GeneratePrompt(filePath, changeType, diff string) []Message {
	// 获取文件扩展名
	ext := filepath.Ext(filePath)
	lang := strings.TrimPrefix(ext, ".")

	// 构建评审重点提示
	var focusPrompt strings.Builder
	focusPrompt.WriteString("\n评审重点关注：\n")
	for _, area := range p.FocusAreas {
		focusPrompt.WriteString(fmt.Sprintf("- %s\n", area))
	}

	// 添加语言特定的最佳实践
	if practices, ok := p.LanguageBestPractices[lang]; ok {
		focusPrompt.WriteString("\n语言特定的最佳实践：\n")
		for _, practice := range practices {
			focusPrompt.WriteString(fmt.Sprintf("- %s\n", practice))
		}
	}

	return []Message{
		{
			Role:    "system",
			Content: p.BasePrompt + focusPrompt.String(),
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("文件: %s\n改动类型: %s\n\n%s", filePath, changeType, diff),
		},
	}
}
