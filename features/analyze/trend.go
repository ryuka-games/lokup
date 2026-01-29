package analyze

import (
	"math"

	"github.com/ryuka-games/lokup/domain"
)

// ── トレンド比較 ─────────────────────────────────────────────

// calculateTrends は今期と前期のメトリクスを比較してトレンドを算出する。
func (s *Service) calculateTrends(current domain.Metrics, prevCommits []Commit, prevIssues []Issue, prevPeriod domain.DateRange) []domain.TrendDelta {
	var trends []domain.TrendDelta

	// コミット数トレンド
	prevCommitCount := len(prevCommits)
	trends = append(trends, buildTrendDelta("コミット数", float64(current.TotalCommits), float64(prevCommitCount)))

	// コミット頻度トレンド
	prevDays := prevPeriod.Days()
	if prevDays == 0 {
		prevDays = 1
	}
	prevRate := float64(prevCommitCount) / float64(prevDays)
	trends = append(trends, buildTrendDelta("コミット頻度", current.FeatureAdditionRate, prevRate))

	// Issueクローズ率トレンド
	prevIS := (&Service{}).calculateIssueStats(prevIssues, prevPeriod)
	trends = append(trends, buildTrendDelta("Issueクローズ率", current.IssueCloseRate, prevIS.CloseRate))

	return trends
}

// buildTrendDelta はトレンドデルタを構築する。
func buildTrendDelta(name string, current, previous float64) domain.TrendDelta {
	deltaPct := 0.0
	if previous > 0 {
		deltaPct = (current - previous) / previous * 100
	}

	direction := "same"
	if math.Abs(deltaPct) > 5 {
		if deltaPct > 0 {
			direction = "up"
		} else {
			direction = "down"
		}
	}

	return domain.TrendDelta{
		MetricName:    name,
		CurrentValue:  current,
		PreviousValue: previous,
		DeltaPct:      deltaPct,
		Direction:     direction,
	}
}
