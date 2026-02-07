package domain

import (
	"testing"
	"time"
)

func TestDateRangeDays(t *testing.T) {
	tests := []struct {
		name string
		from time.Time
		to   time.Time
		want int
	}{
		{
			"30 days",
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			30,
		},
		{
			"same day",
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			0,
		},
		{
			"1 day",
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDateRange(tt.from, tt.to)
			got := dr.Days()
			if got != tt.want {
				t.Errorf("DateRange.Days() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAnalysisResultRiskCount(t *testing.T) {
	result := &AnalysisResult{
		Risks: []Risk{
			{Severity: SeverityHigh},
			{Severity: SeverityHigh},
			{Severity: SeverityMedium},
			{Severity: SeverityLow},
		},
	}

	tests := []struct {
		severity Severity
		want     int
	}{
		{SeverityHigh, 2},
		{SeverityMedium, 1},
		{SeverityLow, 1},
	}
	for _, tt := range tests {
		t.Run(tt.severity.String(), func(t *testing.T) {
			got := result.RiskCount(tt.severity)
			if got != tt.want {
				t.Errorf("RiskCount(%v) = %d, want %d", tt.severity, got, tt.want)
			}
		})
	}
}

func TestAnalysisResultHighRisks(t *testing.T) {
	result := &AnalysisResult{
		Risks: []Risk{
			{Type: RiskTypeLargeFile, Severity: SeverityHigh},
			{Type: RiskTypeLateNight, Severity: SeverityMedium},
			{Type: RiskTypeHighChangeFailure, Severity: SeverityHigh},
		},
	}

	got := result.HighRisks()
	if len(got) != 2 {
		t.Fatalf("HighRisks() len = %d, want 2", len(got))
	}
	if got[0].Type != RiskTypeLargeFile {
		t.Errorf("HighRisks()[0].Type = %q, want %q", got[0].Type, RiskTypeLargeFile)
	}
	if got[1].Type != RiskTypeHighChangeFailure {
		t.Errorf("HighRisks()[1].Type = %q, want %q", got[1].Type, RiskTypeHighChangeFailure)
	}
}

func TestAnalysisResultRiskCount_empty(t *testing.T) {
	result := &AnalysisResult{}
	if got := result.RiskCount(SeverityHigh); got != 0 {
		t.Errorf("RiskCount on empty = %d, want 0", got)
	}
}
