# Lokup ドキュメント

## 目次

| ドキュメント | 内容 |
|-------------|------|
| [prd.md](./prd.md) | プロダクト要件（コンセプト、ターゲット、診断項目） |
| [requirements.md](./requirements.md) | 機能要件詳細 |
| [ui-design.md](./ui-design.md) | 画面設計・導線 |
| [testing-strategy.md](./testing-strategy.md) | テスト戦略 |
| [architecture.md](./architecture.md) | アーキテクチャ（認証フロー・分析フロー・ファイル構成） |
| [metrics.md](./metrics.md) | メトリクス仕様（DORA・スコア・リスク検出） |

## ADR（意思決定記録）

| ADR | 内容 | Status |
|-----|------|--------|
| [001-github-api.md](./adr/001-github-api.md) | GitHub REST API を使用する | Accepted |
| [002-architecture.md](./adr/002-architecture.md) | Go + Vertical Slice + DDD | Accepted |
| [003-development-environment.md](./adr/003-development-environment.md) | Scoop + Go 環境構築 | Accepted |

## ステータス

- [x] PRD作成
- [x] GitHub API 調査・決定
- [x] 技術スタック決定（Go）
- [x] アーキテクチャ決定（Vertical Slice + DDD）
- [x] テスト戦略策定
- [x] AI支援開発環境整備（CLAUDE.md, .project-context/）
- [x] Go プロジェクト初期化
- [x] フォルダ構造・基本コード作成
- [x] CLI 引数処理（flag パッケージ）
- [x] GitHub API 実装（認証 + データ取得）
- [x] HTML レポート出力（3段階開示 + Chart.js）
- [x] ユニットテスト（domain, features/analyze, features/report, cmd）
- [x] GitHub Actions CI（テスト + golangci-lint）
- [x] メトリクス仕様ドキュメント

---

## 今後のタスク

- [ ] レート制限の監視（GitHub API）
- [ ] 日別コミット推移グラフ（レポート追加）
- [ ] サンプルレポートのスクリーンショット（README用）
