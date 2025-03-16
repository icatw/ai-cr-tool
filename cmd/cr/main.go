package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icatw/ai-cr-tool/pkg/cache"
	"github.com/icatw/ai-cr-tool/pkg/cli"
	"github.com/icatw/ai-cr-tool/pkg/git"
	"github.com/icatw/ai-cr-tool/pkg/model"
	"github.com/icatw/ai-cr-tool/pkg/review"
	"github.com/icatw/ai-cr-tool/pkg/types"
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

	// 初始化代码分析器
	analyzer := review.NewAnalyzer(gitClient)

	// 获取代码改动
	var changes []types.FileChange
	switch {
	case opts.Files != "":
		// 评审指定文件
		files := strings.Split(opts.Files, ",")
		changes, err = analyzer.AnalyzeFiles(files)
	case opts.Staged:
		// 评审已暂存的改动
		changes, err = analyzer.AnalyzeStagedChanges()
	case opts.CommitHash != "":
		// 评审指定提交
		changes, err = analyzer.AnalyzeCommit(opts.CommitHash)
	case opts.CommitRange != "":
		// 评审提交范围
		changes, err = analyzer.AnalyzeChanges(opts.CommitRange, "")
	default:
		// 默认评审所有未提交的改动
		changes, err = analyzer.AnalyzeWorkingDirChanges()
	}

	if err != nil {
		log.Fatalf("分析代码改动失败: %v\n", err)
	}

	if len(changes) == 0 {
		if !opts.Quiet {
			fmt.Println("没有发现需要评审的代码改动")
		}
		return
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
	modelCfg := model.NewModelConfigWithKeys(deepseekKey, "", "", qwenKey)

	modelManager, err := model.NewModelManager(modelCfg)
	if err != nil {
		log.Fatalf("初始化模型管理器失败: %v\n", err)
	}

	modelClient, err := modelManager.GetClient(opts.Model)
	if err != nil {
		log.Fatalf("获取模型客户端失败: %v\n", err)
	}

	// 创建评审提示模板
	prompt := model.DefaultReviewPrompt()

	// 创建评审报告生成器
	reporter := review.NewReporter("ai-cr-tool", "HEAD")
	var issues []types.Issue

	// 处理每个改动文件
	for _, change := range changes {
		if !opts.Quiet {
			fmt.Printf("正在评审文件: %s\n", change.FilePath)
		}

		// 检查缓存
		if reviewCache != nil {
			if cached, err := reviewCache.Get(change.DiffContent); err == nil && cached != nil {
				issues = append(issues, types.Issue{
					Title:       "缓存的评审结果",
					FilePath:    change.FilePath,
					Description: cached.ReviewResult,
					Severity:    types.SeverityInfo,
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

		// 添加评审结果
		issues = append(issues, types.Issue{
			Title:       "AI代码评审结果",
			FilePath:    change.FilePath,
			Description: resp.Choices[0].Message.Content,
			Severity:    types.SeverityInfo,
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
	format, err := review.ParseReportFormat(opts.OutputFormat)
	if err != nil {
		log.Fatalf("不支持的输出格式: %v\n", err)
	}

	reportContent, err := reporter.Generate(issues, format)
	if err != nil {
		log.Fatalf("生成评审报告失败: %v\n", err)
	}

	// 保存报告
	if opts.OutputFile != "" {
		if err := os.WriteFile(opts.OutputFile, []byte(reportContent), 0644); err != nil {
			log.Fatalf("保存评审报告失败: %v\n", err)
		}
		fmt.Printf("评审报告已保存到: %s\n", opts.OutputFile)
	} else {
		fmt.Println("\n评审报告:")
		fmt.Println(reportContent)
	}
}
