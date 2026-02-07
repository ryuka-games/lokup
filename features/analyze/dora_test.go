package analyze

import (
	"testing"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

func TestCalculateDeployFrequency(t *testing.T) {
	s := &Service{}
	period := domain.NewDateRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	) // 30 days

	t.Run("no releases", func(t *testing.T) {
		freq, rating := s.calculateDeployFrequency(nil, period)
		if freq != 0 {
			t.Errorf("freq = %v, want 0", freq)
		}
		if rating != "N/A" {
			t.Errorf("rating = %q, want N/A", rating)
		}
	})

	t.Run("releases in period", func(t *testing.T) {
		releases := []Release{
			{PublishedAt: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
			{PublishedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
			{PublishedAt: time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC)},
		}
		freq, rating := s.calculateDeployFrequency(releases, period)
		if freq != 3.0 { // 3 releases / (30/30 month) = 3/month
			t.Errorf("freq = %v, want 3.0", freq)
		}
		if rating != "Medium" {
			t.Errorf("rating = %q, want Medium", rating)
		}
	})

	t.Run("releases outside period excluded", func(t *testing.T) {
		releases := []Release{
			{PublishedAt: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)}, // outside
			{PublishedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},  // inside
		}
		freq, _ := s.calculateDeployFrequency(releases, period)
		if freq != 1.0 {
			t.Errorf("freq = %v, want 1.0", freq)
		}
	})
}

func TestDoraDeployFreqRating(t *testing.T) {
	tests := []struct {
		freq float64
		want string
	}{
		{30, "Elite"},
		{60, "Elite"},
		{4, "High"},
		{10, "High"},
		{1, "Medium"},
		{3, "Medium"},
		{0.5, "Low"},
		{0, "Low"},
	}
	for _, tt := range tests {
		got := doraDeployFreqRating(tt.freq)
		if got != tt.want {
			t.Errorf("doraDeployFreqRating(%v) = %q, want %q", tt.freq, got, tt.want)
		}
	}
}

func TestCalculateChangeFailureRate(t *testing.T) {
	s := &Service{}
	period := domain.NewDateRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	)

	t.Run("no deploys → N/A", func(t *testing.T) {
		cfr, rating := s.calculateChangeFailureRate(nil, nil, nil, period)
		if cfr != 0 {
			t.Errorf("cfr = %v, want 0", cfr)
		}
		if rating != "N/A" {
			t.Errorf("rating = %q, want N/A", rating)
		}
	})

	t.Run("with bug issues", func(t *testing.T) {
		releases := []Release{
			{PublishedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
			{PublishedAt: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)},
		}
		issues := []Issue{
			{
				CreatedAt: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
				Labels:    []string{"bug"},
			},
		}
		cfr, _ := s.calculateChangeFailureRate(issues, releases, nil, period)
		// 1 failure / 2 deploys = 50%
		if cfr != 50.0 {
			t.Errorf("cfr = %v, want 50.0", cfr)
		}
	})
}

func TestDoraChangeFailRating(t *testing.T) {
	tests := []struct {
		cfr  float64
		want string
	}{
		{0, "Elite"},
		{15, "Elite"},
		{16, "High"},
		{30, "High"},
		{31, "Medium"},
		{45, "Medium"},
		{46, "Low"},
	}
	for _, tt := range tests {
		got := doraChangeFailRating(tt.cfr)
		if got != tt.want {
			t.Errorf("doraChangeFailRating(%v) = %q, want %q", tt.cfr, got, tt.want)
		}
	}
}

func TestCalculateMTTR(t *testing.T) {
	s := &Service{}
	period := domain.NewDateRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	)

	t.Run("no bug issues → N/A", func(t *testing.T) {
		mttr, rating := s.calculateMTTR(nil, period)
		if mttr != 0 {
			t.Errorf("mttr = %v, want 0", mttr)
		}
		if rating != "N/A" {
			t.Errorf("rating = %q, want N/A", rating)
		}
	})

	t.Run("bug issues with close time", func(t *testing.T) {
		closedAt := time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC) // 36h later
		issues := []Issue{
			{
				CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
				ClosedAt:  &closedAt,
				Labels:    []string{"bug"},
			},
		}
		mttr, _ := s.calculateMTTR(issues, period)
		if mttr != 36.0 {
			t.Errorf("mttr = %v, want 36.0", mttr)
		}
	})

	t.Run("non-bug issues excluded", func(t *testing.T) {
		closedAt := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)
		issues := []Issue{
			{
				CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
				ClosedAt:  &closedAt,
				Labels:    []string{"enhancement"},
			},
		}
		mttr, rating := s.calculateMTTR(issues, period)
		if mttr != 0 {
			t.Errorf("mttr = %v, want 0 (non-bug excluded)", mttr)
		}
		if rating != "N/A" {
			t.Errorf("rating = %q, want N/A", rating)
		}
	})
}

func TestDoraMTTRRating(t *testing.T) {
	tests := []struct {
		mttr float64
		want string
	}{
		{0.5, "Elite"},
		{1, "High"},
		{23, "High"},
		{24, "Medium"},
		{167, "Medium"},
		{168, "Low"},
		{500, "Low"},
	}
	for _, tt := range tests {
		got := doraMTTRRating(tt.mttr)
		if got != tt.want {
			t.Errorf("doraMTTRRating(%v) = %q, want %q", tt.mttr, got, tt.want)
		}
	}
}

func TestCountRevertCommits(t *testing.T) {
	commits := []Commit{
		{Message: "feat: add feature"},
		{Message: "Revert \"feat: add feature\""},
		{Message: "fix: bug fix"},
		{Message: "Revert \"fix: bug fix\""},
	}
	got := countRevertCommits(commits)
	if got != 2 {
		t.Errorf("countRevertCommits() = %d, want 2", got)
	}
}

func TestCountRevertCommits_none(t *testing.T) {
	commits := []Commit{
		{Message: "feat: add feature"},
		{Message: "fix: bug fix"},
	}
	got := countRevertCommits(commits)
	if got != 0 {
		t.Errorf("countRevertCommits() = %d, want 0", got)
	}
}
