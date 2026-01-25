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
