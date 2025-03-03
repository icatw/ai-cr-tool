package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/icatw/ai-cr-tool/pkg/cache"
	"github.com/icatw/ai-cr-tool/pkg/cli"
	"github.com/icatw/ai-cr-tool/pkg/git"
	"github.com/icatw/ai-cr-tool/pkg/model"
	"github.com/icatw/ai-cr-tool/pkg/review"
)

func main() {
	// 解析命令行参数
	opts, err := cli.ParseFlags()
	if err != nil {
		log.Fatalf("解析参数失败: %v\n", err)
	}

	// 初始化Git客户端
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前工作目录失败: %v\n", err)
	}
	gitClient := git.NewGitClient(wd)
	if gitClient == nil {
		log.Fatalf("初始化Git客户端失败\n")
	}

	// 初始化代码分析器
	analyzer := review.NewAnalyzer(gitClient)

	// 启动进度显示
	go cli.displayProgress(analyzer.Progress)

	// 获取代码改动
	var changes []review.FileChange
	if opts.CommitRange != "" {
		changes, err = analyzer.AnalyzeChanges(opts.CommitRange, "")
	} else {
		// 默认分析最新的改动
		changes, err = analyzer.AnalyzeChanges("HEAD~1", "HEAD")
	}
	if err != nil {
		log.Fatalf("分析代码改动失败: %v\n", err)
	}

	// 初始化缓存
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cr", "cache")
	reviewCache, err := cache.NewReviewCache(cacheDir)
	if err != nil {
		log.Printf("初始化缓存失败: %v\n", err)
	}

	// 初始化AI模型客户端
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	qwenKey := os.Getenv("QWEN_API_KEY")
	// 创建模型配置，只使用默认模型
	modelCfg := model.NewModelConfigWithKeys(deepseekKey, "", "", qwenKey)

	// 创建模型管理器
	modelManager, err := model.NewModelManager(modelCfg)
	if err != nil {
		log.Fatalf("初始化模型管理器失败: %v\n", err)
	}

	// 获取指定或默认的模型客户端
	modelClient, err := modelManager.GetClient(opts.Model)
	if err != nil {
		log.Fatalf("获取模型客户端失败: %v\n", err)
	}

	// 创建评审提示模板
	prompt := model.DefaultReviewPrompt()

	// 创建评审报告生成器
	reporter := review.NewReporter("ai-cr-tool", "HEAD")
	var issues []review.Issue

	// 处理每个改动文件
	for _, change := range changes {
		// 检查缓存
		if reviewCache != nil {
			if cached, err := reviewCache.Get(change.DiffContent); err == nil && cached != nil {
				fmt.Printf("使用缓存的评审结果 - %s\n", change.FilePath)
				issues = append(issues, review.Issue{
					Title:       "AI代码评审结果",
					FilePath:    change.FilePath,
					Severity:    review.SeverityInfo,
					Description: cached.ReviewResult,
					Suggestion:  "请根据AI评审建议进行相应修改",
				})
				continue
			}
		}

		// 生成评审提示
		messages := prompt.GeneratePrompt(change.FilePath, change.ChangeType, change.DiffContent)

		// 调用AI进行评审
		req := &model.ChatRequest{
			Model:       modelCfg.Models[modelCfg.DefaultModel].Model,
			Messages:    messages,
			MaxTokens:   modelCfg.Models[modelCfg.DefaultModel].MaxTokens,
			Temperature: modelCfg.Models[modelCfg.DefaultModel].Temperature,
		}

		resp, err := modelClient.Chat(req)
		if err != nil {
			log.Printf("评审失败 - %s: %v\n", change.FilePath, err)
			continue
		}

		// 添加评审结果到issues
		issues = append(issues, review.Issue{
			Title:       "AI代码评审结果",
			FilePath:    change.FilePath,
			Severity:    review.SeverityInfo,
			Description: resp.Choices[0].Message.Content,
			Suggestion:  "请根据AI评审建议进行相应修改",
		})

		// 缓存评审结果
		if reviewCache != nil {
			expireAfter := 24 * time.Hour
			if err := reviewCache.Set(change.DiffContent, resp.Choices[0].Message.Content, &expireAfter); err != nil {
				log.Printf("缓存评审结果失败: %v\n", err)
			}
		}
	}

	// 生成评审报告
	reportContent, err := reporter.Generate(issues, review.ReportFormat(opts.OutputFormat))
	if err != nil {
		log.Fatalf("生成评审报告失败: %v\n", err)
	}

	// 创建输出目录
	outputDir := filepath.Join(wd, "cr-result")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v\n", err)
	}

	// 生成输出文件名
	timestamp := time.Now().Format("20060102_150405")
	outputFileName := fmt.Sprintf("review_%s.%s", timestamp, opts.OutputFormat)
	outputPath := filepath.Join(outputDir, outputFileName)

	// 如果指定了输出文件，使用指定的路径
	if opts.OutputFile != "" {
		outputPath = opts.OutputFile
	}

	// 保存评审报告到文件
	if err := os.WriteFile(outputPath, reportContent, 0644); err != nil {
		log.Fatalf("保存评审报告失败: %v\n", err)
	}
	fmt.Printf("评审报告已保存到: %s\n", outputPath)

	// 同时输出到控制台
	// if opts.OutputFile == "" {
	// 	fmt.Println("\n评审报告内容:")
	// 	fmt.Println(reportContent)
	// }
}
