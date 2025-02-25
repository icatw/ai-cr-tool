package review

// Format 定义支持的报告格式
type Format = ReportFormat

// String 返回Format的字符串表示
func (f Format) String() string {
	return string(f)
}

// IsValid 检查格式是否有效
func (f Format) IsValid() bool {
	switch f {
	case MarkdownFormat, HTMLFormat, PDFFormat:
		return true
	default:
		return false
	}
}
