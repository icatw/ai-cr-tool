package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/icatw/ai-cr-tool/pkg/review"
)

// Options 定义命令行参数选项
type Options struct {
	// 评审范围相关选项
	Files       string
	CommitRange string

	// 输出相关选项
	OutputFormat string
	OutputFile   string

	// AI模型选项
	Model string

	// 其他选项
	Verbose bool

	// 帮助选项
	Help bool
}

// ParseFlags 解析命令行参数
func ParseFlags() (*Options, error) {
	opts := &Options{}

	// 帮助选项
	flag.BoolVar(&opts.Help, "help", false, "显示帮助信息")
	flag.BoolVar(&opts.Help, "h", false, "显示帮助信息（简写）")

	// 评审范围选项
	flag.StringVar(&opts.Files, "files", "", "指定要评审的文件列表，多个文件用逗号分隔")
	flag.StringVar(&opts.CommitRange, "commit-range", "", "指定要评审的提交范围，例如：HEAD~1..HEAD")

	// 输出选项
	flag.StringVar(&opts.OutputFormat, "format", "markdown", "输出格式：markdown, html, pdf")
	flag.StringVar(&opts.OutputFile, "output", "", "输出文件路径，默认输出到标准输出")

	// AI模型选项
	flag.StringVar(&opts.Model, "model", "", "指定使用的AI模型，可选值：qwen, deepseek, openai, chatglm")

	// 其他选项
	flag.BoolVar(&opts.Verbose, "verbose", false, "显示详细日志信息")

	// 解析参数
	flag.Parse()

	// 如果指定了帮助选项，显示帮助信息
	if opts.Help {
		printHelp()
		return opts, nil
	}

	// 验证参数
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

// displayProgress 显示进度信息
func displayProgress(progress chan review.ProgressInfo) {
	var lastUpdate time.Time
	for info := range progress {
		// 限制更新频率，避免刷新过快
		if time.Since(lastUpdate) < 100*time.Millisecond {
			continue
		}
		// 清除当前行
		fmt.Printf("\r\033[K")
		// 显示进度信息
		fmt.Printf("正在评审: %s [%.1f%%] (%d/%d)", info.CurrentFile, info.Percentage, info.CurrentIndex, info.TotalFiles)
		lastUpdate = time.Now()
	}
	// 评审完成后换行
	fmt.Println()
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println(`AI代码评审工具

用法：
  cr [选项]

选项：
  -h, --help		显示帮助信息
  --files string	指定要评审的文件列表，多个文件用逗号分隔
  --commit-range string	指定要评审的提交范围，例如：HEAD~1..HEAD
  --format string	输出格式：markdown, html, pdf（默认 "markdown"）
  --output string	输出文件路径，默认输出到标准输出
  --model string	指定使用的AI模型，可选值：qwen, deepseek, openai, chatglm
  --verbose		显示详细日志信息

示例：
  # 评审当前工作区未提交的代码变更
  cr

  # 评审指定文件
  cr --files main.go,utils.go

  # 评审最近一次提交的变更
  cr --commit-range HEAD~1..HEAD

  # 使用HTML格式输出到文件
  cr --format html --output report.html

  # 使用指定的AI模型进行评审
  cr --model qwen
`)
}

// validateOptions 验证命令行参数
func validateOptions(opts *Options) error {
	// 检查评审范围参数
	if opts.Files == "" && opts.CommitRange == "" {
		// 如果未指定任何参数，默认评审工作区未提交的代码变更
		opts.CommitRange = "HEAD"
	}

	// 检查输出格式
	switch opts.OutputFormat {
	case "markdown", "html", "pdf":
		// 支持的格式
	default:
		return fmt.Errorf("不支持的输出格式：%s", opts.OutputFormat)
	}

	// 检查AI模型
	if opts.Model != "" {
		switch opts.Model {
		case "qwen", "deepseek", "openai", "chatglm":
			// 支持的模型
		default:
			return fmt.Errorf("不支持的AI模型：%s", opts.Model)
		}
	}

	return nil
}
