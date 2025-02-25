package review

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// ReportFormat 定义支持的报告格式
type ReportFormat string

const (
	MarkdownFormat ReportFormat = "markdown"
	HTMLFormat     ReportFormat = "html"
	PDFFormat      ReportFormat = "pdf"
)

// Reporter 定义报告生成器接口
type Reporter interface {
	Generate(issues []Issue, format ReportFormat) ([]byte, error)
}

// DefaultReporter 默认报告生成器实现
type DefaultReporter struct {
	ProjectName string
	CommitID    string
}

// NewReporter 创建新的报告生成器
func NewReporter(projectName, commitID string) Reporter {
	return &DefaultReporter{
		ProjectName: projectName,
		CommitID:    commitID,
	}
}

// generateMarkdown 生成Markdown格式的报告
func (r *DefaultReporter) generateMarkdown(issues []Issue) ([]byte, error) {
	var buf bytes.Buffer

	// 写入报告头部
	buf.WriteString(fmt.Sprintf("# 代码评审报告\n\n"))
	buf.WriteString(fmt.Sprintf("## 项目信息\n\n"))
	buf.WriteString(fmt.Sprintf("- 项目名称：%s\n", r.ProjectName))
	buf.WriteString(fmt.Sprintf("- 提交ID：%s\n", r.CommitID))
	buf.WriteString(fmt.Sprintf("- 评审时间：%s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// 按严重程度分类统计
	severityCount := make(map[string]int)
	for _, issue := range issues {
		severityCount[issue.Severity]++
	}

	// 写入统计信息
	buf.WriteString("## 评审结果统计\n\n")

	// 添加代码统计信息
	buf.WriteString("### 代码变更统计\n\n")
	buf.WriteString("| 指标 | 数值 |\n")
	buf.WriteString("|------|---------|\n")
	buf.WriteString(fmt.Sprintf("| 评审文件数 | %d |\n", len(getUniqueFiles(issues))))
	buf.WriteString(fmt.Sprintf("| 问题总数 | %d |\n", len(issues)))

	// 写入严重程度统计
	buf.WriteString("\n### 问题严重程度分布\n\n")
	buf.WriteString("| 严重程度 | 数量 |\n")
	buf.WriteString("|---------|---------|\n")
	for severity, count := range severityCount {
		buf.WriteString(fmt.Sprintf("| %s | %d |\n", string(severity), count))
	}
	buf.WriteString("\n")

	// 写入优化建议总结
	buf.WriteString("## 整体优化建议\n\n")
	suggestions := summarizeSuggestions(issues)
	for _, suggestion := range suggestions {
		buf.WriteString(fmt.Sprintf("- %s\n", suggestion))
	}
	buf.WriteString("\n")

	// 写入详细问题列表
	buf.WriteString("## 详细问题列表\n\n")
	for i, issue := range issues {
		buf.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, issue.Title))
		buf.WriteString(fmt.Sprintf("- 文件：`%s`\n", issue.FilePath))
		buf.WriteString(fmt.Sprintf("- 位置：第%d行\n", issue.Line))
		buf.WriteString(fmt.Sprintf("- 严重程度：**%s**\n", issue.Severity))
		buf.WriteString(fmt.Sprintf("- 描述：%s\n", issue.Description))
		if issue.Suggestion != "" {
			buf.WriteString(fmt.Sprintf("- 建议：> %s\n", issue.Suggestion))
		}
		buf.WriteString("\n")

		// 添加代码片段（如果有）
		if issue.CodeSnippet != "" {
			// 获取代码片段的上下文
			lines := strings.Split(issue.CodeSnippet, "\n")
			contextStart := max(0, issue.Line-3)
			contextEnd := min(len(lines), issue.Line+3)

			buf.WriteString("```go\n")
			// 添加行号和语法高亮
			for i := contextStart; i < contextEnd; i++ {
				linePrefix := "  "
				if i == issue.Line-1 { // 高亮问题行
					linePrefix = ">"
				}
				buf.WriteString(fmt.Sprintf("%s %4d │ %s\n", linePrefix, i+1, lines[i]))
			}
			buf.WriteString("```\n\n")
		}
	}

	return buf.Bytes(), nil
}

// generateHTML 生成HTML格式的报告
func (r *DefaultReporter) generateHTML(issues []Issue) ([]byte, error) {
	var buf bytes.Buffer

	// 写入HTML头部
	buf.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>代码评审报告</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 20px; background: #f5f5f5; }
		.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
		.header { background: #fff; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin: 20px 0; }
		.stat-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); transition: transform 0.2s; }
		.stat-card:hover { transform: translateY(-2px); }
		.severity { display: inline-block; padding: 4px 10px; border-radius: 4px; font-size: 0.9em; font-weight: 500; }
		.critical { background: #dc3545; color: white; }
		.high { background: #fd7e14; color: white; }
		.medium { background: #ffc107; color: black; }
		.low { background: #28a745; color: white; }
		.issue { background: white; padding: 25px; margin: 15px 0; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.code { background: #1e1e1e; color: #d4d4d4; padding: 20px; border-radius: 8px; overflow-x: auto; font-family: 'Consolas', monospace; }
		.code .line-number { color: #858585; padding-right: 15px; user-select: none; }
		.code .highlight { background: rgba(255,255,0,0.1); display: block; }
		.suggestion { border-left: 4px solid #007bff; padding: 15px; margin: 15px 0; background: #f8f9fa; border-radius: 0 8px 8px 0; }
		.chart { margin: 20px 0; padding: 20px; background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.code-block { position: relative; margin: 1em 0; }
		.code-block pre { margin: 0; }
		.code-block .language-badge { position: absolute; top: 0; right: 0; padding: 4px 8px; background: rgba(0,0,0,0.5); color: #fff; border-radius: 0 8px 0 4px; font-size: 0.8em; }
		.issue-meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; margin-bottom: 15px; }
		.issue-meta-item { background: #f8f9fa; padding: 10px; border-radius: 4px; }
	</style>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"></script>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/vs2015.min.css">
	<script>hljs.highlightAll();</script>
</head>
<body>
	<div class="container">`)

	// 写入报告头部信息
	buf.WriteString(fmt.Sprintf(`
	<div class="header">
		<h1>代码评审报告</h1>
		<p>项目名称：%s</p>
		<p>提交ID：%s</p>
		<p>评审时间：%s</p>
	</div>`, r.ProjectName, r.CommitID, time.Now().Format("2006-01-02 15:04:05")))

	// 统计信息
	severityCount := make(map[string]int)
	for _, issue := range issues {
		severityCount[issue.Severity]++
	}

	// 写入统计卡片
	buf.WriteString(`
	<div class="stats">`)
	buf.WriteString(fmt.Sprintf(`
		<div class="stat-card">
			<h3>评审文件数</h3>
			<p>%d</p>
		</div>`, len(getUniqueFiles(issues))))
	buf.WriteString(fmt.Sprintf(`
		<div class="stat-card">
			<h3>问题总数</h3>
			<p>%d</p>
		</div>`, len(issues)))

	// 写入严重程度分布
	buf.WriteString(`
	<div class="stat-card">
		<h3>问题严重程度分布</h3>`)
	for severity, count := range severityCount {
		buf.WriteString(fmt.Sprintf(`
		<p><span class="severity %s">%s</span>: %d</p>`, strings.ToLower(string(severity)), severity, count))
	}
	buf.WriteString(`
	</div>
	</div>`)

	// 写入优化建议
	buf.WriteString(`
	<h2>整体优化建议</h2>
	<div class="suggestions">`)
	suggestions := summarizeSuggestions(issues)
	for _, suggestion := range suggestions {
		buf.WriteString(fmt.Sprintf(`
		<div class="suggestion">%s</div>`, suggestion))
	}
	buf.WriteString(`
	</div>`)

	// 写入详细问题列表
	buf.WriteString(`
	<h2>详细问题列表</h2>`)
	for i, issue := range issues {
		buf.WriteString(fmt.Sprintf(`
	<div class="issue">
		<h3>%d. %s</h3>
		<div class="issue-meta">
			<div class="issue-meta-item">
				<strong>文件：</strong>%s
			</div>
			<div class="issue-meta-item">
				<strong>位置：</strong>第%d行
			</div>
			<div class="issue-meta-item">
				<strong>严重程度：</strong><span class="severity %s">%s</span>
			</div>
		</div>
		<p><strong>描述：</strong>%s</p>`,
			i+1, issue.Title, issue.FilePath, issue.Line,
			strings.ToLower(string(issue.Severity)), issue.Severity, issue.Description))

		if issue.Suggestion != "" {
			buf.WriteString(fmt.Sprintf(`
		<div class="suggestion">%s</div>`, issue.Suggestion))
		}

		if issue.CodeSnippet != "" {
			buf.WriteString(`
		<pre class="code">`)
			lines := strings.Split(issue.CodeSnippet, "\n")
			contextStart := max(0, issue.Line-3)
			contextEnd := min(len(lines), issue.Line+3)

			for i := contextStart; i < contextEnd; i++ {
				linePrefix := "  "
				if i == issue.Line-1 {
					linePrefix = ">"
				}
				buf.WriteString(fmt.Sprintf("%s %4d │ %s\n", linePrefix, i+1, lines[i]))
			}
			buf.WriteString(`</pre>`)
		}

		buf.WriteString(`
	</div>`)
	}

	// 写入HTML尾部
	buf.WriteString(`
	</div>
</body>
</html>`)

	return buf.Bytes(), nil
}

// 辅助函数：获取唯一文件列表
func getUniqueFiles(issues []Issue) []string {
	filesMap := make(map[string]bool)
	for _, issue := range issues {
		filesMap[issue.FilePath] = true
	}

	// 将map转换为切片
	files := make([]string, 0, len(filesMap))
	for file := range filesMap {
		files = append(files, file)
	}
	return files
}

// generatePDF 生成PDF格式的报告
func (r *DefaultReporter) generatePDF(issues []Issue) ([]byte, error) {
	// 首先生成HTML报告
	htmlContent, err := r.generateHTML(issues)
	if err != nil {
		return nil, fmt.Errorf("生成HTML报告失败: %v", err)
	}

	// 创建临时文件存储HTML内容
	tmpHTML, err := os.CreateTemp("", "review-*.html")
	if err != nil {
		return nil, fmt.Errorf("创建临时HTML文件失败: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpHTML.Name()); err != nil {
			fmt.Printf("删除临时HTML文件失败: %v\n", err)
		}
	}()

	// 写入HTML内容
	if _, err := tmpHTML.Write(htmlContent); err != nil {
		return nil, fmt.Errorf("写入HTML内容失败: %v", err)
	}
	tmpHTML.Close()

	// 创建临时文件存储PDF内容
	tmpPDF, err := os.CreateTemp("", "review-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("创建临时PDF文件失败: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpPDF.Name()); err != nil {
			fmt.Printf("删除临时PDF文件失败: %v\n", err)
		}
	}()
	tmpPDF.Close()

	// 使用wkhtmltopdf将HTML转换为PDF
	cmd := exec.Command("wkhtmltopdf",
		"--enable-local-file-access",
		"--margin-top", "20",
		"--margin-right", "20",
		"--margin-bottom", "20",
		"--margin-left", "20",
		"--page-size", "A4",
		tmpHTML.Name(),
		tmpPDF.Name(),
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("转换PDF失败: %v", err)
	}

	// 读取生成的PDF文件
	pdfContent, err := os.ReadFile(tmpPDF.Name())
	if err != nil {
		return nil, fmt.Errorf("读取PDF文件失败: %v", err)
	}

	return pdfContent, nil
}

// Generate 生成评审报告
func (r *DefaultReporter) Generate(issues []Issue, format ReportFormat) ([]byte, error) {
	switch format {
	case MarkdownFormat:
		return r.generateMarkdown(issues)
	case HTMLFormat:
		return r.generateHTML(issues)
	case PDFFormat:
		return r.generatePDF(issues)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// summarizeSuggestions 汇总分析评审问题中的建议，生成整体优化建议列表
func summarizeSuggestions(issues []Issue) []string {
	// 使用map对建议进行分类和去重
	suggestionMap := make(map[string]int)
	for _, issue := range issues {
		if issue.Suggestion != "" {
			suggestionMap[issue.Suggestion]++
		}
	}

	// 将建议转换为切片并按出现频率排序
	type suggestionCount struct {
		suggestion string
		count      int
	}
	suggestions := make([]suggestionCount, 0, len(suggestionMap))
	for suggestion, count := range suggestionMap {
		suggestions = append(suggestions, suggestionCount{suggestion, count})
	}

	// 按出现频率降序排序
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].count > suggestions[j].count
	})

	// 生成最终的建议列表
	result := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		result = append(result, s.suggestion)
	}

	return result
}
