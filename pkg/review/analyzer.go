package review

import (
	"fmt"
	"strings"
	"sync"

	"github.com/icatw/ai-cr-tool/pkg/git"
)

// FileChange 表示文件改动的信息
type FileChange struct {
	FilePath    string
	ChangeType  string // "added", "modified", "deleted"
	OldContent  string
	NewContent  string
	DiffContent string
	Lines       []string // 代码行内容
}

// Analyzer 代码分析器
type Analyzer struct {
	gitClient *git.GitClient
	Progress  chan ProgressInfo
}

// ProgressInfo 进度信息
type ProgressInfo struct {
	CurrentFile  string
	TotalFiles   int
	CurrentIndex int
	Percentage   float64
}

// NewAnalyzer 创建新的代码分析器
func NewAnalyzer(gitClient *git.GitClient) *Analyzer {
	return &Analyzer{
		gitClient: gitClient,
		Progress:  make(chan ProgressInfo, 1),
	}
}

// AnalyzeChanges 分析代码改动
func (a *Analyzer) AnalyzeChanges(from, to string) ([]FileChange, error) {
	// 获取改动的文件列表
	files, err := a.gitClient.GetChangedFiles(from, to)
	if err != nil {
		return nil, fmt.Errorf("获取改动文件列表失败: %v", err)
	}

	// 获取详细的差异内容
	diff, err := a.gitClient.GetDiff(from, to)
	if err != nil {
		return nil, fmt.Errorf("获取差异内容失败: %v", err)
	}

	// 并行处理文件改动
	changes := make([]FileChange, len(files))
	var wg sync.WaitGroup
	workers := 4 // 设置合适的并发数
	semaphore := make(chan struct{}, workers)

	for i, file := range files {
		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			change := FileChange{
				FilePath: filePath,
			}

			// 更新进度信息
			a.Progress <- ProgressInfo{
				CurrentFile:  filePath,
				TotalFiles:   len(files),
				CurrentIndex: index + 1,
				Percentage:   float64(index+1) / float64(len(files)) * 100,
			}

			// 根据差异内容判断改动类型
			if strings.Contains(diff, fmt.Sprintf("a/%s", filePath)) && strings.Contains(diff, fmt.Sprintf("b/%s", filePath)) {
				change.ChangeType = "modified"
				// 获取新文件内容
				newContent, err := a.gitClient.GetFileContent(filePath, to)
				if err == nil {
					change.NewContent = newContent
					// 将新文件内容按行分割
					change.Lines = strings.Split(newContent, "\n")
				}
			} else if strings.Contains(diff, fmt.Sprintf("a/%s", filePath)) {
				change.ChangeType = "deleted"
			} else {
				change.ChangeType = "added"
				// 获取新文件内容
				newContent, err := a.gitClient.GetFileContent(filePath, to)
				if err == nil {
					change.NewContent = newContent
					// 将新文件内容按行分割
					change.Lines = strings.Split(newContent, "\n")
				}
			}

			// 提取该文件的差异内容
			change.DiffContent = extractFileDiff(diff, filePath)
			changes[index] = change
		}(i, file)
	}

	wg.Wait()
	return changes, nil
}

// extractFileDiff 从完整的差异内容中提取指定文件的差异
func extractFileDiff(diff, filePath string) string {
	lines := strings.Split(diff, "\n")
	var fileDiff strings.Builder
	inFile := false

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("diff --git a/%s b/%s", filePath, filePath)) {
			inFile = true
		} else if strings.HasPrefix(line, "diff --git") && inFile {
			break
		}

		if inFile {
			fileDiff.WriteString(line)
			fileDiff.WriteString("\n")
		}
	}

	return fileDiff.String()
}
