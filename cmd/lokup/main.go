// Package main は Lokup CLI のエントリーポイント。
//
// Lokup は GitHub リポジトリの健康診断ツール。
// 経営者向けサマリーと技術者向け詳細の2層構造でレポートを出力する。
//
// 使用例:
//
//	lokup facebook/react
//	lokup facebook/react --output report.html
//	lokup facebook/react --days 30
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ryuka-games/lokup/domain"
	"github.com/ryuka-games/lokup/features/analyze"
	"github.com/ryuka-games/lokup/features/report"
	"github.com/ryuka-games/lokup/infrastructure/github"
)

// Config は CLI 引数から解析された設定。
type Config struct {
	Owner  string // リポジトリオーナー（例: facebook）
	Repo   string // リポジトリ名（例: react）
	Output string // 出力ファイルパス
	Days   int    // 分析期間（日数）
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	// GitHub トークン取得（GITHUB_TOKEN → gh auth token → エラー）
	token, err := resolveGitHubToken()
	if err != nil {
		return err
	}

	fmt.Printf("Lokup - GitHub Repository Health Check\n\n")
	fmt.Printf("Repository: %s/%s\n", config.Owner, config.Repo)
	fmt.Printf("Period:     %d days\n", config.Days)
	fmt.Printf("Output:     %s\n", config.Output)
	fmt.Println()

	// 依存関係の組み立て
	client := github.NewClient(token)
	service := analyze.NewService(client)

	// 分析期間の計算
	now := time.Now()
	from := now.AddDate(0, 0, -config.Days)
	period := domain.NewDateRange(from, now)

	// 分析実行
	ctx := context.Background()
	input := analyze.ServiceInput{
		Repository: domain.NewRepository(config.Owner, config.Repo),
		Period:     period,
	}

	fmt.Println("Analyzing...")
	result, err := service.Analyze(ctx, input)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// 結果表示
	printResult(result)

	// HTML レポート生成
	fmt.Printf("\nGenerating report: %s\n", config.Output)
	reportService := report.NewService()
	if err := reportService.Generate(result, config.Output); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}
	fmt.Println("Report generated successfully!")

	return nil
}

// printResult は分析結果を表示する。
func printResult(r *domain.AnalysisResult) {
	fmt.Println("\n========================================")
	fmt.Println("           Analysis Result")
	fmt.Println("========================================")

	fmt.Printf("\nRepository: %s\n", r.Repository.FullName())
	fmt.Printf("Period:     %s ~ %s (%d days)\n",
		r.Period.From.Format("2006-01-02"),
		r.Period.To.Format("2006-01-02"),
		r.Period.Days())

	fmt.Println("\n--- Category Scores ---")
	catNames := map[domain.Category]string{
		domain.CategoryVelocity: "Velocity",
		domain.CategoryQuality:  "Quality",
		domain.CategoryTechDebt: "Tech Debt",
		domain.CategoryHealth:   "Health",
	}
	for _, cat := range []domain.Category{domain.CategoryVelocity, domain.CategoryQuality, domain.CategoryTechDebt, domain.CategoryHealth} {
		if cs, ok := r.CategoryScores[cat]; ok {
			fmt.Printf("%-12s %d/100 (%s) - %s\n", catNames[cat]+":", cs.Score.Value, cs.Score.Grade(), cs.Diagnosis)
		}
	}

	fmt.Println("\n--- Metrics ---")
	fmt.Printf("Total Commits:        %d\n", r.Metrics.TotalCommits)
	fmt.Printf("Feature Addition:     %.2f commits/day\n", r.Metrics.FeatureAdditionRate)
	fmt.Printf("Contributors:         %d\n", r.Metrics.TotalContributors)
	fmt.Printf("Late Night Commits:   %.1f%%\n", r.Metrics.LateNightCommitRate)

	fmt.Println("\n--- DORA Metrics ---")
	fmt.Printf("Deploy Freq:          %.1f/month (%s)\n", r.Metrics.DeployFrequency, r.Metrics.DeployFreqRating)
	fmt.Printf("Change Failure Rate:  %.1f%% (%s)\n", r.Metrics.ChangeFailureRate, r.Metrics.ChangeFailRating)
	fmt.Printf("MTTR:                 %.1fh (%s)\n", r.Metrics.MTTR, r.Metrics.MTTRRating)

	fmt.Println("\n--- Investment Ratio ---")
	fmt.Printf("Feature:   %d PRs (%.1f%%)\n", r.Metrics.FeaturePRCount, r.Metrics.FeatureRatio)
	fmt.Printf("BugFix:    %d PRs (%.1f%%)\n", r.Metrics.BugFixPRCount, r.Metrics.BugFixRatio)
	fmt.Printf("Refactor:  %d PRs (%.1f%%)\n", r.Metrics.RefactorPRCount, r.Metrics.RefactorRatio)
	fmt.Printf("Other:     %d PRs\n", r.Metrics.OtherPRCount)
	fmt.Printf("Revert:    %d commits (%.1f%%)\n", r.Metrics.RevertCommitCount, r.Metrics.RevertRate)

	if len(r.Trends) > 0 {
		fmt.Println("\n--- Trends (vs Previous Period) ---")
		for _, t := range r.Trends {
			arrow := "→"
			if t.Direction == "up" {
				arrow = "↑"
			} else if t.Direction == "down" {
				arrow = "↓"
			}
			fmt.Printf("%s %-16s %+.1f%%\n", arrow, t.MetricName, t.DeltaPct)
		}
	}

	if len(r.Risks) > 0 {
		fmt.Println("\n--- Risks ---")
		for _, risk := range r.Risks {
			severity := "⚪"
			switch risk.Severity {
			case domain.SeverityHigh:
				severity = "🔴"
			case domain.SeverityMedium:
				severity = "🟡"
			case domain.SeverityLow:
				severity = "🟢"
			}
			fmt.Printf("%s %s: %s\n", severity, risk.Type, risk.Description)
		}
	} else {
		fmt.Println("\n--- Risks ---")
		fmt.Println("No significant risks detected.")
	}

	fmt.Println("\n========================================")
}

// parseArgs は CLI 引数を解析して Config を返す。
func parseArgs(args []string) (*Config, error) {
	fs := flag.NewFlagSet("lokup", flag.ContinueOnError)

	// フラグ定義
	output := fs.String("output", "report.html", "Output file path")
	days := fs.Int("days", 30, "Analysis period in days")

	// カスタム Usage
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: lokup <owner/repo> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  owner/repo    GitHub repository (e.g., facebook/react)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  lokup facebook/react\n")
		fmt.Fprintf(os.Stderr, "  lokup facebook/react --output report.html\n")
		fmt.Fprintf(os.Stderr, "  lokup facebook/react --days 90\n")
	}

	// 引数解析
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// 位置引数（owner/repo）の取得
	if fs.NArg() < 1 {
		fs.Usage()
		return nil, errors.New("repository argument required")
	}

	repoArg := fs.Arg(0)
	owner, repo, err := parseRepository(repoArg)
	if err != nil {
		return nil, err
	}

	return &Config{
		Owner:  owner,
		Repo:   repo,
		Output: *output,
		Days:   *days,
	}, nil
}

// parseRepository は "owner/repo" 形式の文字列を分解する。
func parseRepository(s string) (owner, repo string, err error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %q (expected owner/repo)", s)
	}

	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])

	if owner == "" {
		return "", "", errors.New("owner cannot be empty")
	}
	if repo == "" {
		return "", "", errors.New("repo cannot be empty")
	}

	return owner, repo, nil
}

// resolveGitHubToken は GitHub トークンを取得する。
// 優先順位: GITHUB_TOKEN 環境変数 → gh auth token コマンド → エラー
func resolveGitHubToken() (string, error) {
	// 1. 環境変数
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// 2. gh auth token
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		token := strings.TrimSpace(string(out))
		if token != "" {
			return token, nil
		}
	}

	// 3. 認証なし → エラー
	return "", errors.New("GitHub authentication required.\n\n  Option 1: gh auth login\n  Option 2: export GITHUB_TOKEN=ghp_xxxxx...")
}
