package analyze

import (
	"context"
	"fmt"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

// Service は分析のビジネスロジックを担当する。
type Service struct {
	repo Repository
}

// NewService は Service を生成する。
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ServiceInput は Service.Analyze の入力。
type ServiceInput struct {
	Repository domain.Repository
	Period     domain.DateRange
}

// Analyze はリポジトリを分析し、結果を返す。
func (s *Service) Analyze(ctx context.Context, input ServiceInput) (*domain.AnalysisResult, error) {
	// 1. データ取得
	commits, err := s.repo.GetCommits(ctx, input.Repository, input.Period)
	if err != nil {
		return nil, err
	}

	contributors, err := s.repo.GetContributors(ctx, input.Repository)
	if err != nil {
		return nil, err
	}

	// マージ済みPRを取得（リードタイム計算用）
	pullRequests, err := s.repo.GetPullRequests(ctx, input.Repository, "closed")
	if err != nil {
		return nil, err
	}

	// 2. リスク検出
	risks := s.detectRisks(commits, contributors)

	// 3. スコア計算
	efficiencyScore := s.calculateEfficiencyScore(commits, risks)
	healthScore := s.calculateHealthScore(risks)

	// 4. メトリクス計算
	metrics := s.calculateMetrics(commits, contributors, pullRequests, input.Period)

	// 5. 結果を組み立て
	return &domain.AnalysisResult{
		Repository:      input.Repository,
		Period:          input.Period,
		EfficiencyScore: efficiencyScore,
		HealthScore:     healthScore,
		Risks:           risks,
		Metrics:         metrics,
		GeneratedAt:     time.Now(),
	}, nil
}

// detectRisks はコミット履歴からリスクを検出する。
func (s *Service) detectRisks(commits []Commit, contributors []Contributor) []domain.Risk {
	var risks []domain.Risk

	// 変更集中リスクの検出
	risks = append(risks, s.detectChangeConcentration(commits)...)

	// 属人化リスクの検出
	risks = append(risks, s.detectOwnershipRisk(contributors)...)

	// 深夜労働リスクの検出
	risks = append(risks, s.detectLateNightRisk(commits)...)

	return risks
}

// detectChangeConcentration は変更集中リスクを検出する。
func (s *Service) detectChangeConcentration(commits []Commit) []domain.Risk {
	var risks []domain.Risk

	// ファイルごとの変更回数をカウント
	fileChanges := make(map[string]int)
	for _, c := range commits {
		for _, f := range c.Files {
			fileChanges[f]++
		}
	}

	// 閾値を超えたファイルをリスクとして報告
	const warningThreshold = 10
	const criticalThreshold = 20

	for file, count := range fileChanges {
		if count >= criticalThreshold {
			risks = append(risks, domain.NewRisk(
				domain.RiskTypeChangeConcentration,
				domain.SeverityHigh,
				file,
				count,
				criticalThreshold,
			))
		} else if count >= warningThreshold {
			risks = append(risks, domain.NewRisk(
				domain.RiskTypeChangeConcentration,
				domain.SeverityMedium,
				file,
				count,
				warningThreshold,
			))
		}
	}

	return risks
}

// detectOwnershipRisk は属人化リスクを検出する。
func (s *Service) detectOwnershipRisk(contributors []Contributor) []domain.Risk {
	var risks []domain.Risk

	if len(contributors) == 0 {
		return risks
	}

	// 総コミット数を計算
	totalCommits := 0
	for _, c := range contributors {
		totalCommits += c.Contributions
	}

	if totalCommits == 0 {
		return risks
	}

	// トップコントリビューターの割合を計算
	topContributor := contributors[0]
	ratio := float64(topContributor.Contributions) / float64(totalCommits)

	const ownershipThreshold = 0.8 // 80%以上で属人化
	if ratio >= ownershipThreshold {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeOwnership,
			Severity:    domain.SeverityMedium,
			Target:      topContributor.Login,
			Description: "1人のコントリビューターがコミットの大部分を占めています",
			Value:       int(ratio * 100),
			Threshold:   int(ownershipThreshold * 100),
		})
	}

	return risks
}

// detectLateNightRisk は深夜労働リスクを検出する。
func (s *Service) detectLateNightRisk(commits []Commit) []domain.Risk {
	var risks []domain.Risk

	if len(commits) == 0 {
		return risks
	}

	// 深夜コミット（22時〜5時）をカウント
	lateNightCount := 0
	for _, c := range commits {
		hour := c.Date.Hour()
		if hour >= 22 || hour < 5 {
			lateNightCount++
		}
	}

	ratio := float64(lateNightCount) / float64(len(commits))
	const lateNightThreshold = 0.3 // 30%以上で警告

	if ratio >= lateNightThreshold {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLateNight,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: "深夜のコミットが多いです",
			Value:       int(ratio * 100),
			Threshold:   int(lateNightThreshold * 100),
		})
	}

	return risks
}

// calculateEfficiencyScore は開発効率スコアを計算する。
func (s *Service) calculateEfficiencyScore(commits []Commit, risks []domain.Risk) domain.Score {
	// 基本スコア
	score := 100
	breakdown := []domain.ScoreBreakdownItem{
		{Label: "基本スコア", Points: 100},
	}

	// リスクに応じて減点
	for _, r := range risks {
		var points int
		switch r.Severity {
		case domain.SeverityHigh:
			points = -15
		case domain.SeverityMedium:
			points = -10
		case domain.SeverityLow:
			points = -5
		}
		score += points
		breakdown = append(breakdown, domain.ScoreBreakdownItem{
			Label:  r.Type.DisplayName(),
			Points: points,
			Detail: formatRiskDetail(r),
		})
	}

	return domain.NewScoreWithBreakdown(score, breakdown)
}

// formatRiskDetail はリスクの詳細を文字列にフォーマットする。
func formatRiskDetail(r domain.Risk) string {
	if r.Value == 0 && r.Threshold == 0 {
		return ""
	}

	switch r.Type {
	case domain.RiskTypeLateNight:
		return fmt.Sprintf("22-5時のコミットが%d%%、基準%d%%以下", r.Value, r.Threshold)
	case domain.RiskTypeOwnership:
		return fmt.Sprintf("1人で%d%%のコミット、基準%d%%以下", r.Value, r.Threshold)
	case domain.RiskTypeChangeConcentration:
		return fmt.Sprintf("%d回変更、基準%d回以下", r.Value, r.Threshold)
	default:
		return fmt.Sprintf("%d / 基準%d", r.Value, r.Threshold)
	}
}

// calculateHealthScore はコード健全性スコアを計算する。
func (s *Service) calculateHealthScore(risks []domain.Risk) domain.Score {
	score := 100
	breakdown := []domain.ScoreBreakdownItem{
		{Label: "基本スコア", Points: 100},
	}

	for _, r := range risks {
		var points int
		switch r.Severity {
		case domain.SeverityHigh:
			points = -10
		case domain.SeverityMedium:
			points = -5
		case domain.SeverityLow:
			points = -2
		}
		score += points
		breakdown = append(breakdown, domain.ScoreBreakdownItem{
			Label:  r.Type.DisplayName(),
			Points: points,
			Detail: formatRiskDetail(r),
		})
	}

	return domain.NewScoreWithBreakdown(score, breakdown)
}

// calculateMetrics は各種メトリクスを計算する。
func (s *Service) calculateMetrics(commits []Commit, contributors []Contributor, pullRequests []PullRequest, period domain.DateRange) domain.Metrics {
	days := period.Days()
	if days == 0 {
		days = 1
	}

	// 深夜コミット率を計算
	lateNightCount := 0
	for _, c := range commits {
		hour := c.Date.Hour()
		if hour >= 22 || hour < 5 {
			lateNightCount++
		}
	}
	lateNightRate := 0.0
	if len(commits) > 0 {
		lateNightRate = float64(lateNightCount) / float64(len(commits)) * 100
	}

	// PRリードタイム（作成からマージまでの平均日数）を計算
	avgLeadTime := s.calculateAvgLeadTime(pullRequests)

	return domain.Metrics{
		TotalCommits:        len(commits),
		FeatureAdditionRate: float64(len(commits)) / float64(days),
		TotalContributors:   len(contributors),
		LateNightCommitRate: lateNightRate,
		AvgLeadTime:         avgLeadTime,
	}
}

// calculateAvgLeadTime はマージ済みPRの平均リードタイム（日数）を計算する。
func (s *Service) calculateAvgLeadTime(pullRequests []PullRequest) float64 {
	var totalLeadTime float64
	var mergedCount int

	for _, pr := range pullRequests {
		leadTime := pr.LeadTime()
		if leadTime >= 0 { // マージ済みのみ
			totalLeadTime += leadTime
			mergedCount++
		}
	}

	if mergedCount == 0 {
		return 0
	}

	return totalLeadTime / float64(mergedCount)
}
