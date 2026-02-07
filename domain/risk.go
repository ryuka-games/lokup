package domain

// Category はメトリクスのカテゴリを表す。
type Category string

const (
	// CategoryVelocity は開発速度カテゴリ。
	CategoryVelocity Category = "velocity"
	// CategoryQuality はコード品質カテゴリ。
	CategoryQuality Category = "quality"
	// CategoryTechDebt は技術的負債カテゴリ。
	CategoryTechDebt Category = "tech_debt"
	// CategoryHealth はチーム健全性カテゴリ。
	CategoryHealth Category = "health"
)

// RiskType はリスクの種類を表す。
type RiskType string

const (
	// RiskTypeChangeConcentration は変更集中リスク。
	RiskTypeChangeConcentration RiskType = "change_concentration"

	// RiskTypeLargeFile は巨大ファイル。
	RiskTypeLargeFile RiskType = "large_file"

	// RiskTypeOwnership は属人化。
	RiskTypeOwnership RiskType = "ownership"

	// RiskTypeOutdatedDeps は依存の古さ。
	RiskTypeOutdatedDeps RiskType = "outdated_deps"

	// RiskTypeLateNight は深夜労働。
	RiskTypeLateNight RiskType = "late_night"

	// RiskTypeSlowLeadTime はPRリードタイムが長い。
	RiskTypeSlowLeadTime RiskType = "slow_lead_time"

	// RiskTypeSlowReview はレビュー待ち時間が長い。
	RiskTypeSlowReview RiskType = "slow_review"

	// RiskTypeLargePR はPRサイズが大きい。
	RiskTypeLargePR RiskType = "large_pr"

	// RiskTypeLowIssueClose はIssueクローズ率が低い。
	RiskTypeLowIssueClose RiskType = "low_issue_close"

	// RiskTypeBugFixHigh はバグ修正割合が高い。
	RiskTypeBugFixHigh RiskType = "bug_fix_high"

	// RiskTypeLowDeployFreq はデプロイ頻度が低い。
	RiskTypeLowDeployFreq RiskType = "low_deploy_freq"

	// RiskTypeHighChangeFailure は変更失敗率が高い。
	RiskTypeHighChangeFailure RiskType = "high_change_failure"

	// RiskTypeSlowRecovery は復旧時間が長い。
	RiskTypeSlowRecovery RiskType = "slow_recovery"

	// RiskTypeLowFeatureInvestment は機能投資比率が低い。
	RiskTypeLowFeatureInvestment RiskType = "low_feature_investment"
)

// DisplayName はリスク種別の表示名を返す。
func (r RiskType) DisplayName() string {
	names := map[RiskType]string{
		RiskTypeChangeConcentration:  "変更集中リスク",
		RiskTypeLargeFile:            "巨大ファイル",
		RiskTypeOwnership:            "属人化",
		RiskTypeOutdatedDeps:         "依存の古さ",
		RiskTypeLateNight:            "深夜労働",
		RiskTypeSlowLeadTime:         "PRリードタイム超過",
		RiskTypeSlowReview:           "レビュー待ち超過",
		RiskTypeLargePR:              "PRサイズ超過",
		RiskTypeLowIssueClose:        "Issueクローズ率低下",
		RiskTypeBugFixHigh:           "バグ修正割合過多",
		RiskTypeLowDeployFreq:        "デプロイ頻度不足",
		RiskTypeHighChangeFailure:    "変更失敗率過多",
		RiskTypeSlowRecovery:         "復旧時間超過",
		RiskTypeLowFeatureInvestment: "機能投資不足",
	}
	if name, ok := names[r]; ok {
		return name
	}
	return string(r)
}

// Category はリスクタイプが属するカテゴリを返す。
func (r RiskType) Category() Category {
	switch r {
	case RiskTypeSlowLeadTime, RiskTypeSlowReview, RiskTypeLowDeployFreq, RiskTypeSlowRecovery:
		return CategoryVelocity
	case RiskTypeChangeConcentration, RiskTypeLargePR, RiskTypeLowIssueClose, RiskTypeBugFixHigh, RiskTypeHighChangeFailure:
		return CategoryQuality
	case RiskTypeLargeFile, RiskTypeOutdatedDeps, RiskTypeLowFeatureInvestment:
		return CategoryTechDebt
	case RiskTypeLateNight, RiskTypeOwnership:
		return CategoryHealth
	default:
		return CategoryQuality
	}
}

// Severity はリスクの重大度を表す。
type Severity int

const (
	// SeverityLow は低リスク（🟢）。
	SeverityLow Severity = iota
	// SeverityMedium は中リスク（🟡）。
	SeverityMedium
	// SeverityHigh は高リスク（🔴）。
	SeverityHigh
)

// Emoji は重大度を絵文字で返す。
func (s Severity) Emoji() string {
	switch s {
	case SeverityLow:
		return "🟢"
	case SeverityMedium:
		return "🟡"
	case SeverityHigh:
		return "🔴"
	default:
		return "⚪"
	}
}

// String は重大度を文字列で返す。
func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "低"
	case SeverityMedium:
		return "中"
	case SeverityHigh:
		return "高"
	default:
		return "不明"
	}
}

// Risk は検出されたリスクを表すエンティティ。
type Risk struct {
	Type        RiskType // リスクの種類
	Severity    Severity // 重大度
	Target      string   // 対象（ファイル名等）
	Description string   // 説明
	Value       int      // 数値（変更回数、行数等）
	Threshold   int      // 閾値
}

// NewRisk は Risk を生成する。
func NewRisk(riskType RiskType, severity Severity, target string, value, threshold int) Risk {
	return Risk{
		Type:      riskType,
		Severity:  severity,
		Target:    target,
		Value:     value,
		Threshold: threshold,
	}
}
