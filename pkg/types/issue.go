package types

// SeverityLevel 定义问题严重程度
type SeverityLevel string

const (
	SeverityInfo    SeverityLevel = "info"
	SeverityWarning SeverityLevel = "warning"
	SeverityError   SeverityLevel = "error"
)

// Issue 表示代码评审发现的问题
type Issue struct {
	Title       string        // 问题标题
	FilePath    string        // 文件路径
	Line        int           // 行号
	Severity    SeverityLevel // 严重程度
	Description string        // 问题描述
	Suggestion  string        // 改进建议
	CodeSnippet string        // 相关代码片段
}
