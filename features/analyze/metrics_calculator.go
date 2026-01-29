package analyze

import (
	"github.com/ryuka-games/lokup/domain"
)

// metricsInput は calculateMetrics の入力パラメータ。
type metricsInput struct {
	commits           []Commit
	contributors      []Contributor
	closedPRs         []PullRequest
	openPRs           []PullRequest
	allIssues         []Issue
	openIssues        []Issue
	files             []File
	releases          []Release
	period            domain.DateRange
	avgReviewWaitTime float64
	avgPRSize         int
}

// calculateMetrics は各種メトリクスを計算する。
func (s *Service) calculateMetrics(in metricsInput) domain.Metrics {
	days := in.period.Days()
	if days == 0 {
		days = 1
	}

	// 深夜コミット率を計算
	lateNightRate := 0.0
	if len(in.commits) > 0 {
		lateNightRate = float64(countLateNightCommits(in.commits)) / float64(len(in.commits)) * 100
	}

	// PRリードタイム（作成からマージまでの平均日数）を計算
	avgLeadTime := s.calculateAvgLeadTime(in.closedPRs)

	// PR内訳を計算
	prb := s.calculatePRBreakdown(in.closedPRs)

	// Issue統計を計算
	is := s.calculateIssueStats(in.allIssues, in.period)

	// DORA メトリクス
	deployFreq, deployRating := s.calculateDeployFrequency(in.releases, in.period)
	cfr, cfrRating := s.calculateChangeFailureRate(in.allIssues, in.releases, in.commits, in.period)
	mttr, mttrRating := s.calculateMTTR(in.allIssues, in.period)

	// コードチャーン
	revertCount := countRevertCommits(in.commits)
	revertRate := 0.0
	if len(in.commits) > 0 {
		revertRate = float64(revertCount) / float64(len(in.commits)) * 100
	}

	return domain.Metrics{
		// 開発速度
		TotalCommits:        len(in.commits),
		FeatureAdditionRate: float64(len(in.commits)) / float64(days),
		AvgLeadTime:         avgLeadTime,
		AvgReviewWaitTime:   in.avgReviewWaitTime,
		OpenPRCount:         len(in.openPRs),
		OpenIssueCount:      len(in.openIssues),

		// コード品質
		BugFixRatio:    prb.BugFixRatio,
		ReworkRate:     revertRate,
		AvgPRSize:      in.avgPRSize,
		IssueCloseRate: is.CloseRate,
		IssuesCreated:  is.Created,
		IssuesClosed:   is.Closed,

		// PR内訳
		FeaturePRCount: prb.Feature,
		BugFixPRCount:  prb.BugFix,
		OtherPRCount:   prb.Other,

		// DORA メトリクス
		DeployFrequency:   deployFreq,
		DeployFreqRating:  deployRating,
		ChangeFailureRate: cfr,
		ChangeFailRating:  cfrRating,
		MTTR:              mttr,
		MTTRRating:        mttrRating,

		// 投資比率
		RefactorPRCount: prb.Refactor,
		FeatureRatio:    prb.FeatureRatio,
		RefactorRatio:   prb.RefactorRatio,

		// コードチャーン
		RevertCommitCount: revertCount,
		RevertRate:        revertRate,

		// チーム健全性
		TotalFiles:          len(in.files),
		TotalContributors:   len(in.contributors),
		LateNightCommitRate: lateNightRate,
	}
}

// prBreakdown はPR内訳の結果。
type prBreakdown struct {
	Feature       int
	BugFix        int
	Refactor      int
	Other         int
	BugFixRatio   float64
	FeatureRatio  float64
	RefactorRatio float64
}

// calculatePRBreakdown はマージ済みPRの内訳を計算する。
func (s *Service) calculatePRBreakdown(pullRequests []PullRequest) prBreakdown {
	var b prBreakdown
	for _, pr := range pullRequests {
		if pr.MergedAt != nil {
			if pr.IsFeature() {
				b.Feature++
			} else if pr.IsBugFix() {
				b.BugFix++
			} else if pr.IsRefactor() {
				b.Refactor++
			} else {
				b.Other++
			}
		}
	}

	total := b.Feature + b.BugFix + b.Refactor + b.Other
	if total > 0 {
		b.BugFixRatio = float64(b.BugFix) / float64(total) * 100
		b.FeatureRatio = float64(b.Feature) / float64(total) * 100
		b.RefactorRatio = float64(b.Refactor) / float64(total) * 100
	}
	return b
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

// issueStats はIssue統計の結果。
type issueStats struct {
	Created   int
	Closed    int
	CloseRate float64
}

// calculateIssueStats は期間中のIssue作成・クローズ数を計算する。
func (s *Service) calculateIssueStats(issues []Issue, period domain.DateRange) issueStats {
	var st issueStats
	for _, issue := range issues {
		if !issue.CreatedAt.Before(period.From) && !issue.CreatedAt.After(period.To) {
			st.Created++
		}
		if issue.ClosedAt != nil && !issue.ClosedAt.Before(period.From) && !issue.ClosedAt.After(period.To) {
			st.Closed++
		}
	}

	if st.Created > 0 {
		st.CloseRate = float64(st.Closed) / float64(st.Created) * 100
	}
	return st
}
