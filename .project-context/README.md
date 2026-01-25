# .project-context

> **このフォルダの目的**: AI に追加のコンテキストを提供する。

---

## なぜこのフォルダが必要か

AI（Claude, Copilot 等）は以下を知らない:

1. **外部 API の仕様** — GitHub API のエンドポイント、レスポンス形式
2. **ドメイン知識** — Lokup 固有の概念、モデル
3. **参考実装** — 「こういう感じで書いて」の例

このフォルダにまとめておくと、AI に「これ読んで」と渡せる。

---

## フォルダ構成

```
.project-context/
├── README.md                  # このファイル
├── github-api-reference.md    # 使用する GitHub API エンドポイント
├── domain-model.md            # ドメインモデルの定義
└── examples/                  # 参考実装・サンプルコード
    └── vertical-slice-example.go
```

---

## 使い方

### Claude Code の場合

会話の中で:
```
.project-context/github-api-reference.md を読んで、
GitHub API からコミット履歴を取得する関数を書いて
```

### CLAUDE.md から参照

```markdown
## 参照
- .project-context/domain-model.md — ドメインモデル定義
```

---

## 更新タイミング

- API 仕様が変わったとき
- ドメインモデルを変更したとき
- 新しい参考実装ができたとき
