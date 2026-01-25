# ADR-002: アーキテクチャと技術スタック

## Status

Accepted

## Context

Lokup は GitHub リポジトリの健康診断を行う CLI ツール。以下の要件がある:

- サーバー代をかけたくない（ローカル動作）
- 経営者向けにグラフィカルな出力が必要
- 将来的に Web 対応の可能性を残す
- ベストプラクティスを学びながら開発
- AI支援開発を最大限活用

## Decision

### 言語: Go

理由:
- 学習コストが低く、2日で生産的になれる
- シンプルな構文で AI支援との相性が良い
- シングルバイナリで配布が楽
- Vertical Slice Architecture のテンプレートが充実
- Lokup の規模感に適切（Rust はオーバースペック）

参考: [Rust vs Go in 2026 - Bitfield Consulting](https://bitfieldconsulting.com/posts/rust-vs-go)

### アーキテクチャ: Vertical Slice + Clean Architecture

```
src/
├── cmd/
│   └── lokup/
│       └── main.go           # エントリーポイント
├── features/                  # 機能別（Vertical Slice）
│   ├── analyze/              # リポジトリ分析
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── report/               # レポート生成
│   │   ├── handler.go
│   │   ├── html_generator.go
│   │   └── templates/
│   └── compare/              # 期間比較
│       ├── handler.go
│       └── service.go
├── domain/                    # ドメインモデル（DDD）
│   ├── repository.go         # 集約ルート
│   ├── analysis_result.go
│   ├── risk.go
│   └── value_objects.go
├── infrastructure/            # 外部依存
│   ├── github/               # GitHub API クライアント
│   └── cache/                # キャッシュ
└── shared/                    # 共通ユーティリティ
    └── config/
```

理由:
- 機能ごとにフォルダが完結 → AI がコンテキストを理解しやすい
- 依存関係が局所化 → 変更影響が小さい
- 将来 Web API 追加時は `cmd/api/` を足すだけ

参考: [Go Vertical Slice Template](https://github.com/mehdihadeli/go-vertical-slice-template)

### DDD: ライト版

Strategic DDD（境界づけられたコンテキスト）を意識しつつ、以下のパターンを採用:

- **集約**: Repository, AnalysisResult
- **エンティティ**: Risk
- **値オブジェクト**: Score, DateRange, Severity, RiskType

CQRS は読み取り中心のため軽量版に留める。

### 出力形式: CLI → HTML レポート

```bash
$ lokup facebook/react --output report.html
```

- サーバー不要
- グラフ描画は Chart.js 等を HTML に埋め込み
- そのまま経営者に共有可能

### AI支援開発

[Addy Osmani の 2026年ワークフロー](https://addyosmani.com/blog/ai-coding-workflow/) に準拠:

1. **Plan**: spec.md / CLAUDE.md で要件・設計を先に固める
2. **Act**: 小さなタスクに分けて1つずつ実装
3. **Review**: AI出力は必ずレビュー、テスト実行
4. **Commit**: 細かくコミット（ロールバック可能に）

必須ファイル:
- `CLAUDE.md` — AI向け指示書（作成済み）
- `docs/` — 設計ドキュメント（作成済み）
- `docs/adr/` — 意思決定記録（このファイル）

## Consequences

### Positive

- Go のシンプルさで開発速度が出る
- Vertical Slice で機能追加が容易
- AI支援開発との相性が良い
- ローカル動作でサーバー代ゼロ
- 将来の Web 対応も可能

### Negative

- Go は Rust ほどの実行速度は出ない（Lokup では問題なし）
- Vertical Slice は小規模プロジェクトではやや冗長

### 今後の拡張パス

1. **v1 (MVP)**: CLI + HTML レポート
2. **v2**: Web API 追加（`cmd/api/`）
3. **v3**: フロントエンド追加（React/Vue 等）

## References

- [Addy Osmani - LLM Coding Workflow 2026](https://addyosmani.com/blog/ai-coding-workflow/)
- [Rust vs Go in 2026 - Bitfield Consulting](https://bitfieldconsulting.com/posts/rust-vs-go)
- [Clean Architecture and DDD 2025](https://wojciechowski.app/en/articles/clean-architecture-domain-driven-design-2025)
- [Go Vertical Slice Template](https://github.com/mehdihadeli/go-vertical-slice-template)
