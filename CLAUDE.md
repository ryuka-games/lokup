# Lokup

GitHubリポジトリの健康診断ツール。

---

## 協働の原則

> **このセクションの目的**: AI との協働方針を明確にし、より良いシステムを作るため。

### 忖度しない

- 間違っていると思ったら正直に言う
- 「それは良くないと思う」と言える関係
- ユーザーの意見に合わせるだけでなく、より良い提案をする

### ベストプラクティスを常に意識

- 2025-2026年時点の最新のベストプラクティスを優先
- 分からなければネットで調べる（WebSearch を使う）
- 「なんとなく」ではなく、根拠を持って提案する

### 議論を大切にする

- 選択肢があれば提示して議論する
- トレードオフを明確にする
- 決定したら ADR に記録する

### 学びを重視

- なぜその選択をしたか説明する
- ユーザーが後で他の人に説明できるようにする
- ドキュメントに「なぜ」を残す

---

## WHAT

経営者向けサマリーと技術者向け詳細の2層構造で、開発効率とコード健全性を可視化する。

### 技術スタック

- **言語**: Go
- **アーキテクチャ**: Vertical Slice + Clean Architecture + DDD ライト版
- **出力形式**: CLI → HTML レポート
- **認証**: GitHub OAuth（API呼び出し時）

### プロジェクト構造

```
lokup/
├── CLAUDE.md                  # このファイル（AI向け指示書）
├── docs/                      # 設計ドキュメント
│   ├── README.md
│   ├── prd.md
│   ├── requirements.md
│   ├── ui-design.md
│   └── adr/
├── cmd/
│   └── lokup/
│       └── main.go            # エントリーポイント
├── features/                  # 機能別（Vertical Slice）
│   ├── analyze/               # リポジトリ分析
│   ├── report/                # レポート生成
│   └── compare/               # 期間比較
├── domain/                    # ドメインモデル（DDD）
├── infrastructure/            # 外部依存（GitHub API, キャッシュ）
└── shared/                    # 共通ユーティリティ
```

### ドキュメント規約

- **ADR形式**: 技術的な調査・決定は `docs/adr/` に記録
- **命名規則**: `NNN-タイトル.md`（例: `001-github-api.md`）
- **1ファイル1トピック**: 小さく分割して管理

## WHY

経営者に技術投資の必要性を数字で示すため。
「開発効率が落ちてます」「バグ修正に時間取られてます」を可視化し、リファクタリングや技術的負債解消の判断材料を提供する。

## HOW

### ビルド

```bash
go build -o lokup ./cmd/lokup
```

### テスト

```bash
go test ./...
```

### 実行

```bash
./lokup facebook/react --output report.html
```

---

## コーディング規約

> **なぜ規約が必要か**: AIは文脈から「良さそうな」コードを生成するが、プロジェクト固有のルールは知らない。ここに明記することで、AI が一貫したコードを生成できる。

### Go スタイル

- **フォーマッタ**: `gofmt` / `goimports` を必ず使用
  - → 理由: Go コミュニティの標準。議論の余地をなくす
- **命名**: キャメルケース。エクスポートするものは大文字開始
  - → 理由: Go の言語仕様で決まっている
- **エラー処理**: 必ずハンドリング。`_` で握りつぶさない
  - → 理由: 握りつぶすとデバッグ不能になる

```go
// ❌ ダメ
result, _ := doSomething()

// ✅ 良い
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### ファイル構成（Vertical Slice）

- **1機能 = 1フォルダ**: `features/analyze/` 内で完結させる
  - → 理由: AI が1フォルダだけ読めば機能を理解できる
- **依存の方向**: `features/` → `domain/` → OK、逆はNG
  - → 理由: Clean Architecture の依存ルール

### コミットメッセージ

```
<type>: <summary>

<body（任意）>
```

- **type**: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- **言語**: 日本語OK
- → 理由: 後から履歴を見たとき何をしたか分かる

### テスト

- **テーブル駆動テスト**を優先
  - → 理由: Go の標準的なパターン。ケース追加が楽
- **モック**は interface 経由で注入
  - → 理由: テスト時に差し替え可能にする

```go
// テーブル駆動テストの例
func TestCalculateScore(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"zero", 0, 0},
        {"positive", 10, 100},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateScore(tt.input)
            if got != tt.expected {
                t.Errorf("got %d, want %d", got, tt.expected)
            }
        })
    }
}
```

---

## AI支援開発ワークフロー

> **なぜこのワークフローか**: AI は強力だが、監督なしだと暴走する。Plan→Act→Review のサイクルで人間がコントロールを保つ。

### 1. Plan（計画）

- `docs/` にドキュメントを先に書く
- 何を作るか明確にしてから実装

### 2. Act（実装）

- 小さなタスクに分割
- 1タスク = 1コミット目安

### 3. Review（確認）

- AI 出力は必ずレビュー
- テスト実行して動作確認

### 4. Commit（記録）

- 細かくコミット
- 失敗したらロールバックできるように

---

## 参照

### 設計ドキュメント（docs/）

- docs/prd.md — プロダクト要件（コンセプト、ターゲット、診断項目）
- docs/requirements.md — 機能要件詳細
- docs/ui-design.md — 画面設計・導線
- docs/testing-strategy.md — テスト戦略
- docs/adr/ — 技術的な意思決定記録

### AI向けコンテキスト（.project-context/）

> **なぜこのフォルダがあるか**: AI は外部 API の仕様やドメイン知識を知らない。ここに参考情報をまとめておくと、AI に「これ読んで」と渡せる。

- .project-context/github-api-reference.md — GitHub API エンドポイント
- .project-context/domain-model.md — ドメインモデル定義
- .project-context/examples/ — 参考実装
