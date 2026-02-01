# Lokup レポート AI 分析プロンプト

> **他のAI（ChatGPT, Gemini等）でも使える汎用プロンプトです。**
> Claude Code の場合は `/analyze-report report.html` を使ってください。

## 手順

1. 指定された `report.html` を読み込む
2. HTML 内の以下のデータを読み取る:
   - 総合スコア・グレード
   - カテゴリ別スコア（開発速度、コード品質、技術的負債、チーム健全性）
   - 検出されたリスク一覧
   - DORA メトリクス（デプロイ頻度、変更失敗率、MTTR）
   - 投資比率（Feature / BugFix / Refactor / Other）
   - トレンド（前期比較）
3. 以下の構成で分析コメントを作成する:
   - **ヒーローインサイト**: 最も重要な発見1つ + 具体的アクション（1行ずつ）
   - **インサイトカード**: 良い点と課題をカード形式で（各1-3個）
   - **優先アクション**: 今すぐ取り組むべきこと（1-3個）
   - **補足**: データの制約や注意事項（折りたたみ式）
4. `<div id="ai-comments">` の中身を以下のHTMLで置換する

## 出力フォーマット

```html
<!-- ヒーローインサイト: 最も重要な1つだけ -->
<div class="ai-hero-insight">
    <div class="ai-hero-label">最も重要な発見</div>
    <div class="ai-hero-text">（1文で核心を突く）</div>
    <div class="ai-hero-action">→ （1文で具体的アクション）</div>
</div>

<!-- インサイトカード: good=良い点, warn=注意, bad=深刻 -->
<div class="ai-insight-grid">
    <div class="ai-insight-card good">
        <div class="ai-card-icon">✅</div>
        <div class="ai-card-title">指標名</div>
        <div class="ai-card-value">数値</div>
        <div class="ai-card-note">1-2文の解説</div>
    </div>
    <!-- カード3-6個 -->
</div>

<!-- 優先アクション: 最大3個 -->
<div class="ai-actions">
    <h3>優先アクション</h3>
    <div class="ai-action-item">
        <span class="ai-action-num">1</span>
        <div class="ai-action-text"><strong>アクション名</strong>: 具体的な内容</div>
    </div>
</div>

<!-- 補足: 折りたたみ -->
<details class="ai-note-toggle">
    <summary>データの補足・注意事項</summary>
    <div class="ai-note-content">注意事項をここに</div>
</details>
```

## 注意事項

- 既存の HTML 構造を壊さないこと
- `<div id="ai-comments">` の中身だけを変更すること
- 日本語で記述すること
- 具体的な数値を引用して根拠を示すこと
- ヒーローインサイトは1つだけに絞る
- インサイトカードは1カード = 1指標。長文を書かない
