package analyze

import (
	"context"
	"strings"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

// Repository はデータ取得のインターフェース。
// infrastructure/github パッケージで実装される。
//
// なぜ interface か:
// - テスト時にモックに差し替えるため
// - GitHub API 以外のデータソースにも対応できるようにするため
type Repository interface {
	// GetCommits は指定期間のコミット履歴を取得する。
	GetCommits(ctx context.Context, repo domain.Repository, period domain.DateRange) ([]Commit, error)

	// GetContributors はコントリビューター一覧を取得する。
	GetContributors(ctx context.Context, repo domain.Repository) ([]Contributor, error)

	// GetFileContent はファイルの内容を取得する。
	GetFileContent(ctx context.Context, repo domain.Repository, path string) ([]byte, error)

	// GetPullRequests はプルリクエスト一覧を取得する。
	GetPullRequests(ctx context.Context, repo domain.Repository, state string) ([]PullRequest, error)

	// GetFiles はリポジトリ内のファイル一覧を取得する。
	GetFiles(ctx context.Context, repo domain.Repository) ([]File, error)

	// GetDependencies はpackage.json等から依存情報を取得する。
	// 依存ファイルが存在しない場合は空のスライスを返す（エラーではない）。
	GetDependencies(ctx context.Context, repo domain.Repository) ([]Dependency, error)

	// GetIssues はIssue一覧を取得する。
	GetIssues(ctx context.Context, repo domain.Repository, state string, since *time.Time) ([]Issue, error)

	// GetPRReviews はPRのレビュー一覧を取得する。
	GetPRReviews(ctx context.Context, repo domain.Repository, prNumber int) ([]Review, error)

	// GetPRDetail はPRの詳細（additions/deletions含む）を取得する。
	GetPRDetail(ctx context.Context, repo domain.Repository, prNumber int) (*PullRequest, error)

	// GetReleases はリリース一覧を取得する。
	GetReleases(ctx context.Context, repo domain.Repository) ([]Release, error)
}

// File はファイル情報を表す。
type File struct {
	Path string // ファイルパス
	Size int    // サイズ（バイト）
}

// Commit はコミット情報を表す。
type Commit struct {
	SHA       string    // コミットハッシュ
	Author    string    // 作成者
	Email     string    // メールアドレス
	Date      time.Time // コミット日時
	Message   string    // コミットメッセージ
	Files     []string  // 変更されたファイル
	Additions int       // 追加行数
	Deletions int       // 削除行数
}

// Contributor はコントリビューター情報を表す。
type Contributor struct {
	Login         string // ユーザー名
	Contributions int    // コミット数
}

// PullRequest はプルリクエスト情報を表す。
type PullRequest struct {
	Number     int        // PR番号
	Title      string     // タイトル
	Author     string     // 作成者
	HeadBranch string     // ブランチ名（例: "fix/login-bug"）
	CreatedAt  time.Time  // 作成日時
	MergedAt   *time.Time // マージ日時（nilならマージされていない）
	Additions  int        // 追加行数
	Deletions  int        // 削除行数
}

// LeadTime はPRのリードタイム（作成からマージまでの日数）を返す。
// マージされていない場合は-1を返す。
func (pr PullRequest) LeadTime() float64 {
	if pr.MergedAt == nil {
		return -1
	}
	return pr.MergedAt.Sub(pr.CreatedAt).Hours() / 24
}

// IsBugFix はブランチ名からバグ修正PRかどうかを判定する。
func (pr PullRequest) IsBugFix() bool {
	branch := strings.ToLower(pr.HeadBranch)
	return strings.HasPrefix(branch, "fix/") ||
		strings.HasPrefix(branch, "bugfix/") ||
		strings.HasPrefix(branch, "hotfix/")
}

// IsFeature はブランチ名から機能追加PRかどうかを判定する。
func (pr PullRequest) IsFeature() bool {
	branch := strings.ToLower(pr.HeadBranch)
	return strings.HasPrefix(branch, "feature/") ||
		strings.HasPrefix(branch, "feat/")
}

// IsRefactor はブランチ名からリファクタリング系PRかどうかを判定する。
func (pr PullRequest) IsRefactor() bool {
	branch := strings.ToLower(pr.HeadBranch)
	return strings.HasPrefix(branch, "refactor/") ||
		strings.HasPrefix(branch, "chore/") ||
		strings.HasPrefix(branch, "debt/") ||
		strings.HasPrefix(branch, "ci/") ||
		strings.HasPrefix(branch, "docs/")
}

// Dependency は依存パッケージ情報を表す。
type Dependency struct {
	Name        string    // パッケージ名
	Version     string    // 使用中のバージョン
	ReleasedAt  time.Time // そのバージョンのリリース日
	AgeMonths   int       // 何ヶ月前か
	PackageType string    // "npm", "go", etc.
}

// Issue はIssue情報を表す。
type Issue struct {
	Number    int        // Issue番号
	Title     string     // タイトル
	State     string     // "open" or "closed"
	Labels    []string   // ラベル名一覧（"bug", "incident" 等）
	CreatedAt time.Time  // 作成日時
	ClosedAt  *time.Time // クローズ日時（nilならオープン）
}

// Release はリリース情報を表す。
type Release struct {
	ID          int       // リリースID
	TagName     string    // タグ名
	Name        string    // リリース名
	PublishedAt time.Time // 公開日時
}

// Review はPRレビュー情報を表す。
type Review struct {
	ID          int       // レビューID
	Author      string    // レビュアー
	State       string    // "APPROVED", "CHANGES_REQUESTED", "COMMENTED" など
	SubmittedAt time.Time // 投稿日時
}
