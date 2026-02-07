package report

import (
	"testing"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

func newTestResult() *domain.AnalysisResult {
	return &domain.AnalysisResult{
		Repository: domain.NewRepository("facebook", "react"),
		Period: domain.NewDateRange(
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		),
		CategoryScores: map[domain.Category]domain.CategoryScore{
			domain.CategoryVelocity: {
				Category:  domain.CategoryVelocity,
				Score:     domain.NewScore(85),
				Diagnosis: "良好な状態です",
			},
			domain.CategoryQuality: {
				Category:  domain.CategoryQuality,
				Score:     domain.NewScore(70),
				Diagnosis: "改善の余地があります",
			},
			domain.CategoryTechDebt: {
				Category:  domain.CategoryTechDebt,
				Score:     domain.NewScore(90),
				Diagnosis: "良好な状態です",
			},
			domain.CategoryHealth: {
				Category:  domain.CategoryHealth,
				Score:     domain.NewScore(60),
				Diagnosis: "深夜作業が多いです",
			},
		},
		OverallScore: domain.NewScore(76),
		Risks: []domain.Risk{
			{
				Type:        domain.RiskTypeLateNight,
				Severity:    domain.SeverityMedium,
				Target:      "リポジトリ全体",
				Description: "深夜のコミットが多いです",
				Value:       35,
				Threshold:   30,
			},
			{
				Type:        domain.RiskTypeChangeConcentration,
				Severity:    domain.SeverityHigh,
				Target:      "src/main.go",
				Description: "変更が集中しています",
				Value:       25,
				Threshold:   20,
			},
		},
		Metrics: domain.Metrics{
			TotalCommits:        150,
			FeatureAdditionRate: 5.0,
			TotalContributors:   8,
			LateNightCommitRate: 35.0,
			AvgLeadTime:         3.5,
			AvgReviewWaitTime:   12.0,
			OpenPRCount:         5,
			OpenIssueCount:      10,
			BugFixRatio:         25.0,
			AvgPRSize:           200,
			IssueCloseRate:      75.0,
			IssuesCreated:       20,
			IssuesClosed:        15,
			FeaturePRCount:      10,
			BugFixPRCount:       5,
			OtherPRCount:        3,
			DeployFrequency:     4.0,
			DeployFreqRating:    "High",
			ChangeFailureRate:   10.0,
			ChangeFailRating:    "Elite",
			MTTR:                8.0,
			MTTRRating:          "High",
			RefactorPRCount:     4,
			FeatureRatio:        45.5,
			RefactorRatio:       18.2,
			RevertCommitCount:   2,
			RevertRate:          1.3,
			TotalFiles:          500,
		},
		LargeFiles: []domain.LargeFile{
			{Path: "bundle.js", SizeKB: 150, Severity: domain.SeverityHigh},
		},
		OutdatedDeps: []domain.OutdatedDep{
			{Name: "lodash", Version: "3.0.0", Age: "3年", Severity: domain.SeverityHigh},
		},
		PRDetails: []domain.PRDetail{
			{Number: 1, Title: "feat: login", Author: "alice", LeadTimeDays: 2.0, Size: 100},
		},
		ContributorDetails: []domain.ContributorDetail{
			{Name: "alice", Commits: 80, Ratio: 53.3},
			{Name: "bob", Commits: 70, Ratio: 46.7},
		},
		Trends: []domain.TrendDelta{
			{MetricName: "コミット数", CurrentValue: 150, PreviousValue: 120, DeltaPct: 25.0, Direction: "up"},
		},
		GeneratedAt: time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC),
	}
}

func TestPrepareTemplateData(t *testing.T) {
	s := NewService()
	result := newTestResult()
	data := s.prepareTemplateData(result)

	t.Run("basic fields", func(t *testing.T) {
		if data.Repository != "facebook/react" {
			t.Errorf("Repository = %q, want %q", data.Repository, "facebook/react")
		}
		if data.PeriodFrom != "2025-01-01" {
			t.Errorf("PeriodFrom = %q", data.PeriodFrom)
		}
		if data.PeriodDays != 30 {
			t.Errorf("PeriodDays = %d, want 30", data.PeriodDays)
		}
	})

	t.Run("overall score", func(t *testing.T) {
		if data.OverallScore != 76 {
			t.Errorf("OverallScore = %d, want 76", data.OverallScore)
		}
		if data.OverallGrade != "B" {
			t.Errorf("OverallGrade = %q, want B", data.OverallGrade)
		}
		if data.OverallGradeClass != "grade-b" {
			t.Errorf("OverallGradeClass = %q, want grade-b", data.OverallGradeClass)
		}
	})

	t.Run("categories", func(t *testing.T) {
		if len(data.Categories) != 4 {
			t.Fatalf("Categories len = %d, want 4", len(data.Categories))
		}
		// 順番: Velocity, Quality, TechDebt, Health
		if data.Categories[0].Name != "開発速度" {
			t.Errorf("Categories[0].Name = %q, want 開発速度", data.Categories[0].Name)
		}
		if data.Categories[0].Score != 85 {
			t.Errorf("Categories[0].Score = %d, want 85", data.Categories[0].Score)
		}
	})

	t.Run("risks", func(t *testing.T) {
		if !data.HasRisks {
			t.Error("HasRisks = false, want true")
		}
		if len(data.Risks) != 2 {
			t.Errorf("Risks len = %d, want 2", len(data.Risks))
		}
	})

	t.Run("change concentration risks extracted", func(t *testing.T) {
		if len(data.ChangeConcentrationRisks) != 1 {
			t.Errorf("ChangeConcentrationRisks len = %d, want 1", len(data.ChangeConcentrationRisks))
		}
	})

	t.Run("metrics", func(t *testing.T) {
		if data.TotalCommits != 150 {
			t.Errorf("TotalCommits = %d, want 150", data.TotalCommits)
		}
		if data.DeployFrequency != 4.0 {
			t.Errorf("DeployFrequency = %v, want 4.0", data.DeployFrequency)
		}
	})

	t.Run("large files", func(t *testing.T) {
		if data.LargeFileCount != 1 {
			t.Errorf("LargeFileCount = %d, want 1", data.LargeFileCount)
		}
		if data.LargeFiles[0].Path != "bundle.js" {
			t.Errorf("LargeFiles[0].Path = %q", data.LargeFiles[0].Path)
		}
	})

	t.Run("outdated deps", func(t *testing.T) {
		if data.OutdatedDepCount != 1 {
			t.Errorf("OutdatedDepCount = %d, want 1", data.OutdatedDepCount)
		}
	})

	t.Run("generated at", func(t *testing.T) {
		if data.GeneratedAt != "2025-01-31 12:00:00" {
			t.Errorf("GeneratedAt = %q", data.GeneratedAt)
		}
	})
}

func TestRiskTypeToAction(t *testing.T) {
	// 全リスクタイプにアクションがあること
	riskTypes := []domain.RiskType{
		domain.RiskTypeChangeConcentration,
		domain.RiskTypeLargeFile,
		domain.RiskTypeOwnership,
		domain.RiskTypeOutdatedDeps,
		domain.RiskTypeLateNight,
		domain.RiskTypeSlowLeadTime,
		domain.RiskTypeSlowReview,
		domain.RiskTypeLargePR,
		domain.RiskTypeLowIssueClose,
		domain.RiskTypeBugFixHigh,
		domain.RiskTypeLowDeployFreq,
		domain.RiskTypeHighChangeFailure,
		domain.RiskTypeSlowRecovery,
		domain.RiskTypeLowFeatureInvestment,
	}
	for _, rt := range riskTypes {
		action := riskTypeToAction(rt)
		if action == "" {
			t.Errorf("riskTypeToAction(%q) returned empty", rt)
		}
		if action == "詳細を確認し、改善策を検討してください。" {
			t.Errorf("riskTypeToAction(%q) returned fallback", rt)
		}
	}
}

func TestRiskTypeToAction_unknown(t *testing.T) {
	action := riskTypeToAction(domain.RiskType("unknown"))
	if action != "詳細を確認し、改善策を検討してください。" {
		t.Errorf("unexpected action for unknown: %q", action)
	}
}

func TestGenerateOverallDiagnosis(t *testing.T) {
	categories := []CategoryScoreData{
		{Name: "開発速度", Score: 80},
		{Name: "コード品質", Score: 50},
		{Name: "技術的負債", Score: 90},
		{Name: "チーム健全性", Score: 70},
	}

	tests := []struct {
		grade string
		want  string
	}{
		{"A", "全体的に良好な状態です。"},
		{"B", "概ね良好ですが、コード品質に改善の余地があります。"},
		{"C", "コード品質を中心に改善が必要です。"},
		{"D", "コード品質に重大な課題があります。早急な対応を推奨します。"},
	}
	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			got := generateOverallDiagnosis(tt.grade, categories)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDateWithWeekday(t *testing.T) {
	tests := []struct {
		date time.Time
		want string
	}{
		{time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC), "1/25(土)"},
		{time.Date(2025, 1, 26, 0, 0, 0, 0, time.UTC), "1/26(日)"},
		{time.Date(2025, 1, 27, 0, 0, 0, 0, time.UTC), "1/27(月)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDateWithWeekday(tt.date)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerate_createsFile(t *testing.T) {
	s := NewService()
	result := newTestResult()

	tmpFile := t.TempDir() + "/test-report.html"
	err := s.Generate(result, tmpFile)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}
