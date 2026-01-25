# テスト戦略

> **このドキュメントの目的**: テストの方針を明確にし、AI がテストコードを生成するときの指針とする。

---

## なぜテスト戦略が必要か

1. **品質担保**: コードが意図通り動くことを保証
2. **リファクタリング安全網**: 変更しても壊れてないことを確認
3. **AI支援開発の必須要素**: AI 出力を検証する手段

> Addy Osmani の言葉: 「強いテストスイートは AI の生産性を増幅し、微妙なバグを捕まえる」

---

## テストピラミッド

```
        /\
       /  \        E2E テスト（少数）
      /────\       ↑ 実行遅い、メンテ大変
     /      \
    /────────\     統合テスト（中程度）
   /          \
  /────────────\   単体テスト（多数）← ここを厚く
```

### Lokup での適用

| レベル | 対象 | ツール | 量 |
|--------|------|--------|-----|
| **単体テスト** | domain/, features/ の各関数 | go test | 多い |
| **統合テスト** | GitHub API との結合 | go test + モック | 中程度 |
| **E2E テスト** | CLI 全体の動作 | シェルスクリプト | 少数 |

---

## 単体テストの方針

### テーブル駆動テスト

> **なぜ**: Go の標準パターン。ケース追加が楽、可読性が高い。

```go
func TestCalculateRiskLevel(t *testing.T) {
    tests := []struct {
        name           string
        changeCount    int
        expectedLevel  RiskLevel
    }{
        {"低リスク", 5, RiskLow},
        {"中リスク", 15, RiskMedium},
        {"高リスク", 30, RiskHigh},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateRiskLevel(tt.changeCount)
            if got != tt.expectedLevel {
                t.Errorf("got %v, want %v", got, tt.expectedLevel)
            }
        })
    }
}
```

### モック戦略

> **なぜ interface を使うか**: 外部依存（GitHub API）を差し替え可能にするため。

```go
// infrastructure/github/client.go
type GitHubClient interface {
    GetCommits(owner, repo string) ([]Commit, error)
    GetPullRequests(owner, repo string) ([]PullRequest, error)
}

// 本番用
type RealClient struct { ... }

// テスト用
type MockClient struct {
    Commits      []Commit
    PullRequests []PullRequest
    Error        error
}
```

---

## 統合テストの方針

### GitHub API モック

> **なぜモックを使うか**:
> - レート制限を消費しない
> - テストが高速
> - オフラインでも実行可能

```go
// テスト用のHTTPサーバーを立てる
func TestAnalyzeWithMockAPI(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // モックレスポンスを返す
        json.NewEncoder(w).Encode(mockCommits)
    }))
    defer server.Close()

    client := NewGitHubClient(server.URL)
    result, err := Analyze(client, "owner", "repo")
    // ...
}
```

### 実APIテスト（オプション）

```bash
# 環境変数でスキップ可能
SKIP_INTEGRATION=1 go test ./...

# 実APIでテスト（レート制限注意）
GITHUB_TOKEN=xxx go test ./... -tags=integration
```

---

## E2E テストの方針

> **なぜ E2E か**: ユーザー視点で「動く」ことを確認。

```bash
# scripts/e2e-test.sh

#!/bin/bash
set -e

# ビルド
go build -o lokup ./cmd/lokup

# 実行してレポート生成
./lokup facebook/react --output test-report.html

# レポートが生成されたか確認
if [ -f test-report.html ]; then
    echo "✅ E2E テスト成功"
else
    echo "❌ E2E テスト失敗"
    exit 1
fi
```

---

## テストカバレッジ目標

| 層 | 目標 | 理由 |
|----|------|------|
| `domain/` | 90%+ | ビジネスロジックの核心 |
| `features/` | 80%+ | 主要な機能 |
| `infrastructure/` | 70%+ | 外部依存はモックでカバー |
| `cmd/` | 50%+ | エントリーポイントは薄く |

### カバレッジ確認コマンド

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # ブラウザで確認
```

---

## テスト命名規則

```
Test<関数名>_<シナリオ>

例:
TestCalculateRiskLevel_HighChangeCount
TestAnalyze_EmptyRepository
TestGenerateReport_WithComparison
```

> **なぜこの命名か**: 失敗時にどこが壊れたか一目で分かる。

---

## CI での実行

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -race -cover ./...
```

---

## まとめ

| 原則 | 内容 |
|------|------|
| テーブル駆動 | Go 標準パターンに従う |
| interface でモック | 外部依存を差し替え可能に |
| 単体テスト厚め | ピラミッドの土台を固める |
| カバレッジ確認 | 定期的に計測 |
