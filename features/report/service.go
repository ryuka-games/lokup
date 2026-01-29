// Package report は HTML レポート生成機能を提供する。
package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

// templateFuncs はテンプレートで使用する関数。
var templateFuncs = template.FuncMap{
	"lower": strings.ToLower,
	"ge": func(a, b float64) bool {
		return a >= b
	},
	"geInt": func(a, b int) bool {
		return a >= b
	},
	"gt": func(a, b int) bool {
		return a > b
	},
	"lt": func(a, b int) bool {
		return a < b
	},
	"ltFloat": func(a, b float64) bool {
		return a < b
	},
	"eq": func(a, b string) bool {
		return a == b
	},
}

// Service はレポート生成のビジネスロジックを担当する。
type Service struct{}

// NewService は Service を生成する。
func NewService() *Service {
	return &Service{}
}

// Generate は分析結果から HTML レポートを生成する。
func (s *Service) Generate(result *domain.AnalysisResult, outputPath string) error {
	// テンプレートデータの準備
	data := s.prepareTemplateData(result)

	// テンプレート解析
	tmpl, err := template.New("report").Funcs(templateFuncs).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// ファイル作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// テンプレート実行
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// TemplateData はテンプレートに渡すデータ。
type TemplateData struct {
	Repository string
	PeriodFrom string
	PeriodTo   string
	PeriodDays int

	// カテゴリスコア
	Categories []CategoryScoreData

	// メトリクス値
	TotalCommits      int
	FeatureAddition   float64
	Contributors      int
	LateNightRate     float64
	AvgLeadTime       float64
	AvgReviewWaitTime float64
	OpenPRCount       int
	OpenIssueCount    int
	BugFixRatio       float64
	AvgPRSize         int
	IssueCloseRate    float64
	IssuesCreated     int
	IssuesClosed      int
	FeaturePRCount    int
	BugFixPRCount     int
	OtherPRCount      int

	// DORA メトリクス
	DeployFrequency   float64
	DeployFreqRating  string
	ChangeFailureRate float64
	ChangeFailRating  string
	MTTR              float64
	MTTRRating        string

	// 投資比率
	RefactorPRCount int
	FeatureRatio    float64
	RefactorRatio   float64

	// コードチャーン
	RevertCommitCount int
	RevertRate        float64

	// チーム
	TotalFiles int

	// トレンド
	TrendsJSON template.JS

	// 技術的負債
	LargeFileCount   int
	LargeFiles       []LargeFileData
	OutdatedDepCount int
	OutdatedDeps     []OutdatedDepData

	// リスク
	Risks    []RiskData
	HasRisks bool

	// 変更集中リスク一覧（ドリルダウンテーブル用）
	ChangeConcentrationRisks []RiskData

	// グラフ用データ
	CommitsByDay    []int
	CommitDayLabels []string

	// ドリルダウン用JSON（template.JS で安全にスクリプトに埋め込み）
	PRDetailsJSON          template.JS
	ContributorDetailsJSON template.JS
	HourlyCommitsJSON      template.JS

	GeneratedAt string
}

// CategoryScoreData はカテゴリスコアのテンプレートデータ。
type CategoryScoreData struct {
	Icon       string // 📈, ✅, ⚠️, 💚
	Name       string // 開発速度, コード品質, etc.
	CategoryID string // velocity, quality, etc.
	Score      int
	Grade      string
	GradeClass string // grade-a, grade-b, etc.
	Diagnosis  string
	Breakdown  []BreakdownItem
}

// BreakdownItem はスコア内訳の1項目。
type BreakdownItem struct {
	Label  string
	Points int
	Detail string
}

// RiskData はリスク情報。
type RiskData struct {
	Severity     string // "high", "medium", "low"
	SeverityIcon string // 🔴, 🟡, 🟢
	Type         string
	Description  string
	Target       string
	Action       string // 改善提案
}

// PRDetailData はPR詳細のJSON用データ。
type PRDetailData struct {
	Number          int     `json:"number"`
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	LeadTimeDays    float64 `json:"leadTimeDays"`
	Size            int     `json:"size"`
	Additions       int     `json:"additions"`
	Deletions       int     `json:"deletions"`
	ReviewWaitHours float64 `json:"reviewWaitHours"`
}

// ContributorDetailData はコントリビューター詳細のJSON用データ。
type ContributorDetailData struct {
	Name    string  `json:"name"`
	Commits int     `json:"commits"`
	Ratio   float64 `json:"ratio"`
}

// LargeFileData は巨大ファイル情報。
type LargeFileData struct {
	Path        string
	SizeKB      int
	SeverityStr string
}

// OutdatedDepData は古い依存情報。
type OutdatedDepData struct {
	Name        string
	Version     string
	Age         string
	SeverityStr string
}

// prepareTemplateData は分析結果からテンプレートデータを準備する。
func (s *Service) prepareTemplateData(r *domain.AnalysisResult) TemplateData {
	// リスクデータを変換
	risks := make([]RiskData, len(r.Risks))
	var changeConcentrationRisks []RiskData
	for i, risk := range r.Risks {
		severity := "low"
		icon := "🟢"
		switch risk.Severity {
		case domain.SeverityHigh:
			severity = "high"
			icon = "🔴"
		case domain.SeverityMedium:
			severity = "medium"
			icon = "🟡"
		}
		rd := RiskData{
			Severity:     severity,
			SeverityIcon: icon,
			Type:         risk.Type.DisplayName(),
			Description:  risk.Description,
			Target:       risk.Target,
			Action:       riskTypeToAction(risk.Type),
		}
		risks[i] = rd

		if risk.Type == domain.RiskTypeChangeConcentration {
			changeConcentrationRisks = append(changeConcentrationRisks, rd)
		}
	}

	// カテゴリスコアを変換
	categories := s.buildCategoryScoreData(r.CategoryScores)

	// 日別コミットデータをグラフ用に変換
	commitsByDay := make([]int, len(r.DailyCommits))
	commitDayLabels := make([]string, len(r.DailyCommits))
	for i, dc := range r.DailyCommits {
		commitsByDay[i] = dc.Count
		commitDayLabels[i] = formatDateWithWeekday(dc.Date)
	}

	// 巨大ファイルデータを変換
	largeFiles := make([]LargeFileData, len(r.LargeFiles))
	for i, lf := range r.LargeFiles {
		severityStr := "medium"
		if lf.Severity == domain.SeverityHigh {
			severityStr = "high"
		}
		largeFiles[i] = LargeFileData{
			Path:        lf.Path,
			SizeKB:      lf.SizeKB,
			SeverityStr: severityStr,
		}
	}

	// 古い依存データを変換
	outdatedDeps := make([]OutdatedDepData, len(r.OutdatedDeps))
	for i, od := range r.OutdatedDeps {
		severityStr := "medium"
		if od.Severity == domain.SeverityHigh {
			severityStr = "high"
		}
		outdatedDeps[i] = OutdatedDepData{
			Name:        od.Name,
			Version:     od.Version,
			Age:         od.Age,
			SeverityStr: severityStr,
		}
	}

	// ドリルダウン用JSONデータ
	prDetailsJSON := s.marshalPRDetails(r.PRDetails)
	contributorDetailsJSON := s.marshalContributorDetails(r.ContributorDetails)
	hourlyCommitsJSON := s.marshalHourlyCommits(r.HourlyCommits)
	trendsJSON := s.marshalTrends(r.Trends)

	return TemplateData{
		Repository: r.Repository.FullName(),
		PeriodFrom: r.Period.From.Format("2006-01-02"),
		PeriodTo:   r.Period.To.Format("2006-01-02"),
		PeriodDays: r.Period.Days(),

		Categories: categories,

		TotalCommits:      r.Metrics.TotalCommits,
		FeatureAddition:   r.Metrics.FeatureAdditionRate,
		Contributors:      r.Metrics.TotalContributors,
		LateNightRate:     r.Metrics.LateNightCommitRate,
		AvgLeadTime:       r.Metrics.AvgLeadTime,
		AvgReviewWaitTime: r.Metrics.AvgReviewWaitTime,
		OpenPRCount:       r.Metrics.OpenPRCount,
		OpenIssueCount:    r.Metrics.OpenIssueCount,
		BugFixRatio:       r.Metrics.BugFixRatio,
		AvgPRSize:         r.Metrics.AvgPRSize,
		IssueCloseRate:    r.Metrics.IssueCloseRate,
		IssuesCreated:     r.Metrics.IssuesCreated,
		IssuesClosed:      r.Metrics.IssuesClosed,
		FeaturePRCount:    r.Metrics.FeaturePRCount,
		BugFixPRCount:     r.Metrics.BugFixPRCount,
		OtherPRCount:      r.Metrics.OtherPRCount,

		DeployFrequency:   r.Metrics.DeployFrequency,
		DeployFreqRating:  r.Metrics.DeployFreqRating,
		ChangeFailureRate: r.Metrics.ChangeFailureRate,
		ChangeFailRating:  r.Metrics.ChangeFailRating,
		MTTR:              r.Metrics.MTTR,
		MTTRRating:        r.Metrics.MTTRRating,

		RefactorPRCount: r.Metrics.RefactorPRCount,
		FeatureRatio:    r.Metrics.FeatureRatio,
		RefactorRatio:   r.Metrics.RefactorRatio,

		RevertCommitCount: r.Metrics.RevertCommitCount,
		RevertRate:        r.Metrics.RevertRate,

		TotalFiles: r.Metrics.TotalFiles,

		TrendsJSON: trendsJSON,

		LargeFileCount:   len(r.LargeFiles),
		LargeFiles:       largeFiles,
		OutdatedDepCount: len(r.OutdatedDeps),
		OutdatedDeps:     outdatedDeps,

		Risks:                    risks,
		HasRisks:                 len(risks) > 0,
		ChangeConcentrationRisks: changeConcentrationRisks,

		CommitsByDay:    commitsByDay,
		CommitDayLabels: commitDayLabels,

		PRDetailsJSON:          prDetailsJSON,
		ContributorDetailsJSON: contributorDetailsJSON,
		HourlyCommitsJSON:      hourlyCommitsJSON,

		GeneratedAt: r.GeneratedAt.Format("2006-01-02 15:04:05"),
	}
}

// buildCategoryScoreData はカテゴリスコアをテンプレートデータに変換する。
func (s *Service) buildCategoryScoreData(scores map[domain.Category]domain.CategoryScore) []CategoryScoreData {
	type catInfo struct {
		cat  domain.Category
		icon string
		name string
	}

	order := []catInfo{
		{domain.CategoryVelocity, "📈", "開発速度"},
		{domain.CategoryQuality, "✅", "コード品質"},
		{domain.CategoryTechDebt, "⚠️", "技術的負債"},
		{domain.CategoryHealth, "💚", "チーム健全性"},
	}

	var result []CategoryScoreData
	for _, ci := range order {
		cs, ok := scores[ci.cat]
		if !ok {
			cs = domain.CategoryScore{
				Category:  ci.cat,
				Score:     domain.NewScore(100),
				Diagnosis: "良好な状態です",
			}
		}

		breakdown := make([]BreakdownItem, len(cs.Score.Breakdown))
		for i, b := range cs.Score.Breakdown {
			breakdown[i] = BreakdownItem{Label: b.Label, Points: b.Points, Detail: b.Detail}
		}

		result = append(result, CategoryScoreData{
			Icon:       ci.icon,
			Name:       ci.name,
			CategoryID: string(ci.cat),
			Score:      cs.Score.Value,
			Grade:      cs.Score.Grade(),
			GradeClass: "grade-" + strings.ToLower(cs.Score.Grade()),
			Diagnosis:  cs.Diagnosis,
			Breakdown:  breakdown,
		})
	}

	return result
}

// marshalPRDetails はPR詳細をJSON文字列に変換する。
func (s *Service) marshalPRDetails(details []domain.PRDetail) template.JS {
	data := make([]PRDetailData, len(details))
	for i, d := range details {
		data[i] = PRDetailData{
			Number:          d.Number,
			Title:           d.Title,
			Author:          d.Author,
			LeadTimeDays:    d.LeadTimeDays,
			Size:            d.Size,
			Additions:       d.Additions,
			Deletions:       d.Deletions,
			ReviewWaitHours: d.ReviewWaitHours,
		}
	}
	b, _ := json.Marshal(data)
	return template.JS(b)
}

// marshalContributorDetails はコントリビューター詳細をJSON文字列に変換する。
func (s *Service) marshalContributorDetails(details []domain.ContributorDetail) template.JS {
	data := make([]ContributorDetailData, len(details))
	for i, d := range details {
		data[i] = ContributorDetailData{
			Name:    d.Name,
			Commits: d.Commits,
			Ratio:   d.Ratio,
		}
	}
	b, _ := json.Marshal(data)
	return template.JS(b)
}

// marshalHourlyCommits は時間帯別コミット数をJSON文字列に変換する。
func (s *Service) marshalHourlyCommits(hourly [24]int) template.JS {
	b, _ := json.Marshal(hourly[:])
	return template.JS(b)
}

// marshalTrends はトレンドデータをJSON文字列に変換する。
func (s *Service) marshalTrends(trends []domain.TrendDelta) template.JS {
	b, _ := json.Marshal(trends)
	return template.JS(b)
}

// riskTypeToAction はリスクタイプに対する改善提案を返す。
func riskTypeToAction(rt domain.RiskType) string {
	actions := map[domain.RiskType]string{
		domain.RiskTypeChangeConcentration: "このファイルの責務を分割することを検討してください。頻繁な変更はバグの温床になります。",
		domain.RiskTypeLargeFile:           "ファイルを機能ごとに分割してください。大きなファイルは可読性と保守性を下げます。",
		domain.RiskTypeOwnership:           "コードレビューやペアプログラミングで知識を共有してください。担当者が離脱するとリスクになります。",
		domain.RiskTypeOutdatedDeps:        "依存パッケージを更新してください。古いバージョンにはセキュリティ脆弱性がある可能性があります。",
		domain.RiskTypeLateNight:           "深夜作業が多い原因を調査してください。締め切り圧力やリソース不足の兆候かもしれません。",
		domain.RiskTypeSlowLeadTime:        "PRを小さく分割し、レビュー担当をローテーションで明確化してください。",
		domain.RiskTypeSlowReview:          "レビュー時間をカレンダーで確保し、Slackへの通知など見逃さない仕組みを導入してください。",
		domain.RiskTypeLargePR:             "1つのPRで1つの機能/修正に絞り、リファクタリングと機能追加を分けてください。",
		domain.RiskTypeLowIssueClose:       "定期的なトリアージミーティングで優先度を整理し、対応しないものは wontfix でクローズしてください。",
		domain.RiskTypeBugFixHigh:          "テストを充実させてバグを事前に防ぎ、コードレビューの品質を上げてください。",
		domain.RiskTypeLowDeployFreq:       "CI/CDパイプラインを整備し、小さなリリースを頻繁に行う文化を構築してください。",
		domain.RiskTypeHighChangeFailure:   "リリース前のテスト自動化とステージング環境での検証を強化してください。",
		domain.RiskTypeSlowRecovery:        "インシデント対応プロセスを整備し、ロールバック手順を自動化してください。",
		domain.RiskTypeLowFeatureInvestment: "技術的負債の計画的な返済とともに、機能開発への投資バランスを見直してください。",
	}
	if action, ok := actions[rt]; ok {
		return action
	}
	return "詳細を確認し、改善策を検討してください。"
}

// formatDateWithWeekday は日付を "1/25(土)" 形式でフォーマットする。
func formatDateWithWeekday(t time.Time) string {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	return fmt.Sprintf("%d/%d(%s)", t.Month(), t.Day(), weekdays[t.Weekday()])
}
