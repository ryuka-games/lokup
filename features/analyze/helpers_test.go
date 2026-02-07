package analyze

import (
	"testing"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

func TestCountLateNightCommits(t *testing.T) {
	tests := []struct {
		name    string
		commits []Commit
		want    int
	}{
		{
			"no commits",
			nil,
			0,
		},
		{
			"all daytime",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
			},
			0,
		},
		{
			"late night 22-23h",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 22, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 23, 30, 0, 0, time.UTC)},
			},
			2,
		},
		{
			"early morning 0-4h",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 4, 59, 0, 0, time.UTC)},
			},
			3,
		},
		{
			"boundary: 5h is not late night",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 5, 0, 0, 0, time.UTC)},
			},
			0,
		},
		{
			"boundary: 21h is not late night",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 21, 59, 0, 0, time.UTC)},
			},
			0,
		},
		{
			"mixed",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)}, // day
				{Date: time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC)}, // late
				{Date: time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC)},  // late
				{Date: time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)}, // day
			},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countLateNightCommits(tt.commits)
			if got != tt.want {
				t.Errorf("countLateNightCommits() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalcAvgPRSize(t *testing.T) {
	tests := []struct {
		name    string
		details []domain.PRDetail
		want    int
	}{
		{"empty", nil, 0},
		{"single", []domain.PRDetail{{Size: 100}}, 100},
		{
			"average",
			[]domain.PRDetail{{Size: 100}, {Size: 200}, {Size: 300}},
			200,
		},
		{
			"skip zero size",
			[]domain.PRDetail{{Size: 0}, {Size: 200}},
			200,
		},
		{"all zero", []domain.PRDetail{{Size: 0}, {Size: 0}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcAvgPRSize(tt.details)
			if got != tt.want {
				t.Errorf("calcAvgPRSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalcAvgReviewWait(t *testing.T) {
	tests := []struct {
		name    string
		details []domain.PRDetail
		want    float64
	}{
		{"empty", nil, 0},
		{"single", []domain.PRDetail{{ReviewWaitHours: 24.0}}, 24.0},
		{
			"average",
			[]domain.PRDetail{{ReviewWaitHours: 10.0}, {ReviewWaitHours: 20.0}},
			15.0,
		},
		{
			"skip zero",
			[]domain.PRDetail{{ReviewWaitHours: 0}, {ReviewWaitHours: 30.0}},
			30.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcAvgReviewWait(tt.details)
			if got != tt.want {
				t.Errorf("calcAvgReviewWait() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		months int
		want   string
	}{
		{0, "0ヶ月"},
		{6, "6ヶ月"},
		{12, "1年"},
		{24, "2年"},
		{27, "2年3ヶ月"},
		{36, "3年"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatAge(tt.months)
			if got != tt.want {
				t.Errorf("formatAge(%d) = %q, want %q", tt.months, got, tt.want)
			}
		})
	}
}

func TestBuildContributorDetails(t *testing.T) {
	s := &Service{}
	contributors := []Contributor{
		{Login: "alice", Contributions: 75},
		{Login: "bob", Contributions: 25},
	}

	details := s.buildContributorDetails(contributors)

	if len(details) != 2 {
		t.Fatalf("len = %d, want 2", len(details))
	}
	if details[0].Name != "alice" {
		t.Errorf("details[0].Name = %q, want %q", details[0].Name, "alice")
	}
	if details[0].Ratio != 75.0 {
		t.Errorf("details[0].Ratio = %v, want 75.0", details[0].Ratio)
	}
	if details[1].Ratio != 25.0 {
		t.Errorf("details[1].Ratio = %v, want 25.0", details[1].Ratio)
	}
}

func TestBuildContributorDetails_empty(t *testing.T) {
	s := &Service{}
	details := s.buildContributorDetails(nil)
	if len(details) != 0 {
		t.Errorf("len = %d, want 0", len(details))
	}
}

func TestAggregateHourlyCommits(t *testing.T) {
	s := &Service{}
	commits := []Commit{
		{Date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
		{Date: time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC)},
		{Date: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
		{Date: time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)},
	}

	hourly := s.aggregateHourlyCommits(commits)

	if hourly[10] != 3 {
		t.Errorf("hourly[10] = %d, want 3", hourly[10])
	}
	if hourly[14] != 1 {
		t.Errorf("hourly[14] = %d, want 1", hourly[14])
	}
	if hourly[0] != 0 {
		t.Errorf("hourly[0] = %d, want 0", hourly[0])
	}
}

func TestAggregateDailyCommits(t *testing.T) {
	s := &Service{}
	period := domain.NewDateRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
	)
	commits := []Commit{
		{Date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
		{Date: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
		{Date: time.Date(2025, 1, 3, 9, 0, 0, 0, time.UTC)},
	}

	daily := s.aggregateDailyCommits(commits, period)

	if len(daily) != 3 {
		t.Fatalf("len = %d, want 3", len(daily))
	}
	if daily[0].Count != 2 {
		t.Errorf("day 1 count = %d, want 2", daily[0].Count)
	}
	if daily[1].Count != 0 {
		t.Errorf("day 2 count = %d, want 0", daily[1].Count)
	}
	if daily[2].Count != 1 {
		t.Errorf("day 3 count = %d, want 1", daily[2].Count)
	}
}
