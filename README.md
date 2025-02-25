# AI-CR-Tool (Code Review Tool)

一个强大的代码评审工具，支持多种AI模型（包括DeepSeek、OpenAI、ChatGLM和QWEN），通过智能分析Git差异内容，自动生成高质量的代码评审报告。

## ✨ 特性

- 🤖 支持多种AI模型
  - DeepSeek AI
  - OpenAI GPT
  - ChatGLM
  - QWEN
- 🔄 自动获取Git差异内容
- 📊 生成详细的评审报告
  - Markdown格式
  - HTML格式
- 🛠️ 简单易用的CLI界面
- ⚡ 高性能的缓存系统
- 🔌 灵活的Git Hooks集成

## 🚀 安装

```bash
# 使用 go install 安装
go install github.com/icatw/ai-cr-tool/cmd/cr@latest

# 或者从源码安装
git clone https://github.com/icatw/ai-cr-tool.git
cd ai-cr-tool
go install ./cmd/cr
```

## 🔧 配置

### 环境变量配置

你可以通过环境变量配置不同AI模型的API密钥：

```bash
# 配置不同模型的API密钥
export DEEPSEEK_API_KEY=your_deepseek_api_key
export OPENAI_API_KEY=your_openai_api_key
export CHATGLM_API_KEY=your_chatglm_api_key
export QWEN_API_KEY=your_qwen_api_key
```


## 📖 使用指南

### 基本使用

```bash
# 评审最新的代码改动
cr diff

# 评审指定的文件
cr diff --files=main.go,utils.go

# 评审指定范围的提交
cr review --commit-range=HEAD~3..HEAD

# 使用指定的AI模型
cr diff --model=qwen
```

### Git Hooks集成

在项目根目录下执行以下命令安装Git hooks：

```bash
cr install-hooks
```

这将自动安装pre-commit和pre-push钩子，在代码提交和推送时自动进行代码评审。

## 🤝 贡献

欢迎提交问题和改进建议！如果你想贡献代码，请：

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交你的改动 (`git commit -m 'feat: add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启一个 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情
