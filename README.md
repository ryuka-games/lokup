# Lokup

GitHub リポジトリの健康診断ツール。コミット履歴やPR、Issue、リリース情報を分析し、開発チームの状態を可視化します。

## 特徴

- **総合スコア**: 4カテゴリの平均スコアとグレード（A〜D）で一目でわかる健康状態
- **4カテゴリ評価**: 開発速度・コード品質・技術的負債・チーム健全性を100点満点で評価
- **DORA Four Keys**: デプロイ頻度・変更失敗率・MTTRをDORAレーティング（Elite/High/Medium/Low）で表示
- **リスク検出**: 深夜労働、属人化、変更集中、巨大ファイル、古い依存など14種類のリスクを自動検出
- **投資比率**: PR分類（Feature/BugFix/Refactor/Other）による開発リソースの配分を可視化
- **トレンド比較**: 前期比の変化率（↑↓→）で改善・悪化を表示
- **3段階開示レポート**: 総合グレード → カテゴリカード → 展開式詳細の段階的開示で、経営者にも技術者にも読みやすい
- **AI分析**: 生成AIによるレポート分析コメントの追記に対応（Claude Code スキル / 汎用プロンプト）

## 使い方

```bash
# 基本的な使い方
lokup facebook/react

# 分析期間を指定（デフォルト: 30日）
lokup facebook/react --days 90

# 出力ファイルを指定
lokup facebook/react --output my-report.html
```

### GitHub 認証（必須）

GitHub APIを使用するため、認証が必要です。

**方法1: GitHub CLI（推奨）**

```bash
# 初回のみ
gh auth login

# 以降、Lokup は自動的にトークンを取得します
lokup facebook/react
```

**方法2: 環境変数**

```bash
export GITHUB_TOKEN=ghp_xxxxx...
lokup facebook/react
```

トークンの優先順位: `GITHUB_TOKEN` 環境変数 → `gh auth token`

## レポート構造

レポートは3段階の段階的開示（Progressive Disclosure）で構成されています。

```
Level 1: 総合グレード（A〜D）+ 一行診断
Level 2: カテゴリカード（スコア + グレードのみ）
         検出されたリスク一覧
Level 3: カテゴリ詳細（展開式）
         トレンド（展開式）
         AI分析コメント
```

## AI分析

生成AIにレポートを読ませて、分析コメントを追記できます。

**Claude Code の場合:**

```bash
# スキルを使って分析コメントを追記
/analyze-report report.html
```

**他のAI（ChatGPT, Gemini等）の場合:**

`scripts/analyze-report-prompt.md` のプロンプトを使用してください。
レポートHTMLの内容と一緒にAIに渡すと、同じ形式で分析コメントを生成できます。

## 診断項目

### 開発速度 (Velocity)
- PRリードタイム（PR作成からマージまでの平均日数）
- コミット頻度（1日あたりの平均コミット数）
- レビュー待ち時間（PR作成から最初のレビューまで）
- デプロイ頻度（DORA: リリース/月）
- MTTR（DORA: バグIssueの平均復旧時間）

### コード品質 (Quality)
- バグ修正割合（ブランチ名から自動分類）
- 変更集中（ホットスポットの検出）
- PRサイズ（平均変更行数）
- Issueクローズ率
- 変更失敗率（DORA: 障害数/デプロイ数）
- コードチャーン（Revertコミット率）

### 技術的負債 (Tech Debt)
- 巨大ファイル（50KB/100KB超）
- 古い依存パッケージ（npm, Go, Python, NuGet対応）
- 機能投資比率（Feature PRの割合）

### チーム健全性 (Health)
- 深夜コミット率（22時〜5時）
- 属人化リスク（コミットの偏り）

詳細な仕様は [docs/metrics.md](docs/metrics.md) を参照。

## 技術スタック

- **言語**: Go
- **アーキテクチャ**: Vertical Slice + Clean Architecture
- **API**: GitHub REST API
- **レポート**: html/template + Chart.js

## インストール

```bash
go install github.com/ryuka-games/lokup/cmd/lokup@latest
```

または、ソースからビルド：

```bash
git clone https://github.com/ryuka-games/lokup.git
cd lokup
go build -o lokup ./cmd/lokup
```

## 開発

```bash
# テスト実行
go test ./...

# ビルド
go build -o lokup ./cmd/lokup

# 実行
./lokup facebook/react --days 7
```

## ライセンス

MIT
