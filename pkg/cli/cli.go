package cli

import (
	"flag"
	"fmt"
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
}

// ParseFlags 解析命令行参数
func ParseFlags() (*Options, error) {
	opts := &Options{}

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

	// 验证参数
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

// validateOptions 验证命令行参数
func validateOptions(opts *Options) error {
	// 检查评审范围参数
	if opts.Files == "" && opts.CommitRange == "" {
		// 如果未指定任何参数，默认使用HEAD~1..HEAD
		opts.CommitRange = "HEAD~1..HEAD"
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
