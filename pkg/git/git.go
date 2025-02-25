package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
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
