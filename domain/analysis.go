package domain

import "time"

// DateRange は分析期間を表す値オブジェクト。
type DateRange struct {
	From time.Time
	To   time.Time
}

// NewDateRange は DateRange を生成する。
func NewDateRange(from, to time.Time) DateRange {
	return DateRange{From: from, To: to}
}

// Days は期間の日数を返す。
func (d DateRange) Days() int {
	return int(d.To.Sub(d.From).Hours() / 24)
}

// CategoryScore はカテゴリごとのスコアと診断。
type CategoryScore struct {
	Category  Category // カテゴリ
	Score     Score    // スコア（0-100）
	Diagnosis string   // 一行診断テキスト
}

// PRDetail はPRの詳細情報（ドリルダウン表示用）。
type PRDetail struct {
	Number          int     // PR番号
	Title           string  // タイトル
	Author          string  // 作成者
	LeadTimeDays    float64 // リードタイム（日）
	Size            int     // 変更行数（追加+削除）
	Additions       int     // 追加行数
	Deletions       int     // 削除行数
	ReviewWaitHours float64 // レビュー待ち時間（時間）
}

// TrendDelta は前期比較のデルタ値を表す。
type TrendDelta struct {
	MetricName    string  `json:"metricName"`    // メトリクス名
	CurrentValue  float64 `json:"currentValue"`  // 今期の値
	PreviousValue float64 `json:"previousValue"` // 前期の値
	DeltaPct      float64 `json:"deltaPct"`      // 変化率（%）
	Direction     string  `json:"direction"`     // "up", "down", "same"
}

// ContributorDetail はコントリビューターの詳細（ドリルダウン表示用）。
type ContributorDetail struct {
	Name    string  // ユーザー名
	Commits int     // コミット数
	Ratio   float64 // 全体に占める割合（%）
}

// AnalysisResult は分析結果を表す集約。
// これが集約ルートであり、診断結果全体を束ねる。
type AnalysisResult struct {
	Repository      Repository                // 対象リポジトリ
	Period          DateRange                 // 分析期間
	CategoryScores  map[Category]CategoryScore // カテゴリ別スコア
	OverallScore    Score                     // 総合スコア（カテゴリ平均）
	Risks           []Risk                    // 検出されたリスク
	Metrics         Metrics                   // 各種メトリクス
	DailyCommits    []DailyCommit             // 日別コミット数
	LargeFiles      []LargeFile               // 巨大ファイル一覧
	OutdatedDeps    []OutdatedDep             // 古い依存一覧
	PRDetails       []PRDetail                // PR詳細一覧（ドリルダウン用）
	ContributorDetails []ContributorDetail     // コントリビューター詳細（ドリルダウン用）
	HourlyCommits   [24]int                   // 時間帯別コミット数（ドリルダウン用）
	Trends          []TrendDelta              // 前期比較トレンド
	GeneratedAt     time.Time                 // レポート生成日時
}

// DailyCommit は1日分のコミット数を表す。
type DailyCommit struct {
	Date  time.Time
	Count int
}

// LargeFile は巨大ファイル情報を表す。
type LargeFile struct {
	Path     string   // ファイルパス
	SizeKB   int      // サイズ（KB）
	Severity Severity // 重大度
}

// OutdatedDep は古い依存情報を表す。
type OutdatedDep struct {
	Name     string   // パッケージ名
	Version  string   // 使用中のバージョン
	Age      string   // 経過期間（例: "2年3ヶ月"）
	Severity Severity // 重大度
}

// Metrics は各種メトリクスを表す。
type Metrics struct {
	// 開発速度メトリクス
	TotalCommits        int     // 総コミット数
	FeatureAdditionRate float64 // 機能追加速度（コミット/日）
	AvgLeadTime         float64 // PR作成→マージの平均日数
	AvgReviewWaitTime   float64 // 最初のレビューまでの平均時間（時間）
	OpenPRCount         int     // オープンPR数
	OpenIssueCount      int     // オープンIssue数

	// コード品質メトリクス
	BugFixRatio      float64 // バグ修正の割合（%）
	ReworkRate       float64 // 手戻り率（%）
	AvgPRSize        int     // PRあたりの平均変更行数
	IssueCloseRate   float64 // Issueクローズ率（%）
	IssuesCreated    int     // 期間中に作成されたIssue数
	IssuesClosed     int     // 期間中にクローズされたIssue数

	// PR内訳
	FeaturePRCount int // feature PRの件数
	BugFixPRCount  int // bugfix PRの件数
	OtherPRCount   int // その他PRの件数

	// DORA メトリクス
	DeployFrequency   float64 // デプロイ頻度（リリース/月）
	DeployFreqRating  string  // DORAレーティング（Elite/High/Medium/Low）
	ChangeFailureRate float64 // 変更失敗率（%）
	ChangeFailRating  string  // DORAレーティング
	MTTR              float64 // 平均復旧時間（時間）
	MTTRRating        string  // DORAレーティング

	// 投資比率（PR分類拡張）
	RefactorPRCount int     // リファクタリングPR数
	FeatureRatio    float64 // 機能追加率（%）
	RefactorRatio   float64 // リファクタリング率（%）

	// コードチャーン
	RevertCommitCount int     // Revertコミット数
	RevertRate        float64 // Revert率（%）

	// チーム健全性メトリクス
	TotalFiles          int     // 総ファイル数
	TotalContributors   int     // コントリビューター数
	LateNightCommitRate float64 // 深夜コミット率（%）
}

// RiskCount は重大度別のリスク数を返す。
func (a *AnalysisResult) RiskCount(severity Severity) int {
	count := 0
	for _, r := range a.Risks {
		if r.Severity == severity {
			count++
		}
	}
	return count
}

// HighRisks は高リスク（🔴）の一覧を返す。
func (a *AnalysisResult) HighRisks() []Risk {
	var risks []Risk
	for _, r := range a.Risks {
		if r.Severity == SeverityHigh {
			risks = append(risks, r)
		}
	}
	return risks
}
