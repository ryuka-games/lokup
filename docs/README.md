# Lokup ドキュメント

## 目次

| ドキュメント | 内容 |
|-------------|------|
| [prd.md](./prd.md) | プロダクト要件（コンセプト、ターゲット、診断項目） |
| [requirements.md](./requirements.md) | 機能要件詳細 |
| [ui-design.md](./ui-design.md) | 画面設計・導線 |
| [testing-strategy.md](./testing-strategy.md) | テスト戦略 |

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
- [ ] 機能要件詳細
- [ ] 画面設計

---

## 次回タスク

### 0. VSCode セットアップ
- [ ] Go 拡張インストール（`golang.go`）
- [ ] Vim 拡張インストール（`vscodevim.vim`）
- [ ] gopls インストール（初回起動時に聞かれる）
- [ ] lokup プロジェクトを VSCode で開いて動作確認

### 1. CLI 引数処理 ✅
- [x] `lokup facebook/react` の形式で引数を受け取る
- [x] `--output report.html` オプション
- [x] `--days 30` オプション（分析期間）
- [x] 標準の `flag` パッケージを使用（YAGNI: cobra は不要）
- [x] ユニットテスト追加

### 2. GitHub API 実際に叩く ✅
- [x] features/analyze を実機能として動かす
- [x] 認証トークンの扱い（環境変数 `GITHUB_TOKEN`）
- [ ] レート制限の監視（後で追加）

### 3. HTML レポート出力 ✅
- [x] features/report を実装
- [x] テンプレートエンジン（`html/template`）
- [x] Chart.js でグラフ描画（スコア比較バーチャート）
- [x] 経営サマリー + 詳細の2層構造
- [ ] 日別コミット推移グラフ（追加予定）
