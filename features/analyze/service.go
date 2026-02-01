package analyze

import (
	"context"
	"log"
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
	closedPRs, err := s.repo.GetPullRequests(ctx, input.Repository, "closed")
	if err != nil {
		return nil, err
	}

	// オープンPRを取得
	openPRs, err := s.repo.GetPullRequests(ctx, input.Repository, "open")
	if err != nil {
		return nil, err
	}

	// Issue一覧を取得（期間内の作成・クローズを計算）
	periodStart := input.Period.From
	allIssues, err := s.repo.GetIssues(ctx, input.Repository, "all", &periodStart)
	if err != nil {
		return nil, err
	}

	// オープンIssue数を取得
	openIssues, err := s.repo.GetIssues(ctx, input.Repository, "open", nil)
	if err != nil {
		return nil, err
	}

	// ファイル一覧を取得（巨大ファイル検出用）
	files, err := s.repo.GetFiles(ctx, input.Repository)
	if err != nil {
		return nil, err
	}

	// 依存情報を取得（古い依存検出用）
	dependencies, err := s.repo.GetDependencies(ctx, input.Repository)
	if err != nil {
		return nil, err
	}

	// リリース一覧を取得（DORA デプロイ頻度用）
	releases, err := s.repo.GetReleases(ctx, input.Repository)
	if err != nil {
		log.Printf("Warning: failed to get releases: %v", err)
		releases = nil
	}

	// 前期データを取得（トレンド比較用）
	prevPeriodDays := input.Period.Days()
	prevTo := input.Period.From.AddDate(0, 0, -1)
	prevFrom := prevTo.AddDate(0, 0, -prevPeriodDays)
	prevPeriod := domain.NewDateRange(prevFrom, prevTo)

	prevCommits, err := s.repo.GetCommits(ctx, input.Repository, prevPeriod)
	if err != nil {
		log.Printf("Warning: failed to get previous period commits: %v", err)
		prevCommits = nil
	}

	prevPeriodStart := prevPeriod.From
	prevIssues, err := s.repo.GetIssues(ctx, input.Repository, "all", &prevPeriodStart)
	if err != nil {
		log.Printf("Warning: failed to get previous period issues: %v", err)
		prevIssues = nil
	}

	// レビュー情報を取得しPR詳細を構築（APIコール共有）
	prDetails := s.buildPRDetails(ctx, input.Repository, closedPRs)

	// レビュー待ち時間の平均を計算
	avgReviewWaitTime := calcAvgReviewWait(prDetails)

	// PRサイズの平均をPR詳細から計算
	avgPRSize := calcAvgPRSize(prDetails)

	// 2. リスク検出
	risks, largeFiles := s.detectRisks(commits, contributors, files)

	// 古い依存の検出
	outdatedRisks, outdatedDeps := s.detectOutdatedDeps(dependencies)
	risks = append(risks, outdatedRisks...)

	// 3. メトリクス計算
	metrics := s.calculateMetrics(metricsInput{
		commits:           commits,
		contributors:      contributors,
		closedPRs:         closedPRs,
		openPRs:           openPRs,
		allIssues:         allIssues,
		openIssues:        openIssues,
		files:             files,
		releases:          releases,
		period:            input.Period,
		avgReviewWaitTime: avgReviewWaitTime,
		avgPRSize:         avgPRSize,
	})

	// 4. メトリクスベースのリスク検出
	metricRisks := s.detectMetricRisks(metrics)
	risks = append(risks, metricRisks...)

	// 5. カテゴリ別スコア計算
	categoryScores := s.calculateCategoryScores(risks)

	// 5b. 総合スコア計算
	overallScore := calculateOverallScore(categoryScores)

	// 6. 日別コミット数を集計
	dailyCommits := s.aggregateDailyCommits(commits, input.Period)

	// 7. ドリルダウンデータ構築
	contributorDetails := s.buildContributorDetails(contributors)
	hourlyCommits := s.aggregateHourlyCommits(commits)

	// 8. トレンド比較
	trends := s.calculateTrends(metrics, prevCommits, prevIssues, prevPeriod)

	// 9. 結果を組み立て
	return &domain.AnalysisResult{
		Repository:         input.Repository,
		Period:             input.Period,
		CategoryScores:     categoryScores,
		OverallScore:       overallScore,
		Risks:              risks,
		Metrics:            metrics,
		DailyCommits:       dailyCommits,
		LargeFiles:         largeFiles,
		OutdatedDeps:       outdatedDeps,
		PRDetails:          prDetails,
		ContributorDetails: contributorDetails,
		HourlyCommits:      hourlyCommits,
		Trends:             trends,
		GeneratedAt:        time.Now(),
	}, nil
}
