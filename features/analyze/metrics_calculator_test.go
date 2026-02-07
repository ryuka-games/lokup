package analyze

import (
	"testing"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

func TestCalculatePRBreakdown(t *testing.T) {
	s := &Service{}
	merged := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	prs := []PullRequest{
		{HeadBranch: "feature/login", MergedAt: &merged},
		{HeadBranch: "feature/signup", MergedAt: &merged},
		{HeadBranch: "fix/bug-123", MergedAt: &merged},
		{HeadBranch: "refactor/cleanup", MergedAt: &merged},
		{HeadBranch: "misc/something", MergedAt: &merged},
		{HeadBranch: "feature/not-merged", MergedAt: nil}, // not merged
	}

	b := s.calculatePRBreakdown(prs)

	if b.Feature != 2 {
		t.Errorf("Feature = %d, want 2", b.Feature)
	}
	if b.BugFix != 1 {
		t.Errorf("BugFix = %d, want 1", b.BugFix)
	}
	if b.Refactor != 1 {
		t.Errorf("Refactor = %d, want 1", b.Refactor)
	}
	if b.Other != 1 {
		t.Errorf("Other = %d, want 1", b.Other)
	}

	// 5 merged total
	if b.FeatureRatio != 40.0 {
		t.Errorf("FeatureRatio = %v, want 40.0", b.FeatureRatio)
	}
	if b.BugFixRatio != 20.0 {
		t.Errorf("BugFixRatio = %v, want 20.0", b.BugFixRatio)
	}
	if b.RefactorRatio != 20.0 {
		t.Errorf("RefactorRatio = %v, want 20.0", b.RefactorRatio)
	}
}

func TestCalculatePRBreakdown_empty(t *testing.T) {
	s := &Service{}
	b := s.calculatePRBreakdown(nil)
	if b.Feature != 0 || b.BugFix != 0 || b.Refactor != 0 || b.Other != 0 {
		t.Error("expected all zeros")
	}
	if b.FeatureRatio != 0 {
		t.Errorf("FeatureRatio = %v, want 0", b.FeatureRatio)
	}
}

func TestCalculateAvgLeadTime(t *testing.T) {
	s := &Service{}

	t.Run("merged PRs", func(t *testing.T) {
		m1 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
		m2 := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
		prs := []PullRequest{
			{CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), MergedAt: &m1}, // 2 days
			{CreatedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), MergedAt: &m2}, // 4 days
			{CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), MergedAt: nil}, // not merged
		}
		avg := s.calculateAvgLeadTime(prs)
		if avg != 3.0 {
			t.Errorf("avg = %v, want 3.0", avg)
		}
	})

	t.Run("no merged PRs", func(t *testing.T) {
		prs := []PullRequest{
			{MergedAt: nil},
		}
		avg := s.calculateAvgLeadTime(prs)
		if avg != 0 {
			t.Errorf("avg = %v, want 0", avg)
		}
	})

	t.Run("empty", func(t *testing.T) {
		avg := s.calculateAvgLeadTime(nil)
		if avg != 0 {
			t.Errorf("avg = %v, want 0", avg)
		}
	})
}

func TestCalculateIssueStats(t *testing.T) {
	s := &Service{}
	period := domain.NewDateRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	)

	closedAt := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	issues := []Issue{
		{CreatedAt: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), ClosedAt: &closedAt},  // created+closed in period
		{CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), ClosedAt: nil},       // created, not closed
		{CreatedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), ClosedAt: &closedAt}, // created outside, closed in period
		{CreatedAt: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC), ClosedAt: nil},       // created, not closed
	}

	st := s.calculateIssueStats(issues, period)

	if st.Created != 3 {
		t.Errorf("Created = %d, want 3", st.Created)
	}
	if st.Closed != 2 {
		t.Errorf("Closed = %d, want 2", st.Closed)
	}
	// CloseRate = 2/3 * 100 ≈ 66.67
	expectedRate := float64(2) / float64(3) * 100
	if st.CloseRate != expectedRate {
		t.Errorf("CloseRate = %v, want %v", st.CloseRate, expectedRate)
	}
}

func TestCalculateIssueStats_empty(t *testing.T) {
	s := &Service{}
	period := domain.NewDateRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	st := s.calculateIssueStats(nil, period)
	if st.Created != 0 || st.Closed != 0 || st.CloseRate != 0 {
		t.Error("expected all zeros")
	}
}
