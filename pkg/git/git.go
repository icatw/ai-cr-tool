package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/icatw/ai-cr-tool/pkg/types"
)

// GitClient 提供Git操作的封装
type GitClient struct {
	repoPath string
}

// NewGitClient 创建新的Git客户端
func NewGitClient(repoPath string) *GitClient {
	return &GitClient{repoPath: repoPath}
}

// GetDiff 获取指定范围的代码差异
func (g *GitClient) GetDiff(from, to string) (string, error) {
	args := []string{"diff", "--unified=3"}

	// 如果提供了范围，则使用范围比较
	if from != "" && to != "" {
		args = append(args, fmt.Sprintf("%s..%s", from, to))
	} else if from != "" {
		// 与指定提交比较
		args = append(args, from)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff failed: %v\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetChangedFiles 获取改动的文件列表
func (g *GitClient) GetChangedFiles(from, to string) ([]string, error) {
	args := []string{"diff", "--name-only"}

	if from != "" && to != "" {
		args = append(args, fmt.Sprintf("%s..%s", from, to))
	} else if from != "" {
		args = append(args, from)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff --name-only failed: %v\n%s", err, stderr.String())
	}

	files := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// GetFileContent 获取指定提交中的文件内容
func (g *GitClient) GetFileContent(filePath string, commitHash string) (string, error) {
	args := []string{"show", fmt.Sprintf("%s:%s", commitHash, filePath)}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("获取文件内容失败: %v\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetFileDiff 获取指定文件的改动内容
func (c *GitClient) GetFileDiff(file string) (string, error) {
	cmd := exec.Command("git", "diff", "HEAD", "--", file)
	cmd.Dir = c.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetStagedChanges 获取已暂存的改动
func (c *GitClient) GetStagedChanges() ([]types.FileChange, error) {
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = c.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return c.parseDiff(string(output))
}

// GetCommitChanges 获取指定提交的改动
func (c *GitClient) GetCommitChanges(commitHash string) ([]types.FileChange, error) {
	cmd := exec.Command("git", "diff", commitHash+"^", commitHash)
	cmd.Dir = c.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return c.parseDiff(string(output))
}

// GetWorkingDirChanges 获取工作区的改动
func (c *GitClient) GetWorkingDirChanges() ([]types.FileChange, error) {
	cmd := exec.Command("git", "diff")
	cmd.Dir = c.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return c.parseDiff(string(output))
}

// parseDiff 解析git diff输出
func (c *GitClient) parseDiff(diffOutput string) ([]types.FileChange, error) {
	if diffOutput == "" {
		return []types.FileChange{}, nil
	}

	var changes []types.FileChange
	// 按文件分割diff输出
	diffFiles := strings.Split(diffOutput, "diff --git")

	for _, diffFile := range diffFiles[1:] { // 跳过第一个空字符串
		lines := strings.Split(diffFile, "\n")
		if len(lines) < 2 {
			continue
		}

		// 解析文件路径
		// diff --git a/file.go b/file.go
		pathLine := lines[0]
		parts := strings.Fields(pathLine)
		if len(parts) < 4 {
			continue
		}
		filePath := strings.TrimPrefix(parts[3], "b/")

		// 确定改动类型
		changeType := "modified"
		if strings.Contains(diffFile, "new file mode") {
			changeType = "added"
		} else if strings.Contains(diffFile, "deleted file mode") {
			changeType = "deleted"
		}

		change := types.FileChange{
			FilePath:    filePath,
			ChangeType:  changeType,
			DiffContent: "diff --git" + diffFile,
		}

		changes = append(changes, change)
	}

	return changes, nil
}
