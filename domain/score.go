package domain

// Score は0-100の範囲のスコアを表す値オブジェクト。
type Score struct {
	Value int
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
