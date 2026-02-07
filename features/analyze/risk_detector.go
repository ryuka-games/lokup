package analyze

import (
	"fmt"

	"github.com/ryuka-games/lokup/domain"
)

// ── リスク検出の閾値 ─────────────────────────────────────────

const (
	// 変更集中リスク
	changeConcentrationWarning  = 10 // 変更回数（warning）
	changeConcentrationCritical = 20 // 変更回数（critical）

	// 属人化リスク
	ownershipThreshold = 0.8 // コミット割合（80%以上で属人化）

	// 深夜労働リスク
	lateNightStartHour     = 22  // 深夜開始（22時）
	lateNightEndHour       = 5   // 深夜終了（5時）
	lateNightRateThreshold = 0.3 // 深夜コミット割合（30%以上で警告）

	// 巨大ファイル
	largeFileWarningBytes  = 50 * 1024  // 50KB
	largeFileCriticalBytes = 100 * 1024 // 100KB

	// 古い依存
	outdatedDepWarningMonths  = 24 // 2年
	outdatedDepCriticalMonths = 36 // 3年

	// メトリクスベースのリスク閾値
	leadTimeThresholdDays      = 7.0  // PRリードタイム（日）
	reviewWaitThresholdHours   = 48.0 // レビュー待ち（時間）
	prSizeThresholdLines       = 500  // PRサイズ（行）
	issueCloseRateThresholdPct = 50.0 // Issueクローズ率（%）
	bugFixRatioThresholdPct    = 50.0 // バグ修正割合（%）

	// DORA メトリクス閾値
	deployFreqThresholdPerMonth   = 1.0  // 月1回未満でリスク
	changeFailureThresholdPct     = 30.0 // 30%超でリスク
	mttrThresholdHours            = 24.0 // 24時間超でリスク
	featureInvestmentThresholdPct = 30.0 // 機能追加30%未満でリスク

	// スコア計算
	baseScore     = 100 // カテゴリスコアの初期値
	penaltyHigh   = -15 // SeverityHigh の減点
	penaltyMedium = -10 // SeverityMedium の減点
	penaltyLow    = -5  // SeverityLow の減点
)

// ── データソースに基づくリスク検出 ──────────────────────────────

// detectRisks はコミット履歴からリスクを検出する。
// リスク一覧と巨大ファイル一覧を返す。
func (s *Service) detectRisks(commits []Commit, contributors []Contributor, files []File) ([]domain.Risk, []domain.LargeFile) {
	var risks []domain.Risk

	// 変更集中リスクの検出
	risks = append(risks, s.detectChangeConcentration(commits)...)

	// 属人化リスクの検出
	risks = append(risks, s.detectOwnershipRisk(contributors)...)

	// 深夜労働リスクの検出
	risks = append(risks, s.detectLateNightRisk(commits)...)

	// 巨大ファイルリスクの検出
	largeFileRisks, largeFiles := s.detectLargeFiles(files)
	risks = append(risks, largeFileRisks...)

	return risks, largeFiles
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
	for file, count := range fileChanges {
		if count >= changeConcentrationCritical {
			risks = append(risks, domain.NewRisk(
				domain.RiskTypeChangeConcentration,
				domain.SeverityHigh,
				file,
				count,
				changeConcentrationCritical,
			))
		} else if count >= changeConcentrationWarning {
			risks = append(risks, domain.NewRisk(
				domain.RiskTypeChangeConcentration,
				domain.SeverityMedium,
				file,
				count,
				changeConcentrationWarning,
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

	lateNightCount := countLateNightCommits(commits)
	ratio := float64(lateNightCount) / float64(len(commits))

	if ratio >= lateNightRateThreshold {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLateNight,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: "深夜のコミットが多いです",
			Value:       int(ratio * 100),
			Threshold:   int(lateNightRateThreshold * 100),
		})
	}

	return risks
}

// detectLargeFiles は巨大ファイルリスクを検出する。
// 集計されたリスク（重大度ごとに1件）と、詳細なファイル一覧を返す。
func (s *Service) detectLargeFiles(files []File) ([]domain.Risk, []domain.LargeFile) {
	var risks []domain.Risk
	var largeFiles []domain.LargeFile

	var highCount, mediumCount int

	for _, f := range files {
		if f.Size >= largeFileCriticalBytes {
			highCount++
			largeFiles = append(largeFiles, domain.LargeFile{
				Path:     f.Path,
				SizeKB:   f.Size / 1024,
				Severity: domain.SeverityHigh,
			})
		} else if f.Size >= largeFileWarningBytes {
			mediumCount++
			largeFiles = append(largeFiles, domain.LargeFile{
				Path:     f.Path,
				SizeKB:   f.Size / 1024,
				Severity: domain.SeverityMedium,
			})
		}
	}

	// 集計されたリスクを作成
	if highCount > 0 {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLargeFile,
			Severity:    domain.SeverityHigh,
			Target:      fmt.Sprintf("%d件", highCount),
			Description: fmt.Sprintf("%dKB以上の巨大ファイルがあります", largeFileCriticalBytes/1024),
			Value:       highCount,
			Threshold:   largeFileCriticalBytes / 1024,
		})
	}
	if mediumCount > 0 {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLargeFile,
			Severity:    domain.SeverityMedium,
			Target:      fmt.Sprintf("%d件", mediumCount),
			Description: fmt.Sprintf("%dKB以上の大きいファイルがあります", largeFileWarningBytes/1024),
			Value:       mediumCount,
			Threshold:   largeFileWarningBytes / 1024,
		})
	}

	return risks, largeFiles
}

// detectOutdatedDeps は古い依存を検出する。
// 集計されたリスク（重大度ごとに1件）と、詳細な依存一覧を返す。
func (s *Service) detectOutdatedDeps(dependencies []Dependency) ([]domain.Risk, []domain.OutdatedDep) {
	var risks []domain.Risk
	var outdatedDeps []domain.OutdatedDep

	var highCount, mediumCount int

	for _, dep := range dependencies {
		if dep.AgeMonths >= outdatedDepCriticalMonths {
			highCount++
			outdatedDeps = append(outdatedDeps, domain.OutdatedDep{
				Name:     dep.Name,
				Version:  dep.Version,
				Age:      formatAge(dep.AgeMonths),
				Severity: domain.SeverityHigh,
			})
		} else if dep.AgeMonths >= outdatedDepWarningMonths {
			mediumCount++
			outdatedDeps = append(outdatedDeps, domain.OutdatedDep{
				Name:     dep.Name,
				Version:  dep.Version,
				Age:      formatAge(dep.AgeMonths),
				Severity: domain.SeverityMedium,
			})
		}
	}

	// 集計されたリスクを作成
	if highCount > 0 {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeOutdatedDeps,
			Severity:    domain.SeverityHigh,
			Target:      fmt.Sprintf("%d件", highCount),
			Description: fmt.Sprintf("%d年以上前の古い依存があります", outdatedDepCriticalMonths/12),
			Value:       highCount,
			Threshold:   outdatedDepCriticalMonths,
		})
	}
	if mediumCount > 0 {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeOutdatedDeps,
			Severity:    domain.SeverityMedium,
			Target:      fmt.Sprintf("%d件", mediumCount),
			Description: fmt.Sprintf("%d年以上前の古い依存があります", outdatedDepWarningMonths/12),
			Value:       mediumCount,
			Threshold:   outdatedDepWarningMonths,
		})
	}

	return risks, outdatedDeps
}

// ── メトリクスベースのリスク検出 ─────────────────────────────────

// detectMetricRisks はメトリクス値に基づいてリスクを検出する。
func (s *Service) detectMetricRisks(metrics domain.Metrics) []domain.Risk {
	var risks []domain.Risk

	// PRリードタイム
	if metrics.AvgLeadTime > leadTimeThresholdDays {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeSlowLeadTime,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("PRリードタイムが平均%.1f日です", metrics.AvgLeadTime),
			Value:       int(metrics.AvgLeadTime * 10),
			Threshold:   int(leadTimeThresholdDays),
		})
	}

	// レビュー待ち
	if metrics.AvgReviewWaitTime > reviewWaitThresholdHours {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeSlowReview,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("レビュー待ち時間が平均%.1f時間です", metrics.AvgReviewWaitTime),
			Value:       int(metrics.AvgReviewWaitTime * 10),
			Threshold:   int(reviewWaitThresholdHours),
		})
	}

	// PRサイズ
	if metrics.AvgPRSize > prSizeThresholdLines {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLargePR,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("PRの平均サイズが%d行です", metrics.AvgPRSize),
			Value:       metrics.AvgPRSize,
			Threshold:   prSizeThresholdLines,
		})
	}

	// Issueクローズ率（Issue作成がある場合のみ）
	if metrics.IssuesCreated > 0 && metrics.IssueCloseRate < issueCloseRateThresholdPct {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLowIssueClose,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("Issueクローズ率が%.1f%%です", metrics.IssueCloseRate),
			Value:       int(metrics.IssueCloseRate),
			Threshold:   int(issueCloseRateThresholdPct),
		})
	}

	// バグ修正割合
	if metrics.BugFixRatio > bugFixRatioThresholdPct {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeBugFixHigh,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("バグ修正PRの割合が%.1f%%です", metrics.BugFixRatio),
			Value:       int(metrics.BugFixRatio),
			Threshold:   int(bugFixRatioThresholdPct),
		})
	}

	// DORA: デプロイ頻度
	if metrics.DeployFrequency > 0 && metrics.DeployFrequency < deployFreqThresholdPerMonth {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLowDeployFreq,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("デプロイ頻度が月%.1f回です", metrics.DeployFrequency),
			Value:       int(metrics.DeployFrequency * 10),
			Threshold:   int(deployFreqThresholdPerMonth * 10),
		})
	}

	// DORA: 変更失敗率
	if metrics.ChangeFailureRate > changeFailureThresholdPct {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeHighChangeFailure,
			Severity:    domain.SeverityHigh,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("変更失敗率が%.1f%%です", metrics.ChangeFailureRate),
			Value:       int(metrics.ChangeFailureRate),
			Threshold:   int(changeFailureThresholdPct),
		})
	}

	// DORA: MTTR
	if metrics.MTTR > mttrThresholdHours {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeSlowRecovery,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("平均復旧時間が%.1f時間です", metrics.MTTR),
			Value:       int(metrics.MTTR * 10),
			Threshold:   int(mttrThresholdHours * 10),
		})
	}

	// 機能投資比率
	totalPRs := metrics.FeaturePRCount + metrics.BugFixPRCount + metrics.RefactorPRCount + metrics.OtherPRCount
	if totalPRs > 0 && metrics.FeatureRatio < featureInvestmentThresholdPct {
		risks = append(risks, domain.Risk{
			Type:        domain.RiskTypeLowFeatureInvestment,
			Severity:    domain.SeverityMedium,
			Target:      "リポジトリ全体",
			Description: fmt.Sprintf("機能追加PRの割合が%.1f%%です", metrics.FeatureRatio),
			Value:       int(metrics.FeatureRatio),
			Threshold:   int(featureInvestmentThresholdPct),
		})
	}

	return risks
}

// ── スコア計算・診断テキスト ─────────────────────────────────────

// calculateCategoryScores はカテゴリ別スコアを計算する。
func (s *Service) calculateCategoryScores(risks []domain.Risk) map[domain.Category]domain.CategoryScore {
	categories := []domain.Category{
		domain.CategoryVelocity,
		domain.CategoryQuality,
		domain.CategoryTechDebt,
		domain.CategoryHealth,
	}

	scores := make(map[domain.Category]domain.CategoryScore, len(categories))

	for _, cat := range categories {
		score := baseScore
		breakdown := []domain.ScoreBreakdownItem{
			{Label: "基本スコア", Points: baseScore},
		}

		// カテゴリに属するリスクのみで減点
		var worstRisk *domain.Risk
		var worstPoints int
		for _, r := range risks {
			if r.Type.Category() != cat {
				continue
			}
			var points int
			switch r.Severity {
			case domain.SeverityHigh:
				points = penaltyHigh
			case domain.SeverityMedium:
				points = penaltyMedium
			case domain.SeverityLow:
				points = penaltyLow
			}
			score += points
			breakdown = append(breakdown, domain.ScoreBreakdownItem{
				Label:  r.Type.DisplayName(),
				Points: points,
				Detail: formatRiskDetail(r),
			})
			if points < worstPoints {
				worstPoints = points
				rCopy := r
				worstRisk = &rCopy
			}
		}

		diagnosis := generateDiagnosis(cat, domain.NewScore(score), worstRisk)

		scores[cat] = domain.CategoryScore{
			Category:  cat,
			Score:     domain.NewScoreWithBreakdown(score, breakdown),
			Diagnosis: diagnosis,
		}
	}

	return scores
}

// calculateOverallScore はカテゴリ別スコアの平均から総合スコアを計算する。
func calculateOverallScore(categoryScores map[domain.Category]domain.CategoryScore) domain.Score {
	if len(categoryScores) == 0 {
		return domain.NewScore(0)
	}
	total := 0
	for _, cs := range categoryScores {
		total += cs.Score.Value
	}
	return domain.NewScore(total / len(categoryScores))
}

// generateDiagnosis はカテゴリスコアに応じた一行診断テキストを生成する。
func generateDiagnosis(cat domain.Category, score domain.Score, worstRisk *domain.Risk) string {
	if score.Grade() == "A" {
		return "良好な状態です"
	}

	if worstRisk == nil {
		return "良好な状態です"
	}

	switch worstRisk.Type {
	case domain.RiskTypeSlowLeadTime:
		return "PRリードタイムが長く、開発速度が低下しています"
	case domain.RiskTypeSlowReview:
		return "レビュー待ち時間が長く、フィードバックが遅延しています"
	case domain.RiskTypeChangeConcentration:
		return "特定ファイルへの変更が集中しており、品質リスクがあります"
	case domain.RiskTypeLargePR:
		return "PRサイズが大きく、レビューの質が低下する可能性があります"
	case domain.RiskTypeLowIssueClose:
		return "Issueの消化が追いつかず、負債が蓄積しています"
	case domain.RiskTypeBugFixHigh:
		return "バグ修正の割合が高く、品質に課題があります"
	case domain.RiskTypeLargeFile:
		return "巨大ファイルが多数あり、保守性に課題があります"
	case domain.RiskTypeOutdatedDeps:
		return "古い依存パッケージがあり、セキュリティリスクがあります"
	case domain.RiskTypeLateNight:
		return "深夜作業が多く、チームの持続可能性に懸念があります"
	case domain.RiskTypeOwnership:
		return "知識が特定の人に偏っており、属人化リスクがあります"
	case domain.RiskTypeLowDeployFreq:
		return "デプロイ頻度が低く、価値提供のスピードが遅れています"
	case domain.RiskTypeHighChangeFailure:
		return "変更失敗率が高く、リリース品質に課題があります"
	case domain.RiskTypeSlowRecovery:
		return "障害からの復旧時間が長く、運用に課題があります"
	case domain.RiskTypeLowFeatureInvestment:
		return "機能追加への投資比率が低く、負債対応に追われています"
	default:
		return "改善の余地があります"
	}
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
	case domain.RiskTypeLargeFile:
		return fmt.Sprintf("%d件、%dKB以上", r.Value, r.Threshold)
	case domain.RiskTypeOutdatedDeps:
		years := r.Threshold / 12
		return fmt.Sprintf("%d件、%d年以上前", r.Value, years)
	case domain.RiskTypeSlowLeadTime:
		return fmt.Sprintf("平均%.1f日、基準%d日以下", float64(r.Value)/10, r.Threshold)
	case domain.RiskTypeSlowReview:
		return fmt.Sprintf("平均%.1f時間、基準%d時間以下", float64(r.Value)/10, r.Threshold)
	case domain.RiskTypeLargePR:
		return fmt.Sprintf("平均%d行、基準%d行以下", r.Value, r.Threshold)
	case domain.RiskTypeLowIssueClose:
		return fmt.Sprintf("クローズ率%d%%、基準%d%%以上", r.Value, r.Threshold)
	case domain.RiskTypeBugFixHigh:
		return fmt.Sprintf("バグ修正%d%%、基準%d%%以下", r.Value, r.Threshold)
	case domain.RiskTypeLowDeployFreq:
		return fmt.Sprintf("月%.1f回、基準月%.1f回以上", float64(r.Value)/10, float64(r.Threshold)/10)
	case domain.RiskTypeHighChangeFailure:
		return fmt.Sprintf("失敗率%d%%、基準%d%%以下", r.Value, r.Threshold)
	case domain.RiskTypeSlowRecovery:
		return fmt.Sprintf("平均%.1f時間、基準%.1f時間以下", float64(r.Value)/10, float64(r.Threshold)/10)
	case domain.RiskTypeLowFeatureInvestment:
		return fmt.Sprintf("機能追加%d%%、基準%d%%以上", r.Value, r.Threshold)
	default:
		return fmt.Sprintf("%d / 基準%d", r.Value, r.Threshold)
	}
}
