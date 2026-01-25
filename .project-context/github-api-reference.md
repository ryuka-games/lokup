# GitHub API リファレンス

> **このドキュメントの目的**: Lokup が使用する GitHub API エンドポイントをまとめる。AI がコードを生成するときの参照用。

---

## 認証

```
Authorization: Bearer <token>
```

- OAuth トークンまたは Personal Access Token
- 認証あり: 5,000 リクエスト/時
- 認証なし: 60 リクエスト/時

### レート制限の確認

レスポンスヘッダーに含まれる:

```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1234567890  # Unix timestamp
```

---

## 使用するエンドポイント

### 1. コミット一覧

**用途**: 変更集中リスク、深夜労働の分析

```
GET /repos/{owner}/{repo}/commits
```

**パラメータ**:
- `per_page`: 1ページの件数（max 100）
- `since`: この日時以降（ISO 8601）
- `until`: この日時まで（ISO 8601）

**レスポンス例**:
```json
[
  {
    "sha": "abc123...",
    "commit": {
      "author": {
        "name": "John Doe",
        "email": "john@example.com",
        "date": "2026-01-24T10:30:00Z"
      },
      "message": "feat: add new feature"
    }
  }
]
```

---

### 2. コミット詳細

**用途**: 変更ファイル、追加/削除行数

```
GET /repos/{owner}/{repo}/commits/{sha}
```

**レスポンス例**:
```json
{
  "sha": "abc123...",
  "files": [
    {
      "filename": "src/main.go",
      "status": "modified",
      "additions": 10,
      "deletions": 5
    }
  ]
}
```

---

### 3. プルリクエスト一覧

**用途**: リリースまでの日数（リードタイム）

```
GET /repos/{owner}/{repo}/pulls
```

**パラメータ**:
- `state`: `open`, `closed`, `all`
- `per_page`: 1ページの件数

**レスポンス例**:
```json
[
  {
    "number": 123,
    "title": "Add feature",
    "created_at": "2026-01-20T10:00:00Z",
    "merged_at": "2026-01-21T15:00:00Z",
    "user": {
      "login": "contributor"
    }
  }
]
```

**リードタイム計算**:
```go
leadTime := mergedAt.Sub(createdAt)
```

---

### 4. ファイル内容

**用途**: 巨大ファイル検出、依存パッケージ確認

```
GET /repos/{owner}/{repo}/contents/{path}
```

**レスポンス例**:
```json
{
  "name": "package.json",
  "path": "package.json",
  "size": 1234,
  "content": "base64エンコードされた内容",
  "encoding": "base64"
}
```

**デコード**:
```go
decoded, _ := base64.StdEncoding.DecodeString(content)
lines := strings.Split(string(decoded), "\n")
lineCount := len(lines)
```

---

### 5. コントリビューター一覧

**用途**: 属人化の検出

```
GET /repos/{owner}/{repo}/contributors
```

**レスポンス例**:
```json
[
  {
    "login": "user1",
    "contributions": 500
  },
  {
    "login": "user2",
    "contributions": 200
  }
]
```

**属人化判定**:
```go
total := sum(all contributions)
topContributor := contributors[0].contributions
ratio := float64(topContributor) / float64(total)
if ratio > 0.8 {
    // 属人化リスク
}
```

---

## Go での実装パターン

```go
package github

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

func NewClient(token string) *Client {
    return &Client{
        baseURL:    "https://api.github.com",
        token:      token,
        httpClient: &http.Client{},
    }
}

func (c *Client) GetCommits(owner, repo string) ([]Commit, error) {
    url := fmt.Sprintf("%s/repos/%s/%s/commits", c.baseURL, owner, repo)

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var commits []Commit
    json.NewDecoder(resp.Body).Decode(&commits)
    return commits, nil
}
```

---

## レート制限対策

```go
func (c *Client) checkRateLimit(resp *http.Response) {
    remaining := resp.Header.Get("X-RateLimit-Remaining")
    if remaining == "0" {
        reset := resp.Header.Get("X-RateLimit-Reset")
        // 警告を出す or 待機
    }
}
```

---

## 参考リンク

- [GitHub REST API ドキュメント](https://docs.github.com/en/rest)
- [レート制限について](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api)
