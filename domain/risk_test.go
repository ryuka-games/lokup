package domain

import "testing"

func TestRiskTypeDisplayName(t *testing.T) {
	tests := []struct {
		riskType RiskType
		want     string
	}{
		{RiskTypeChangeConcentration, "変更集中リスク"},
		{RiskTypeLargeFile, "巨大ファイル"},
		{RiskTypeOwnership, "属人化"},
		{RiskTypeOutdatedDeps, "依存の古さ"},
		{RiskTypeLateNight, "深夜労働"},
		{RiskTypeSlowLeadTime, "PRリードタイム超過"},
		{RiskTypeSlowReview, "レビュー待ち超過"},
		{RiskTypeLargePR, "PRサイズ超過"},
		{RiskTypeLowIssueClose, "Issueクローズ率低下"},
		{RiskTypeBugFixHigh, "バグ修正割合過多"},
		{RiskTypeLowDeployFreq, "デプロイ頻度不足"},
		{RiskTypeHighChangeFailure, "変更失敗率過多"},
		{RiskTypeSlowRecovery, "復旧時間超過"},
		{RiskTypeLowFeatureInvestment, "機能投資不足"},
	}
	for _, tt := range tests {
		t.Run(string(tt.riskType), func(t *testing.T) {
			got := tt.riskType.DisplayName()
			if got != tt.want {
				t.Errorf("RiskType(%q).DisplayName() = %q, want %q", tt.riskType, got, tt.want)
			}
		})
	}
}

func TestRiskTypeDisplayName_unknown(t *testing.T) {
	unknown := RiskType("unknown_type")
	got := unknown.DisplayName()
	if got != "unknown_type" {
		t.Errorf("unknown RiskType.DisplayName() = %q, want %q", got, "unknown_type")
	}
}

func TestRiskTypeCategory(t *testing.T) {
	tests := []struct {
		riskType RiskType
		want     Category
	}{
		// Velocity
		{RiskTypeSlowLeadTime, CategoryVelocity},
		{RiskTypeSlowReview, CategoryVelocity},
		{RiskTypeLowDeployFreq, CategoryVelocity},
		{RiskTypeSlowRecovery, CategoryVelocity},
		// Quality
		{RiskTypeChangeConcentration, CategoryQuality},
		{RiskTypeLargePR, CategoryQuality},
		{RiskTypeLowIssueClose, CategoryQuality},
		{RiskTypeBugFixHigh, CategoryQuality},
		{RiskTypeHighChangeFailure, CategoryQuality},
		// Tech Debt
		{RiskTypeLargeFile, CategoryTechDebt},
		{RiskTypeOutdatedDeps, CategoryTechDebt},
		{RiskTypeLowFeatureInvestment, CategoryTechDebt},
		// Health
		{RiskTypeLateNight, CategoryHealth},
		{RiskTypeOwnership, CategoryHealth},
	}
	for _, tt := range tests {
		t.Run(string(tt.riskType), func(t *testing.T) {
			got := tt.riskType.Category()
			if got != tt.want {
				t.Errorf("RiskType(%q).Category() = %q, want %q", tt.riskType, got, tt.want)
			}
		})
	}
}

func TestSeverityEmoji(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityLow, "🟢"},
		{SeverityMedium, "🟡"},
		{SeverityHigh, "🔴"},
		{Severity(99), "⚪"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.Emoji()
			if got != tt.want {
				t.Errorf("Severity(%d).Emoji() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityLow, "低"},
		{SeverityMedium, "中"},
		{SeverityHigh, "高"},
		{Severity(99), "不明"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("Severity(%d).String() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestNewRisk(t *testing.T) {
	r := NewRisk(RiskTypeLargeFile, SeverityHigh, "main.go", 120, 100)

	if r.Type != RiskTypeLargeFile {
		t.Errorf("Type = %q, want %q", r.Type, RiskTypeLargeFile)
	}
	if r.Severity != SeverityHigh {
		t.Errorf("Severity = %d, want %d", r.Severity, SeverityHigh)
	}
	if r.Target != "main.go" {
		t.Errorf("Target = %q, want %q", r.Target, "main.go")
	}
	if r.Value != 120 {
		t.Errorf("Value = %d, want 120", r.Value)
	}
	if r.Threshold != 100 {
		t.Errorf("Threshold = %d, want 100", r.Threshold)
	}
}
