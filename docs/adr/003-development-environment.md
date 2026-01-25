# ADR-003: 開発環境のセットアップ

## Status

Accepted

## Context

Windows 環境で Go の開発環境を構築する必要がある。

### 要件

- バージョン管理がしやすい
- アップデートが楽
- アンインストールがクリーン
- 将来的に他のツールも管理できる

### 選択肢

| 方法 | 特徴 |
|------|------|
| 公式インストーラー | 手動管理、アンインストール面倒 |
| winget | Windows 標準、シンプル |
| Chocolatey | 歴史あり、管理者権限必要 |
| Scoop | 開発者向け、管理者権限不要、クリーン |
| gvm | Go 専用のバージョンマネージャー |

## Decision

**Scoop** を使用する。

理由:
- 管理者権限不要（UAC ポップアップなし）
- インストール先が `~/scoop/` に統一されクリーン
- アンインストールも残骸が残らない
- Go 以外のツール（Git, Node.js 等）も同じ方法で管理できる
- Linux/Mac の Homebrew に近い感覚で使える
- 2025-2026 年の Windows 開発者のベストプラクティス

## セットアップ手順

### 1. Scoop インストール

```powershell
# PowerShell で実行
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
```

### 2. Go インストール

```powershell
scoop install go
```

### 3. 確認

```powershell
go version
# go version go1.22.x windows/amd64
```

### 4. アップデート（将来）

```powershell
scoop update go
```

## Consequences

### Positive

- 環境構築が再現可能（手順が明確）
- 複数の開発ツールを統一的に管理
- チームメンバーにも同じ手順を共有できる

### Negative

- Scoop 自体を先にインストールする必要がある
- Windows 専用（Mac/Linux では使えない）

## エディタ

**選択**: VSCode + Go 拡張 + Vim 拡張

### なぜ GoLand ではないか

- 年$199 は「金で解決」感がある
- 無料ツールで工夫する方が学びになる
- gopls が優秀になり、VSCode でも十分な体験

### なぜ Neovim ではないか（今は）

- Go 学習と Neovim 学習を同時にやると辛い
- まず Go に集中すべき

### 移行パス

1. **Phase 1**: VSCode + Go 拡張 + Vim 拡張
   - Go に集中しつつ Vim キーバインドに慣れる
2. **Phase 2**: Neovim + gopls に移行（慣れてから）

### セットアップ

VSCode 拡張機能:
- `golang.go` — Go 公式拡張
- `vscodevim.vim` — Vim キーバインド

## References

- [Scoop 公式](https://scoop.sh/)
- [Chocolatey vs Scoop vs Winget 比較](https://www.xda-developers.com/chocolatey-vs-winget-vs-scoop/)
- [GoLand vs VSCode 比較](https://medium.com/codex/an-in-depth-comparison-goland-vs-visual-studio-code-for-go-development-b7cda8e8918b)
