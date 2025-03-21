package review

import (
	"fmt"
	"strings"

	"github.com/icatw/ai-cr-tool/pkg/git"
	"github.com/icatw/ai-cr-tool/pkg/types"
)

// Analyzer 代码分析器
type Analyzer struct {
	gitClient *git.GitClient
}

// NewAnalyzer 创建新的代码分析器
func NewAnalyzer(gitClient *git.GitClient) *Analyzer {
	return &Analyzer{gitClient: gitClient}
}

// AnalyzeChanges 分析代码改动
func (a *Analyzer) AnalyzeChanges(from, to string) ([]types.FileChange, error) {
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

	// 解析差异内容
	changes := make([]types.FileChange, 0, len(files))
	for _, file := range files {
		change := types.FileChange{
			FilePath: file,
		}

		// 根据差异内容判断改动类型
		if strings.Contains(diff, fmt.Sprintf("a/%s", file)) && strings.Contains(diff, fmt.Sprintf("b/%s", file)) {
			change.ChangeType = "modified"
			// 获取新文件内容
			newContent, err := a.gitClient.GetFileContent(file, to)
			if err == nil {
				change.NewContent = newContent
				// 将新文件内容按行分割
				change.Lines = strings.Split(newContent, "\n")
			}
		} else if strings.Contains(diff, fmt.Sprintf("a/%s", file)) {
			change.ChangeType = "deleted"
		} else {
			change.ChangeType = "added"
			// 获取新文件内容
			newContent, err := a.gitClient.GetFileContent(file, to)
			if err == nil {
				change.NewContent = newContent
				// 将新文件内容按行分割
				change.Lines = strings.Split(newContent, "\n")
			}
		}

		// 提取该文件的差异内容
		change.DiffContent = extractFileDiff(diff, file)
		changes = append(changes, change)
	}

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

// AnalyzeFiles 分析指定文件的改动
func (a *Analyzer) AnalyzeFiles(files []string) ([]types.FileChange, error) {
	var changes []types.FileChange
	for _, file := range files {
		diff, err := a.gitClient.GetFileDiff(file)
		if err != nil {
			return nil, fmt.Errorf("获取文件 %s 的改动失败: %v", file, err)
		}
		if diff != "" {
			changes = append(changes, types.FileChange{
				FilePath:    file,
				ChangeType:  "modified",
				DiffContent: diff,
			})
		}
	}
	return changes, nil
}

// AnalyzeStagedChanges 分析已暂存的改动
func (a *Analyzer) AnalyzeStagedChanges() ([]types.FileChange, error) {
	return a.gitClient.GetStagedChanges()
}

// AnalyzeCommit 分析指定提交的改动
func (a *Analyzer) AnalyzeCommit(commitHash string) ([]types.FileChange, error) {
	return a.gitClient.GetCommitChanges(commitHash)
}

// AnalyzeWorkingDirChanges 分析工作区的改动
func (a *Analyzer) AnalyzeWorkingDirChanges() ([]types.FileChange, error) {
	return a.gitClient.GetWorkingDirChanges()
}
