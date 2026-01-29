package analyze

import (
	"strings"

	"github.com/ryuka-games/lokup/domain"
)

// ── DORA メトリクス計算 ──────────────────────────────────────

// calculateDeployFrequency は期間内のデプロイ頻度（リリース/月）とDORAレーティングを計算する。
func (s *Service) calculateDeployFrequency(releases []Release, period domain.DateRange) (float64, string) {
	if len(releases) == 0 {
		return 0, "N/A"
	}

	count := 0
	for _, r := range releases {
		if !r.PublishedAt.Before(period.From) && !r.PublishedAt.After(period.To) {
			count++
		}
	}

	days := period.Days()
	if days == 0 {
		days = 1
	}
	freq := float64(count) / (float64(days) / 30.0)

	rating := doraDeployFreqRating(freq)
	return freq, rating
}

// doraDeployFreqRating はデプロイ頻度からDORAレーティングを返す。
func doraDeployFreqRating(freq float64) string {
	switch {
	case freq >= 30: // daily or more
		return "Elite"
	case freq >= 4: // weekly
		return "High"
	case freq >= 1: // monthly
		return "Medium"
	default:
		return "Low"
	}
}

// calculateChangeFailureRate は変更失敗率（%）とDORAレーティングを計算する。
func (s *Service) calculateChangeFailureRate(issues []Issue, releases []Release, commits []Commit, period domain.DateRange) (float64, string) {
	// デプロイ数 = 期間内リリース数
	deployCount := 0
	for _, r := range releases {
		if !r.PublishedAt.Before(period.From) && !r.PublishedAt.After(period.To) {
			deployCount++
		}
	}
	if deployCount == 0 {
		return 0, "N/A"
	}

	// 障害指標: bug/incident/hotfixラベルのIssue + Revertコミット
	failureCount := 0
	for _, issue := range issues {
		if !issue.CreatedAt.Before(period.From) && !issue.CreatedAt.After(period.To) {
			for _, label := range issue.Labels {
				lower := strings.ToLower(label)
				if lower == "bug" || lower == "incident" || lower == "hotfix" {
					failureCount++
					break
				}
			}
		}
	}
	failureCount += countRevertCommits(commits)

	cfr := float64(failureCount) / float64(deployCount) * 100
	rating := doraChangeFailRating(cfr)
	return cfr, rating
}

// doraChangeFailRating は変更失敗率からDORAレーティングを返す。
func doraChangeFailRating(cfr float64) string {
	switch {
	case cfr <= 15:
		return "Elite"
	case cfr <= 30:
		return "High"
	case cfr <= 45:
		return "Medium"
	default:
		return "Low"
	}
}

// calculateMTTR は平均復旧時間（時間）とDORAレーティングを計算する。
func (s *Service) calculateMTTR(issues []Issue, period domain.DateRange) (float64, string) {
	var totalHours float64
	var count int

	for _, issue := range issues {
		if issue.ClosedAt == nil {
			continue
		}
		if issue.CreatedAt.Before(period.From) || issue.CreatedAt.After(period.To) {
			continue
		}
		// bugラベルのIssueのみ対象
		isBug := false
		for _, label := range issue.Labels {
			lower := strings.ToLower(label)
			if lower == "bug" || lower == "incident" || lower == "hotfix" {
				isBug = true
				break
			}
		}
		if !isBug {
			continue
		}

		hours := issue.ClosedAt.Sub(issue.CreatedAt).Hours()
		if hours >= 0 {
			totalHours += hours
			count++
		}
	}

	if count == 0 {
		return 0, "N/A"
	}

	mttr := totalHours / float64(count)
	rating := doraMTTRRating(mttr)
	return mttr, rating
}

// doraMTTRRating はMTTRからDORAレーティングを返す。
func doraMTTRRating(mttr float64) string {
	switch {
	case mttr < 1:
		return "Elite"
	case mttr < 24:
		return "High"
	case mttr < 168: // 1 week
		return "Medium"
	default:
		return "Low"
	}
}

// countRevertCommits はRevertコミット数をカウントする。
func countRevertCommits(commits []Commit) int {
	count := 0
	for _, c := range commits {
		if strings.HasPrefix(c.Message, "Revert ") {
			count++
		}
	}
	return count
}
