package types

// FileChange 表示文件改动的信息
type FileChange struct {
	FilePath    string
	ChangeType  string // "added", "modified", "deleted"
	OldContent  string
	NewContent  string
	DiffContent string
	Lines       []string // 代码行内容
}
