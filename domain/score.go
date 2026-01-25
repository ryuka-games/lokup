package domain

// Score は0-100の範囲のスコアを表す値オブジェクト。
type Score struct {
	Value     int
	Breakdown []ScoreBreakdownItem // スコアの内訳
}

// ScoreBreakdownItem はスコア内訳の1項目。
type ScoreBreakdownItem struct {
	Label  string // 項目名（例: "基本スコア", "深夜労働リスク"）
	Points int    // 点数（正: 加点、負: 減点）
	Detail string // 詳細（例: "32% / 基準30%"）
}

// NewScore は Score を生成する。
// 値は0-100の範囲にクランプされる。
func NewScore(value int) Score {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return Score{Value: value}
}

// NewScoreWithBreakdown は内訳付きの Score を生成する。
func NewScoreWithBreakdown(value int, breakdown []ScoreBreakdownItem) Score {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return Score{Value: value, Breakdown: breakdown}
}

// Grade はスコアをグレード（A/B/C/D）で返す。
//
//	A: 80-100（良好）
//	B: 60-79（普通）
//	C: 40-59（要改善）
//	D: 0-39（危険）
func (s Score) Grade() string {
	switch {
	case s.Value >= 80:
		return "A"
	case s.Value >= 60:
		return "B"
	case s.Value >= 40:
		return "C"
	default:
		return "D"
	}
}

// GradeDescription はグレードの説明を返す。
func (s Score) GradeDescription() string {
	switch s.Grade() {
	case "A":
		return "良好"
	case "B":
		return "普通"
	case "C":
		return "要改善"
	case "D":
		return "危険"
	default:
		return "不明"
	}
}
