package domain

import "testing"

func TestNewScore(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{"zero", 0, 0},
		{"normal", 72, 72},
		{"max", 100, 100},
		{"negative clamps to 0", -10, 0},
		{"over 100 clamps to 100", 150, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScore(tt.input)
			if got.Value != tt.want {
				t.Errorf("NewScore(%d).Value = %d, want %d", tt.input, got.Value, tt.want)
			}
		})
	}
}

func TestNewScoreWithBreakdown(t *testing.T) {
	breakdown := []ScoreBreakdownItem{
		{Label: "基本スコア", Points: 100},
		{Label: "リスク", Points: -15, Detail: "テスト"},
	}
	got := NewScoreWithBreakdown(85, breakdown)

	if got.Value != 85 {
		t.Errorf("Value = %d, want 85", got.Value)
	}
	if len(got.Breakdown) != 2 {
		t.Errorf("Breakdown len = %d, want 2", len(got.Breakdown))
	}
	if got.Breakdown[0].Label != "基本スコア" {
		t.Errorf("Breakdown[0].Label = %q, want %q", got.Breakdown[0].Label, "基本スコア")
	}
}

func TestScoreGrade(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		{"0 is D", 0, "D"},
		{"39 is D", 39, "D"},
		{"40 is C", 40, "C"},
		{"59 is C", 59, "C"},
		{"60 is B", 60, "B"},
		{"79 is B", 79, "B"},
		{"80 is A", 80, "A"},
		{"100 is A", 100, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScore(tt.score).Grade()
			if got != tt.want {
				t.Errorf("Score(%d).Grade() = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

func TestScoreGradeDescription(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "良好"},
		{70, "普通"},
		{50, "要改善"},
		{20, "危険"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := NewScore(tt.score).GradeDescription()
			if got != tt.want {
				t.Errorf("Score(%d).GradeDescription() = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}
