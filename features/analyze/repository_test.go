package analyze

import (
	"testing"
	"time"
)

func TestPullRequestLeadTime(t *testing.T) {
	merged := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		pr   PullRequest
		want float64
	}{
		{
			"merged in 3 days",
			PullRequest{
				CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				MergedAt:  &merged,
			},
			3.0,
		},
		{
			"not merged returns -1",
			PullRequest{
				CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				MergedAt:  nil,
			},
			-1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pr.LeadTime()
			if got != tt.want {
				t.Errorf("LeadTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPullRequestIsBugFix(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"fix/login-bug", true},
		{"bugfix/issue-123", true},
		{"hotfix/urgent", true},
		{"feature/new-feature", false},
		{"chore/update-deps", false},
		{"FIX/uppercase", true},
	}
	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			pr := PullRequest{HeadBranch: tt.branch}
			if got := pr.IsBugFix(); got != tt.want {
				t.Errorf("IsBugFix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPullRequestIsFeature(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"feature/new-ui", true},
		{"feat/add-login", true},
		{"fix/bug", false},
		{"chore/deps", false},
		{"FEATURE/caps", true},
	}
	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			pr := PullRequest{HeadBranch: tt.branch}
			if got := pr.IsFeature(); got != tt.want {
				t.Errorf("IsFeature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPullRequestIsRefactor(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"refactor/cleanup", true},
		{"chore/update-deps", true},
		{"debt/reduce-tech-debt", true},
		{"ci/fix-pipeline", true},
		{"docs/update-readme", true},
		{"feature/new-thing", false},
		{"fix/bug", false},
	}
	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			pr := PullRequest{HeadBranch: tt.branch}
			if got := pr.IsRefactor(); got != tt.want {
				t.Errorf("IsRefactor() = %v, want %v", got, tt.want)
			}
		})
	}
}
