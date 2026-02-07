package analyze

import (
	"testing"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

func TestDetectChangeConcentration(t *testing.T) {
	s := &Service{}

	// 1ファイルに20回以上の変更 → SeverityHigh
	// 1ファイルに10-19回の変更 → SeverityMedium
	commits := make([]Commit, 25)
	for i := range commits {
		commits[i] = Commit{Files: []string{"hot-file.go"}}
	}
	// medium 用に12回変更されるファイルを追加
	for i := 0; i < 12; i++ {
		commits[i].Files = append(commits[i].Files, "warm-file.go")
	}

	risks := s.detectChangeConcentration(commits)

	var highCount, mediumCount int
	for _, r := range risks {
		if r.Type != domain.RiskTypeChangeConcentration {
			t.Errorf("unexpected risk type: %v", r.Type)
		}
		switch r.Severity {
		case domain.SeverityHigh:
			highCount++
		case domain.SeverityMedium:
			mediumCount++
		}
	}
	if highCount != 1 {
		t.Errorf("high risks = %d, want 1", highCount)
	}
	if mediumCount != 1 {
		t.Errorf("medium risks = %d, want 1", mediumCount)
	}
}

func TestDetectChangeConcentration_noRisk(t *testing.T) {
	s := &Service{}
	commits := []Commit{
		{Files: []string{"a.go", "b.go"}},
		{Files: []string{"c.go"}},
	}
	risks := s.detectChangeConcentration(commits)
	if len(risks) != 0 {
		t.Errorf("expected no risks, got %d", len(risks))
	}
}

func TestDetectOwnershipRisk(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name         string
		contributors []Contributor
		wantRisks    int
	}{
		{
			"single contributor (100%) → risk",
			[]Contributor{{Login: "alice", Contributions: 100}},
			1,
		},
		{
			"dominant contributor (80%) → risk",
			[]Contributor{
				{Login: "alice", Contributions: 80},
				{Login: "bob", Contributions: 20},
			},
			1,
		},
		{
			"balanced (50/50) → no risk",
			[]Contributor{
				{Login: "alice", Contributions: 50},
				{Login: "bob", Contributions: 50},
			},
			0,
		},
		{
			"empty → no risk",
			nil,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risks := s.detectOwnershipRisk(tt.contributors)
			if len(risks) != tt.wantRisks {
				t.Errorf("got %d risks, want %d", len(risks), tt.wantRisks)
			}
		})
	}
}

func TestDetectLateNightRisk(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name      string
		commits   []Commit
		wantRisks int
	}{
		{"empty", nil, 0},
		{
			"below threshold (20%)",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC)},
			},
			0,
		},
		{
			"above threshold (50%)",
			[]Commit{
				{Date: time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
				{Date: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risks := s.detectLateNightRisk(tt.commits)
			if len(risks) != tt.wantRisks {
				t.Errorf("got %d risks, want %d", len(risks), tt.wantRisks)
			}
		})
	}
}

func TestDetectLargeFiles(t *testing.T) {
	s := &Service{}
	files := []File{
		{Path: "small.go", Size: 1024},            // 1KB - OK
		{Path: "medium.go", Size: 60 * 1024},      // 60KB - Medium
		{Path: "large.go", Size: 120 * 1024},      // 120KB - High
		{Path: "also-medium.go", Size: 80 * 1024}, // 80KB - Medium
	}

	risks, largeFiles := s.detectLargeFiles(files)

	// リスクは集計される（High x 1件, Medium x 2件 → 2リスク）
	if len(risks) != 2 {
		t.Errorf("risks = %d, want 2", len(risks))
	}

	// 詳細ファイル一覧は3件
	if len(largeFiles) != 3 {
		t.Errorf("largeFiles = %d, want 3", len(largeFiles))
	}
}

func TestDetectOutdatedDeps(t *testing.T) {
	s := &Service{}
	deps := []Dependency{
		{Name: "fresh", AgeMonths: 6},
		{Name: "old", AgeMonths: 26, Version: "1.0.0"},     // 2年以上 → Medium
		{Name: "ancient", AgeMonths: 40, Version: "0.5.0"}, // 3年以上 → High
	}

	risks, outdatedDeps := s.detectOutdatedDeps(deps)

	if len(risks) != 2 {
		t.Errorf("risks = %d, want 2", len(risks))
	}
	if len(outdatedDeps) != 2 {
		t.Errorf("outdatedDeps = %d, want 2", len(outdatedDeps))
	}
}

func TestDetectMetricRisks(t *testing.T) {
	s := &Service{}

	t.Run("slow lead time", func(t *testing.T) {
		m := domain.Metrics{AvgLeadTime: 10.0} // > 7 days
		risks := s.detectMetricRisks(m)
		found := false
		for _, r := range risks {
			if r.Type == domain.RiskTypeSlowLeadTime {
				found = true
			}
		}
		if !found {
			t.Error("expected RiskTypeSlowLeadTime")
		}
	})

	t.Run("slow review", func(t *testing.T) {
		m := domain.Metrics{AvgReviewWaitTime: 72.0} // > 48h
		risks := s.detectMetricRisks(m)
		found := false
		for _, r := range risks {
			if r.Type == domain.RiskTypeSlowReview {
				found = true
			}
		}
		if !found {
			t.Error("expected RiskTypeSlowReview")
		}
	})

	t.Run("large PR", func(t *testing.T) {
		m := domain.Metrics{AvgPRSize: 600} // > 500
		risks := s.detectMetricRisks(m)
		found := false
		for _, r := range risks {
			if r.Type == domain.RiskTypeLargePR {
				found = true
			}
		}
		if !found {
			t.Error("expected RiskTypeLargePR")
		}
	})

	t.Run("high bug fix ratio", func(t *testing.T) {
		m := domain.Metrics{BugFixRatio: 60.0} // > 50%
		risks := s.detectMetricRisks(m)
		found := false
		for _, r := range risks {
			if r.Type == domain.RiskTypeBugFixHigh {
				found = true
			}
		}
		if !found {
			t.Error("expected RiskTypeBugFixHigh")
		}
	})

	t.Run("no risks when metrics are good", func(t *testing.T) {
		m := domain.Metrics{
			AvgLeadTime:       3.0,
			AvgReviewWaitTime: 12.0,
			AvgPRSize:         200,
			BugFixRatio:       20.0,
			IssueCloseRate:    80.0,
			DeployFrequency:   4.0,
			ChangeFailureRate: 10.0,
			MTTR:              2.0,
			FeatureRatio:      50.0,
			FeaturePRCount:    5,
			BugFixPRCount:     2,
		}
		risks := s.detectMetricRisks(m)
		if len(risks) != 0 {
			t.Errorf("expected no risks, got %d", len(risks))
			for _, r := range risks {
				t.Logf("  risk: %v", r.Type)
			}
		}
	})
}

func TestCalculateCategoryScores(t *testing.T) {
	s := &Service{}

	t.Run("no risks → all 100", func(t *testing.T) {
		scores := s.calculateCategoryScores(nil)
		for cat, cs := range scores {
			if cs.Score.Value != 100 {
				t.Errorf("category %v score = %d, want 100", cat, cs.Score.Value)
			}
		}
	})

	t.Run("high risk reduces score by 15", func(t *testing.T) {
		risks := []domain.Risk{
			{Type: domain.RiskTypeHighChangeFailure, Severity: domain.SeverityHigh},
		}
		scores := s.calculateCategoryScores(risks)
		qualityScore := scores[domain.CategoryQuality].Score.Value
		if qualityScore != 85 {
			t.Errorf("quality score = %d, want 85", qualityScore)
		}
	})

	t.Run("medium risk reduces score by 10", func(t *testing.T) {
		risks := []domain.Risk{
			{Type: domain.RiskTypeLateNight, Severity: domain.SeverityMedium},
		}
		scores := s.calculateCategoryScores(risks)
		healthScore := scores[domain.CategoryHealth].Score.Value
		if healthScore != 90 {
			t.Errorf("health score = %d, want 90", healthScore)
		}
	})

	t.Run("multiple risks accumulate", func(t *testing.T) {
		risks := []domain.Risk{
			{Type: domain.RiskTypeLateNight, Severity: domain.SeverityMedium}, // Health -10
			{Type: domain.RiskTypeOwnership, Severity: domain.SeverityMedium}, // Health -10
		}
		scores := s.calculateCategoryScores(risks)
		healthScore := scores[domain.CategoryHealth].Score.Value
		if healthScore != 80 {
			t.Errorf("health score = %d, want 80", healthScore)
		}
	})

	t.Run("score floor is 0", func(t *testing.T) {
		// 7 x High = -105 → clamped to 0
		var risks []domain.Risk
		for i := 0; i < 7; i++ {
			risks = append(risks, domain.Risk{
				Type:     domain.RiskTypeChangeConcentration,
				Severity: domain.SeverityHigh,
			})
		}
		scores := s.calculateCategoryScores(risks)
		qualityScore := scores[domain.CategoryQuality].Score.Value
		if qualityScore != 0 {
			t.Errorf("quality score = %d, want 0", qualityScore)
		}
	})
}

func TestCalculateOverallScore(t *testing.T) {
	tests := []struct {
		name   string
		scores map[domain.Category]domain.CategoryScore
		want   int
	}{
		{"empty", nil, 0},
		{
			"all 100",
			map[domain.Category]domain.CategoryScore{
				domain.CategoryVelocity: {Score: domain.NewScore(100)},
				domain.CategoryQuality:  {Score: domain.NewScore(100)},
				domain.CategoryTechDebt: {Score: domain.NewScore(100)},
				domain.CategoryHealth:   {Score: domain.NewScore(100)},
			},
			100,
		},
		{
			"mixed",
			map[domain.Category]domain.CategoryScore{
				domain.CategoryVelocity: {Score: domain.NewScore(80)},
				domain.CategoryQuality:  {Score: domain.NewScore(60)},
				domain.CategoryTechDebt: {Score: domain.NewScore(100)},
				domain.CategoryHealth:   {Score: domain.NewScore(40)},
			},
			70,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOverallScore(tt.scores)
			if got.Value != tt.want {
				t.Errorf("calculateOverallScore() = %d, want %d", got.Value, tt.want)
			}
		})
	}
}

func TestGenerateDiagnosis(t *testing.T) {
	t.Run("grade A → good", func(t *testing.T) {
		got := generateDiagnosis(domain.CategoryHealth, domain.NewScore(90), nil)
		if got != "良好な状態です" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("grade B with late night risk", func(t *testing.T) {
		risk := &domain.Risk{Type: domain.RiskTypeLateNight}
		got := generateDiagnosis(domain.CategoryHealth, domain.NewScore(70), risk)
		if got != "深夜作業が多く、チームの持続可能性に懸念があります" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("no worst risk → good", func(t *testing.T) {
		got := generateDiagnosis(domain.CategoryQuality, domain.NewScore(70), nil)
		if got != "良好な状態です" {
			t.Errorf("got %q", got)
		}
	})
}
