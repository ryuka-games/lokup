// Package report は HTML レポート生成機能を提供する。
package report

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/ryuka-games/lokup/domain"
)

// templateFuncs はテンプレートで使用する関数。
var templateFuncs = template.FuncMap{
	"lower": strings.ToLower,
	"ge": func(a, b float64) bool {
		return a >= b
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
	Repository        string
	PeriodFrom        string
	PeriodTo          string
	PeriodDays        int
	EfficiencyScore   int
	EfficiencyGrade   string
	HealthScore       int
	HealthGrade       string
	TotalCommits      int
	FeatureAddition   float64
	Contributors      int
	LateNightRate     float64
	Risks             []RiskData
	HasRisks          bool
	CommitsByDay      []int    // 日別コミット数（グラフ用）
	CommitDayLabels   []string // 日付ラベル（グラフ用）
	GeneratedAt       string
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

	return TemplateData{
		Repository:       r.Repository.FullName(),
		PeriodFrom:       r.Period.From.Format("2006-01-02"),
		PeriodTo:         r.Period.To.Format("2006-01-02"),
		PeriodDays:       r.Period.Days(),
		EfficiencyScore:  r.EfficiencyScore.Value,
		EfficiencyGrade:  r.EfficiencyScore.Grade(),
		HealthScore:      r.HealthScore.Value,
		HealthGrade:      r.HealthScore.Grade(),
		TotalCommits:     r.Metrics.TotalCommits,
		FeatureAddition:  r.Metrics.FeatureAdditionRate,
		Contributors:     r.Metrics.TotalContributors,
		LateNightRate:    r.Metrics.LateNightCommitRate,
		Risks:            risks,
		HasRisks:         len(risks) > 0,
		GeneratedAt:      r.GeneratedAt.Format("2006-01-02 15:04:05"),
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
