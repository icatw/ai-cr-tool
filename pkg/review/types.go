package review

// Issue 表示代码评审中发现的问题
type Issue struct {
	Title       string // 问题标题
	FilePath    string // 问题所在文件路径
	Line        int    // 问题所在行号
	Severity    string // 问题严重程度
	Description string // 问题描述信息
	Suggestion  string // 修复建议
	CodeSnippet string // 相关代码片段
}

// Severity levels for review issues
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)
