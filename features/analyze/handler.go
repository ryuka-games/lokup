// Package analyze はリポジトリ分析機能を提供する。
//
// Vertical Slice Architecture に従い、この機能に必要なものは
// このパッケージ内で完結する。
//
// 構成:
//   - handler.go  : エントリーポイント（CLI から呼ばれる）
//   - service.go  : ビジネスロジック
//   - repository.go : データ取得（GitHub API 呼び出し）
package analyze

import (
	"context"
	"fmt"
	"time"

	"github.com/ryuka-games/lokup/domain"
)

// Handler は analyze 機能のエントリーポイント。
// CLI からの入力を受け取り、結果を返す。
type Handler struct {
	service *Service
}

// NewHandler は Handler を生成する。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Input は分析の入力パラメータ。
type Input struct {
	Owner string // リポジトリオーナー
	Repo  string // リポジトリ名
	Days  int    // 分析期間（日数）
}

// Handle は分析を実行する。
func (h *Handler) Handle(ctx context.Context, input Input) (*domain.AnalysisResult, error) {
	// 入力バリデーション
	if input.Owner == "" || input.Repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}
	if input.Days <= 0 {
		input.Days = 30 // デフォルト30日
	}

	// 期間の計算
	to := time.Now()
	from := to.AddDate(0, 0, -input.Days)

	// サービス呼び出し
	result, err := h.service.Analyze(ctx, ServiceInput{
		Repository: domain.NewRepository(input.Owner, input.Repo),
		Period:     domain.NewDateRange(from, to),
	})
	if err != nil {
		return nil, fmt.Errorf("analyze failed: %w", err)
	}

	return result, nil
}
