---
name: analyze-report
description: Lokupが生成したHTMLレポートを読み取り、AIによる分析コメントを追記する。レポート分析、健康診断コメント追加時に使用。
---

# Lokup レポート AI 分析

$ARGUMENTS のHTMLレポートを読み取り、AI分析コメントを追記してください。

引数が空の場合は `report.html` を対象にしてください。

## 手順

1. 対象の HTML ファイルを読み込む（大きい場合は分割して読む）
2. HTML 内のデータを読み取る:
   - 総合スコア・グレード（`overall-grade` クラスの要素）
   - カテゴリ別スコア（`category-card` クラスの要素4つ）
   - 各メトリクスの値（`metric-detail` 内の数値）
   - 検出されたリスク一覧（`risk-item` クラスの要素）
   - DORA メトリクス（`dora-badge` クラスの要素）
   - 投資比率（Feature/BugFix/Refactor/Other の件数）
   - トレンド（`trendsData` の JSON）
3. 以下の構成で分析コメントを作成する:
   - **ヒーローインサイト**: 最も重要な発見1つ + 具体的アクション（1行ずつ）
   - **インサイトカード**: 良い点（1-3個）と課題（1-3個）をカード形式で
   - **優先アクション**: 今すぐ取り組むべきこと（1-3個）を番号付きで
   - **補足**: データの制約や注意事項（折りたたみ式）
4. `<div id="ai-comments">` の中身を分析結果の HTML で置換する

## 出力フォーマット

`<div id="ai-comments">` の中身を以下の形式で **そのまま** 置換すること。
CSSクラスは `features/report/template.html` に定義済みなのでそのまま使う。

```html
<!-- ヒーローインサイト: 最も重要な1つだけ。短く。 -->
<div class="ai-hero-insight">
    <div class="ai-hero-label">最も重要な発見</div>
    <div class="ai-hero-text">（1文で核心を突く。例: 「変更失敗率66.7%がリリース品質を大きく損なっています」）</div>
    <div class="ai-hero-action">→ （1文で具体的アクション）</div>
</div>

<!-- インサイトカード: 良い点 = good、注意 = warn、深刻 = bad -->
<div class="ai-insight-grid">
    <div class="ai-insight-card good">
        <div class="ai-card-icon">✅</div>
        <div class="ai-card-title">（指標名）</div>
        <div class="ai-card-value">（数値）</div>
        <div class="ai-card-note">（1-2文の解説）</div>
    </div>
    <div class="ai-insight-card warn">
        <div class="ai-card-icon">⚠️</div>
        <div class="ai-card-title">（指標名）</div>
        <div class="ai-card-value">（数値）</div>
        <div class="ai-card-note">（1-2文の解説）</div>
    </div>
    <div class="ai-insight-card bad">
        <div class="ai-card-icon">🔴</div>
        <div class="ai-card-title">（指標名）</div>
        <div class="ai-card-value">（数値）</div>
        <div class="ai-card-note">（1-2文の解説）</div>
    </div>
    <!-- カードは3-6個。良い点と課題をバランスよく。 -->
</div>

<!-- 優先アクション: 最大3個。番号付き。 -->
<div class="ai-actions">
    <h3>優先アクション</h3>
    <div class="ai-action-item">
        <span class="ai-action-num">1</span>
        <div class="ai-action-text"><strong>（アクション名）</strong>: （具体的な内容を1-2文で）</div>
    </div>
    <div class="ai-action-item">
        <span class="ai-action-num">2</span>
        <div class="ai-action-text"><strong>（アクション名）</strong>: （具体的な内容を1-2文で）</div>
    </div>
</div>

<!-- 補足: 折りたたみ。データの制約や注意事項。 -->
<details class="ai-note-toggle">
    <summary>データの補足・注意事項</summary>
    <div class="ai-note-content">
        （箇条書きまたは短い段落で。例: 投資比率の分類が効かない理由、API上限の影響など）
    </div>
</details>
```

## ルール

- 既存の HTML 構造を壊さないこと
- `<div id="ai-comments">` の中身だけを編集すること
- 日本語で記述すること
- 具体的な数値を引用して分析の根拠を示すこと
- 曖昧な表現を避け、各カードの文は短く具体的に
- ヒーローインサイトは **1つだけ** に絞ること（最重要の発見）
- インサイトカードは **1カード = 1指標**。長文を書かない
- 補足は details タグで折りたたむこと
