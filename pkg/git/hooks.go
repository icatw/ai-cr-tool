package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HookType 定义Git钩子类型
type HookType string

const (
	PreCommitHook HookType = "pre-commit"
	PrePushHook   HookType = "pre-push"
)

// HookConfig 钩子配置
type HookConfig struct {
	Enabled bool
	Options map[string]string
}

// HookManager Git钩子管理器
type HookManager struct {
	repoPath string
	config   map[HookType]HookConfig
}

// NewHookManager 创建新的钩子管理器
func NewHookManager(repoPath string) *HookManager {
	return &HookManager{
		repoPath: repoPath,
		config:   make(map[HookType]HookConfig),
	}
}

// ConfigureHook 配置Git钩子
func (m *HookManager) ConfigureHook(hookType HookType, config HookConfig) {
	m.config[hookType] = config
}

// InstallHook 安装Git钩子
func (m *HookManager) InstallHook(hookType HookType) error {
	hookPath := filepath.Join(m.repoPath, ".git", "hooks", string(hookType))

	// 检查钩子配置
	config, ok := m.config[hookType]
	if !ok {
		config = HookConfig{Enabled: true, Options: make(map[string]string)}
	}

	// 如果钩子未启用，则跳过安装
	if !config.Enabled {
		return nil
	}

	// 检查钩子文件是否已存在
	if _, err := os.Stat(hookPath); err == nil {
		// 备份已存在的钩子
		backupPath := hookPath + ".backup"
		if err := os.Rename(hookPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing hook: %v", err)
		}
	}

	// 创建钩子脚本内容
	content := m.generateHookScript(hookType)

	// 写入钩子文件
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write hook file: %v", err)
	}

	return nil
}

// RemoveHook 移除Git钩子
func (m *HookManager) RemoveHook(hookType HookType) error {
	hookPath := filepath.Join(m.repoPath, ".git", "hooks", string(hookType))

	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove hook: %v", err)
	}

	return nil
}

// generateHookScript 生成钩子脚本内容
func (m *HookManager) generateHookScript(hookType HookType) string {
	var script strings.Builder

	// 添加脚本头部
	script.WriteString("#!/bin/sh\n\n")

	// 添加错误处理
	script.WriteString("set -e\n\n")

	// 添加日志函数
	script.WriteString("log() {\n")
	script.WriteString("    echo \"[ai-cr-tool] $1\"\n")
	script.WriteString("}\n\n")

	// 根据钩子类型生成不同的脚本内容
	switch hookType {
	case PreCommitHook:
		script.WriteString("# 运行代码评审工具\n")
		script.WriteString("ai-cr-tool review pre-commit || exit 1\n")
	case PrePushHook:
		script.WriteString("# 运行代码评审工具\n")
		script.WriteString("ai-cr-tool review pre-push || exit 1\n")
	}

	return script.String()
}
