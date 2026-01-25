// Package examples は Vertical Slice Architecture の実装例を示す。
// このファイルは参考実装であり、実際のコードではない。
//
// ## なぜこの構成か
//
// Vertical Slice では「1機能 = 1フォルダ」で完結させる。
// これにより:
// - AI が1フォルダ読めば機能を理解できる
// - 変更の影響範囲が局所化する
// - 機能追加時に他の機能を壊しにくい
//
// ## フォルダ構成例
//
// features/
// └── analyze/
//     ├── handler.go      ← エントリーポイント（CLI から呼ばれる）
//     ├── service.go      ← ビジネスロジック
//     ├── repository.go   ← データ取得（GitHub API 呼び出し）
//     └── types.go        ← この機能固有の型
//
package examples

import (
	"context"
	"fmt"
	"time"
)

// ============================================================
// handler.go - エントリーポイント
// ============================================================

// Handler は CLI からの入力を受け取り、結果を返す。
// 責務: 入力のバリデーション、サービスの呼び出し、結果の整形
type Handler struct {
	service *AnalyzeService
}

func NewHandler(service *AnalyzeService) *Handler {
	return &Handler{service: service}
}

// Handle は CLI のメイン処理。
// なぜ context を受け取るか: キャンセル処理、タイムアウト制御のため
func (h *Handler) Handle(ctx context.Context, owner, repo string, days int) (*AnalyzeOutput, error) {
	// 1. 入力バリデーション
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	// 2. 期間の計算
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	// 3. サービス呼び出し
	result, err := h.service.Analyze(ctx, AnalyzeInput{
		Owner: owner,
		Repo:  repo,
		From:  from,
		To:    to,
	})
	if err != nil {
		return nil, fmt.Errorf("analyze failed: %w", err)
	}

	return result, nil
}

// ============================================================
// service.go - ビジネスロジック
// ============================================================

// AnalyzeService は分析のビジネスロジックを担当。
// 責務: リスク検出、スコア計算、結果の組み立て
type AnalyzeService struct {
	repo AnalyzeRepository
}

func NewAnalyzeService(repo AnalyzeRepository) *AnalyzeService {
	return &AnalyzeService{repo: repo}
}

// Analyze はリポジトリを分析する。
// なぜ AnalyzeInput を使うか: 引数が増えても signature が変わらない
func (s *AnalyzeService) Analyze(ctx context.Context, input AnalyzeInput) (*AnalyzeOutput, error) {
	// 1. データ取得（Repository 経由）
	commits, err := s.repo.GetCommits(ctx, input.Owner, input.Repo, input.From, input.To)
	if err != nil {
		return nil, err
	}

	// 2. リスク検出
	risks := s.detectRisks(commits)

	// 3. スコア計算
	score := s.calculateScore(risks)

	// 4. 結果を組み立て
	return &AnalyzeOutput{
		Score: score,
		Risks: risks,
	}, nil
}

func (s *AnalyzeService) detectRisks(commits []Commit) []Risk {
	var risks []Risk

	// 変更集中リスクの検出
	fileChanges := make(map[string]int)
	for _, c := range commits {
		for _, f := range c.Files {
			fileChanges[f]++
		}
	}

	for file, count := range fileChanges {
		if count >= 10 {
			severity := SeverityMedium
			if count >= 20 {
				severity = SeverityHigh
			}
			risks = append(risks, Risk{
				Type:     RiskTypeChangeConcentration,
				Severity: severity,
				Target:   file,
				Value:    count,
			})
		}
	}

	return risks
}

func (s *AnalyzeService) calculateScore(risks []Risk) int {
	score := 100

	for _, r := range risks {
		switch r.Severity {
		case SeverityHigh:
			score -= 10
		case SeverityMedium:
			score -= 5
		case SeverityLow:
			score -= 2
		}
	}

	if score < 0 {
		score = 0
	}
	return score
}

// ============================================================
// repository.go - データ取得
// ============================================================

// AnalyzeRepository はデータ取得のインターフェース。
// なぜ interface か: テスト時にモックに差し替えるため
type AnalyzeRepository interface {
	GetCommits(ctx context.Context, owner, repo string, from, to time.Time) ([]Commit, error)
}

// GitHubAnalyzeRepository は GitHub API を使った実装。
type GitHubAnalyzeRepository struct {
	// client github.Client  // 実際には infrastructure/github から注入
}

func (r *GitHubAnalyzeRepository) GetCommits(ctx context.Context, owner, repo string, from, to time.Time) ([]Commit, error) {
	// GitHub API を呼び出し
	// ...
	return nil, nil
}

// MockAnalyzeRepository はテスト用のモック。
type MockAnalyzeRepository struct {
	Commits []Commit
	Error   error
}

func (r *MockAnalyzeRepository) GetCommits(ctx context.Context, owner, repo string, from, to time.Time) ([]Commit, error) {
	return r.Commits, r.Error
}

// ============================================================
// types.go - この機能固有の型
// ============================================================

// AnalyzeInput は Analyze の入力。
type AnalyzeInput struct {
	Owner string
	Repo  string
	From  time.Time
	To    time.Time
}

// AnalyzeOutput は Analyze の出力。
type AnalyzeOutput struct {
	Score int
	Risks []Risk
}

// Commit はコミット情報（簡略版）。
type Commit struct {
	SHA     string
	Author  string
	Date    time.Time
	Message string
	Files   []string
}

// Risk はリスク情報。
type Risk struct {
	Type     RiskType
	Severity Severity
	Target   string
	Value    int
}

// RiskType はリスクの種類。
type RiskType string

const (
	RiskTypeChangeConcentration RiskType = "change_concentration"
)

// Severity は重大度。
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
)
