# ドメインモデル

> **このドキュメントの目的**: Lokup のドメインモデルを定義する。AI がコードを生成するときの設計図。

---

## なぜドメインモデルを先に定義するか

1. **共通言語**: 「リスク」「スコア」の意味を統一
2. **設計の指針**: どんな型・構造体を作るべきか明確に
3. **AI への指示**: 「Risk 型を使って」と言える

---

## 集約（Aggregate）

### AnalysisResult（分析結果）

> **集約ルート**: 診断結果全体を束ねる

```go
type AnalysisResult struct {
    Repository      Repository      // 対象リポジトリ
    Period          DateRange       // 分析期間
    EfficiencyScore Score           // 経営向けスコア
    Risks           []Risk          // 検出されたリスク
    Metrics         Metrics         // 各種メトリクス
    GeneratedAt     time.Time       // 生成日時
}
```

---

## エンティティ（Entity）

### Risk（リスク）

> **識別子を持つ**: 同じ種類でも別のファイルなら別のリスク

```go
type Risk struct {
    ID          string      // 一意識別子
    Type        RiskType    // リスクの種類
    Severity    Severity    // 重大度
    Target      string      // 対象（ファイル名等）
    Description string      // 説明
    Value       int         // 数値（変更回数、行数等）
    Threshold   int         // 閾値
}
```

---

## 値オブジェクト（Value Object）

> **不変**: 一度作ったら変更しない

### Repository

```go
type Repository struct {
    Owner string  // 例: "facebook"
    Name  string  // 例: "react"
}

func (r Repository) FullName() string {
    return r.Owner + "/" + r.Name
}
```

### DateRange

```go
type DateRange struct {
    From time.Time
    To   time.Time
}

func (d DateRange) Days() int {
    return int(d.To.Sub(d.From).Hours() / 24)
}
```

### Score

```go
type Score struct {
    Value int  // 0-100
}

func NewScore(value int) Score {
    if value < 0 {
        value = 0
    }
    if value > 100 {
        value = 100
    }
    return Score{Value: value}
}

func (s Score) Grade() string {
    switch {
    case s.Value >= 80:
        return "A"
    case s.Value >= 60:
        return "B"
    case s.Value >= 40:
        return "C"
    default:
        return "D"
    }
}
```

### Severity（重大度）

```go
type Severity int

const (
    SeverityLow    Severity = iota  // 🟢
    SeverityMedium                   // 🟡
    SeverityHigh                     // 🔴
)

func (s Severity) Emoji() string {
    switch s {
    case SeverityLow:
        return "🟢"
    case SeverityMedium:
        return "🟡"
    case SeverityHigh:
        return "🔴"
    default:
        return "⚪"
    }
}
```

### RiskType（リスク種別）

```go
type RiskType string

const (
    RiskTypeChangeConcentration RiskType = "change_concentration"  // 変更集中
    RiskTypeLargeFile           RiskType = "large_file"            // 巨大ファイル
    RiskTypeAbandoned           RiskType = "abandoned"             // 放置ファイル
    RiskTypeOwnership           RiskType = "ownership"             // 属人化
    RiskTypeOutdatedDeps        RiskType = "outdated_deps"         // 依存の古さ
    RiskTypeLateNight           RiskType = "late_night"            // 深夜労働
)

func (r RiskType) DisplayName() string {
    names := map[RiskType]string{
        RiskTypeChangeConcentration: "変更集中リスク",
        RiskTypeLargeFile:           "巨大ファイル",
        RiskTypeAbandoned:           "放置ファイル",
        RiskTypeOwnership:           "属人化",
        RiskTypeOutdatedDeps:        "依存の古さ",
        RiskTypeLateNight:           "深夜労働",
    }
    return names[r]
}
```

---

## Metrics（メトリクス）

```go
type Metrics struct {
    // 経営向け
    FeatureAdditionRate float64  // 機能追加速度（コミット/日）
    BugFixRatio         float64  // バグ修正の割合（%）
    ReworkRate          float64  // 手戻り率（%）
    LeadTime            float64  // PR作成→マージの平均日数

    // 技術向け
    TotalCommits        int
    TotalFiles          int
    TotalContributors   int
    LateNightCommitRate float64  // 深夜コミット率（%）
}
```

---

## ドメインサービス

### Analyzer（分析器）

```go
type Analyzer interface {
    Analyze(repo Repository, period DateRange) (*AnalysisResult, error)
}
```

### RiskDetector（リスク検出器）

```go
type RiskDetector interface {
    Detect(commits []Commit, files []File) []Risk
}
```

---

## 閾値の設定

```go
type Thresholds struct {
    ChangeConcentration struct {
        WarningCount  int  // 警告: 30日で10回以上
        CriticalCount int  // 危険: 30日で20回以上
    }
    LargeFile struct {
        WarningLines  int  // 警告: 500行超
        CriticalLines int  // 危険: 1000行超
    }
    Ownership struct {
        WarningRatio float64  // 警告: 1人が80%以上
    }
    LateNight struct {
        WarningRatio float64  // 警告: 22時〜5時が30%以上
    }
}

var DefaultThresholds = Thresholds{
    ChangeConcentration: struct {
        WarningCount  int
        CriticalCount int
    }{10, 20},
    LargeFile: struct {
        WarningLines  int
        CriticalLines int
    }{500, 1000},
    Ownership: struct {
        WarningRatio float64
    }{0.8},
    LateNight: struct {
        WarningRatio float64
    }{0.3},
}
```

---

## 関係図

```
┌─────────────────────────────────────────────┐
│              AnalysisResult                 │
│  ┌─────────────┐  ┌─────────────┐          │
│  │ Repository  │  │  DateRange  │          │
│  └─────────────┘  └─────────────┘          │
│  ┌─────────────┐  ┌─────────────┐          │
│  │    Score    │  │   Metrics   │          │
│  └─────────────┘  └─────────────┘          │
│  ┌──────────────────────────────┐          │
│  │         []Risk               │          │
│  │  ┌────────┐ ┌──────────┐    │          │
│  │  │RiskType│ │ Severity │    │          │
│  │  └────────┘ └──────────┘    │          │
│  └──────────────────────────────┘          │
└─────────────────────────────────────────────┘
```
