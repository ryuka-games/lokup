// Package report は HTML レポート生成機能を提供する。
package report

import (
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
	"gt": func(a, b int) bool {
		return a > b
	},
	"lt": func(a, b int) bool {
		return a < b
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
	Repository            string
	PeriodFrom            string
	PeriodTo              string
	PeriodDays            int
	EfficiencyScore       int
	EfficiencyGrade       string
	EfficiencyBreakdown   []BreakdownItem
	HealthScore           int
	HealthGrade           string
	HealthBreakdown       []BreakdownItem
	TotalCommits          int
	FeatureAddition       float64
	Contributors          int
	LateNightRate         float64
	AvgLeadTime           float64 // PRリードタイム（日）
	BugFixRatio           float64 // バグ修正割合（%）
	FeaturePRCount        int     // feature PRの件数
	BugFixPRCount         int     // bugfix PRの件数
	OtherPRCount          int     // その他PRの件数
	Risks                 []RiskData
	HasRisks              bool
	CommitsByDay          []int    // 日別コミット数（グラフ用）
	CommitDayLabels       []string // 日付ラベル（グラフ用）
	GeneratedAt           string
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

// prepareTemplateData は分析結果からテンプレートデータを準備する。
func (s *Service) prepareTemplateData(r *domain.AnalysisResult) TemplateData {
	risks := make([]RiskData, len(r.Risks))
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
		risks[i] = RiskData{
			Severity:     severity,
			SeverityIcon: icon,
			Type:         riskTypeToJapanese(risk.Type),
			Description:  risk.Description,
			Target:       risk.Target,
			Action:       riskTypeToAction(risk.Type),
		}
	}

	// スコア内訳を変換
	efficiencyBreakdown := make([]BreakdownItem, len(r.EfficiencyScore.Breakdown))
	for i, b := range r.EfficiencyScore.Breakdown {
		efficiencyBreakdown[i] = BreakdownItem{Label: b.Label, Points: b.Points, Detail: b.Detail}
	}
	healthBreakdown := make([]BreakdownItem, len(r.HealthScore.Breakdown))
	for i, b := range r.HealthScore.Breakdown {
		healthBreakdown[i] = BreakdownItem{Label: b.Label, Points: b.Points, Detail: b.Detail}
	}

	// 日別コミットデータをグラフ用に変換
	commitsByDay := make([]int, len(r.DailyCommits))
	commitDayLabels := make([]string, len(r.DailyCommits))
	for i, dc := range r.DailyCommits {
		commitsByDay[i] = dc.Count
		commitDayLabels[i] = formatDateWithWeekday(dc.Date) // "1/25(土)" 形式
	}

	return TemplateData{
		Repository:          r.Repository.FullName(),
		PeriodFrom:          r.Period.From.Format("2006-01-02"),
		PeriodTo:            r.Period.To.Format("2006-01-02"),
		PeriodDays:          r.Period.Days(),
		EfficiencyScore:     r.EfficiencyScore.Value,
		EfficiencyGrade:     r.EfficiencyScore.Grade(),
		EfficiencyBreakdown: efficiencyBreakdown,
		HealthScore:         r.HealthScore.Value,
		HealthGrade:         r.HealthScore.Grade(),
		HealthBreakdown:     healthBreakdown,
		TotalCommits:        r.Metrics.TotalCommits,
		FeatureAddition:     r.Metrics.FeatureAdditionRate,
		Contributors:        r.Metrics.TotalContributors,
		LateNightRate:       r.Metrics.LateNightCommitRate,
		AvgLeadTime:         r.Metrics.AvgLeadTime,
		BugFixRatio:         r.Metrics.BugFixRatio,
		FeaturePRCount:      r.Metrics.FeaturePRCount,
		BugFixPRCount:       r.Metrics.BugFixPRCount,
		OtherPRCount:        r.Metrics.OtherPRCount,
		Risks:               risks,
		HasRisks:            len(risks) > 0,
		CommitsByDay:        commitsByDay,
		CommitDayLabels:     commitDayLabels,
		GeneratedAt:         r.GeneratedAt.Format("2006-01-02 15:04:05"),
	}
}

// riskTypeToJapanese はリスクタイプを日本語に変換する。
func riskTypeToJapanese(rt domain.RiskType) string {
	return rt.DisplayName()
}

// riskTypeToAction はリスクタイプに対する改善提案を返す。
func riskTypeToAction(rt domain.RiskType) string {
	actions := map[domain.RiskType]string{
		domain.RiskTypeChangeConcentration: "💡 提案: このファイルの責務を分割することを検討してください。頻繁な変更はバグの温床になります。",
		domain.RiskTypeLargeFile:           "💡 提案: ファイルを機能ごとに分割してください。大きなファイルは可読性と保守性を下げます。",
		domain.RiskTypeAbandoned:           "💡 提案: このコードが本当に必要か確認してください。不要なら削除、必要ならドキュメントを追加しましょう。",
		domain.RiskTypeOwnership:           "💡 提案: コードレビューやペアプログラミングで知識を共有してください。担当者が離脱するとリスクになります。",
		domain.RiskTypeOutdatedDeps:        "💡 提案: 依存パッケージを更新してください。古いバージョンにはセキュリティ脆弱性がある可能性があります。",
		domain.RiskTypeLateNight:           "💡 提案: 深夜作業が多い原因を調査してください。締め切り圧力やリソース不足の兆候かもしれません。",
	}
	if action, ok := actions[rt]; ok {
		return action
	}
	return "💡 提案: 詳細を確認し、改善策を検討してください。"
}

// formatDateWithWeekday は日付を "1/25(土)" 形式でフォーマットする。
func formatDateWithWeekday(t time.Time) string {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	return fmt.Sprintf("%d/%d(%s)", t.Month(), t.Day(), weekdays[t.Weekday()])
}
