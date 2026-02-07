# 機能要件詳細

## 参照

機能要件の詳細（計算ロジック・閾値・スコアリング）は以下のドキュメントに記載:

- [metrics.md](./metrics.md) — メトリクス仕様（DORA・スコア・リスク検出・投資比率）
- [prd.md](./prd.md) — 診断項目の概要とコンセプト
- [architecture.md](./architecture.md) — 分析フロー・ファイル構成

## 実装済み機能

### 入力
- CLI で `lokup owner/repo` 形式で指定
- `--days` オプションで分析期間を指定（デフォルト30日）
- `--output` オプションで出力ファイルパスを指定（デフォルト report.html）
- GitHub 認証: `GITHUB_TOKEN` 環境変数 or `gh auth token`

### 分析
- 4カテゴリスコア: Velocity, Quality, TechDebt, Health（各100点、リスクで減点）
- 14種類のリスク検出（変更集中、属人化、深夜労働、巨大ファイル、古い依存、DORA系等）
- DORAメトリクス: デプロイ頻度、変更失敗率、MTTR
- 投資比率: Feature/BugFix/Refactor PR分類
- トレンド比較（前期 vs 今期）
- 依存チェック: npm, Go, Python, .NET

### 出力
- CLI にサマリー表示
- HTML レポート: 経営者向けサマリー + 技術者向け詳細の3段階開示
- Chart.js によるスコアバーチャート
