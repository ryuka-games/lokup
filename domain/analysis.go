package domain

import "time"

// DateRange は分析期間を表す値オブジェクト。
type DateRange struct {
	From time.Time
	To   time.Time
}

// NewDateRange は DateRange を生成する。
func NewDateRange(from, to time.Time) DateRange {
	return DateRange{From: from, To: to}
}

// Days は期間の日数を返す。
func (d DateRange) Days() int {
	return int(d.To.Sub(d.From).Hours() / 24)
}

// AnalysisResult は分析結果を表す集約。
// これが集約ルートであり、診断結果全体を束ねる。
type AnalysisResult struct {
	Repository      Repository    // 対象リポジトリ
	Period          DateRange     // 分析期間
	EfficiencyScore Score         // 開発効率スコア（経営向け）
	HealthScore     Score         // コード健全性スコア（技術向け）
	Risks           []Risk        // 検出されたリスク
	Metrics         Metrics       // 各種メトリクス
	DailyCommits    []DailyCommit // 日別コミット数
	GeneratedAt     time.Time     // レポート生成日時
}

// DailyCommit は1日分のコミット数を表す。
type DailyCommit struct {
	Date  time.Time
	Count int
}

// Metrics は各種メトリクスを表す。
type Metrics struct {
	// 経営向けメトリクス
	TotalCommits        int     // 総コミット数
	FeatureAdditionRate float64 // 機能追加速度（コミット/日）
	BugFixRatio         float64 // バグ修正の割合（%）
	ReworkRate          float64 // 手戻り率（%）
	AvgLeadTime         float64 // PR作成→マージの平均日数

	// PR内訳
	FeaturePRCount int // feature PRの件数
	BugFixPRCount  int // bugfix PRの件数
	OtherPRCount   int // その他PRの件数

	// 技術向けメトリクス
	TotalFiles          int     // 総ファイル数
	TotalContributors   int     // コントリビューター数
	LateNightCommitRate float64 // 深夜コミット率（%）
}

// RiskCount は重大度別のリスク数を返す。
func (a *AnalysisResult) RiskCount(severity Severity) int {
	count := 0
	for _, r := range a.Risks {
		if r.Severity == severity {
			count++
		}
	}
	return count
}

// HighRisks は高リスク（🔴）の一覧を返す。
func (a *AnalysisResult) HighRisks() []Risk {
	var risks []Risk
	for _, r := range a.Risks {
		if r.Severity == SeverityHigh {
			risks = append(risks, r)
		}
	}
	return risks
}
