package hooks

import (
	"fmt"
	"os"
	"path/filepath"
)

const preCommitScript = `#!/bin/sh
# 运行代码评审
cr diff --staged --quiet

# 检查评审结果
if [ $? -ne 0 ]; then
    echo "代码评审未通过，请修复问题后再提交"
    exit 1
fi
`

func InstallPreCommitHook(gitDir string) error {
	hookPath := filepath.Join(gitDir, ".git", "hooks", "pre-commit")

	// 写入 hook 脚本
	if err := os.WriteFile(hookPath, []byte(preCommitScript), 0755); err != nil {
		return fmt.Errorf("安装 pre-commit hook 失败: %v", err)
	}

	return nil
}
