package analyze

import (
	"context"
	"fmt"

	"github.com/ryuka-games/lokup/domain"
)

// PR詳細取得の上限
const maxPRDetailsCount = 20

// countLateNightCommits は深夜（22時〜5時）のコミット数を返す。
func countLateNightCommits(commits []Commit) int {
	count := 0
	for _, c := range commits {
		hour := c.Date.Hour()
		if hour >= lateNightStartHour || hour < lateNightEndHour {
			count++
		}
	}
	return count
}

// buildPRDetails はマージ済みPRからPR詳細一覧を構築する。
// レビュー情報もここで取得し、PRDetailに含める。
func (s *Service) buildPRDetails(ctx context.Context, repo domain.Repository, pullRequests []PullRequest) []domain.PRDetail {
	var details []domain.PRDetail

	// 最新の20件のマージ済みPRから詳細を構築（APIコール節約）
	count := 0
	for _, pr := range pullRequests {
		if pr.MergedAt == nil {
			continue
		}
		if count >= maxPRDetailsCount {
			break
		}
		count++

		leadTime := pr.LeadTime()

		// PR詳細を取得（additions/deletions）
		size := 0
		prDetail, detailErr := s.repo.GetPRDetail(ctx, repo, pr.Number)
		if detailErr == nil {
			size = prDetail.Additions + prDetail.Deletions
		}

		// レビュー待ち時間を計算
		var reviewWaitHours float64
		reviews, err := s.repo.GetPRReviews(ctx, repo, pr.Number)
		if err == nil && len(reviews) > 0 {
			firstReview := reviews[0]
			for _, r := range reviews {
				if r.SubmittedAt.Before(firstReview.SubmittedAt) {
					firstReview = r
				}
			}
			waitTime := firstReview.SubmittedAt.Sub(pr.CreatedAt).Hours()
			if waitTime >= 0 {
				reviewWaitHours = waitTime
			}
		}

		additions := 0
		deletions := 0
		if detailErr == nil {
			additions = prDetail.Additions
			deletions = prDetail.Deletions
		}

		details = append(details, domain.PRDetail{
			Number:          pr.Number,
			Title:           pr.Title,
			Author:          pr.Author,
			LeadTimeDays:    leadTime,
			Size:            size,
			Additions:       additions,
			Deletions:       deletions,
			ReviewWaitHours: reviewWaitHours,
		})
	}

	return details
}

// calcAvgPRSize はPR詳細一覧から平均PRサイズを計算する。
func calcAvgPRSize(details []domain.PRDetail) int {
	var total, count int
	for _, d := range details {
		if d.Size > 0 {
			total += d.Size
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

// calcAvgReviewWait はPR詳細一覧から平均レビュー待ち時間を計算する。
func calcAvgReviewWait(details []domain.PRDetail) float64 {
	var total float64
	var count int
	for _, d := range details {
		if d.ReviewWaitHours > 0 {
			total += d.ReviewWaitHours
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// buildContributorDetails はコントリビューター詳細一覧を構築する。
func (s *Service) buildContributorDetails(contributors []Contributor) []domain.ContributorDetail {
	totalCommits := 0
	for _, c := range contributors {
		totalCommits += c.Contributions
	}

	details := make([]domain.ContributorDetail, len(contributors))
	for i, c := range contributors {
		ratio := 0.0
		if totalCommits > 0 {
			ratio = float64(c.Contributions) / float64(totalCommits) * 100
		}
		details[i] = domain.ContributorDetail{
			Name:    c.Login,
			Commits: c.Contributions,
			Ratio:   ratio,
		}
	}

	return details
}

// aggregateHourlyCommits はコミットを時間帯別に集計する。
func (s *Service) aggregateHourlyCommits(commits []Commit) [24]int {
	var hourly [24]int
	for _, c := range commits {
		hourly[c.Date.Hour()]++
	}
	return hourly
}

// aggregateDailyCommits はコミットを日別に集計する。
func (s *Service) aggregateDailyCommits(commits []Commit, period domain.DateRange) []domain.DailyCommit {
	// 日付ごとのコミット数をカウント
	countByDate := make(map[string]int)
	for _, c := range commits {
		dateKey := c.Date.Format("2006-01-02")
		countByDate[dateKey]++
	}

	// 期間内の全日付を生成（コミットがない日も0として含める）
	var result []domain.DailyCommit
	current := period.From
	for !current.After(period.To) {
		dateKey := current.Format("2006-01-02")
		result = append(result, domain.DailyCommit{
			Date:  current,
			Count: countByDate[dateKey],
		})
		current = current.AddDate(0, 0, 1)
	}

	return result
}

// formatAge は月数を「X年Yヶ月」形式にフォーマットする。
func formatAge(months int) string {
	years := months / 12
	remainingMonths := months % 12

	if years == 0 {
		return fmt.Sprintf("%dヶ月", remainingMonths)
	}
	if remainingMonths == 0 {
		return fmt.Sprintf("%d年", years)
	}
	return fmt.Sprintf("%d年%dヶ月", years, remainingMonths)
}
