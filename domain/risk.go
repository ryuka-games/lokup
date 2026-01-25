package domain

// RiskType はリスクの種類を表す。
type RiskType string

const (
	// RiskTypeChangeConcentration は変更集中リスク。
	// 同じファイルが短期間に何度も変更されている。
	RiskTypeChangeConcentration RiskType = "change_concentration"

	// RiskTypeLargeFile は巨大ファイル。
	// ファイルの行数が閾値を超えている。
	RiskTypeLargeFile RiskType = "large_file"

	// RiskTypeAbandoned は放置ファイル。
	// 長期間変更されていないコード。
	RiskTypeAbandoned RiskType = "abandoned"

	// RiskTypeOwnership は属人化。
	// 特定の人しか触っていないファイル。
	RiskTypeOwnership RiskType = "ownership"

	// RiskTypeOutdatedDeps は依存の古さ。
	// 依存パッケージのバージョンが古い。
	RiskTypeOutdatedDeps RiskType = "outdated_deps"

	// RiskTypeLateNight は深夜労働。
	// 深夜のコミットが多い。
	RiskTypeLateNight RiskType = "late_night"
)

// DisplayName はリスク種別の表示名を返す。
func (r RiskType) DisplayName() string {
	names := map[RiskType]string{
		RiskTypeChangeConcentration: "変更集中リスク",
		RiskTypeLargeFile:           "巨大ファイル",
		RiskTypeAbandoned:           "放置ファイル",
		RiskTypeOwnership:           "属人化",
		RiskTypeOutdatedDeps:        "依存の古さ",
		RiskTypeLateNight:           "深夜労働",
	}
	if name, ok := names[r]; ok {
		return name
	}
	return string(r)
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
