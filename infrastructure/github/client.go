// Package github は GitHub API クライアントを提供する。
//
// このパッケージは infrastructure 層に属し、
// features/analyze の Repository インターフェースを実装する。
package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ryuka-games/lokup/domain"
	"github.com/ryuka-games/lokup/features/analyze"
)

// Client は GitHub API クライアント。
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient は Client を生成する。
func NewClient(token string) *Client {
	return &Client{
		baseURL:    "https://api.github.com",
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest は HTTP リクエストを実行する。
func (c *Client) doRequest(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "lokup")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// GetCommits は指定期間のコミット履歴を取得する。
func (c *Client) GetCommits(ctx context.Context, repo domain.Repository, period domain.DateRange) ([]analyze.Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?since=%s&until=%s&per_page=100",
		c.baseURL,
		repo.Owner,
		repo.Name,
		period.From.Format(time.RFC3339),
		period.To.Format(time.RFC3339),
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var apiCommits []apiCommit
	if err := json.NewDecoder(resp.Body).Decode(&apiCommits); err != nil {
		return nil, fmt.Errorf("failed to decode commits: %w", err)
	}

	// TODO: 各コミットの詳細（変更ファイル）を取得する
	// レート制限を考慮して、必要に応じて実装

	commits := make([]analyze.Commit, len(apiCommits))
	for i, ac := range apiCommits {
		commits[i] = analyze.Commit{
			SHA:     ac.SHA,
			Author:  ac.Commit.Author.Name,
			Email:   ac.Commit.Author.Email,
			Date:    ac.Commit.Author.Date,
			Message: ac.Commit.Message,
		}
	}

	return commits, nil
}

// GetContributors はコントリビューター一覧を取得する。
func (c *Client) GetContributors(ctx context.Context, repo domain.Repository) ([]analyze.Contributor, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contributors?per_page=100",
		c.baseURL,
		repo.Owner,
		repo.Name,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contributors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var apiContributors []apiContributor
	if err := json.NewDecoder(resp.Body).Decode(&apiContributors); err != nil {
		return nil, fmt.Errorf("failed to decode contributors: %w", err)
	}

	contributors := make([]analyze.Contributor, len(apiContributors))
	for i, ac := range apiContributors {
		contributors[i] = analyze.Contributor{
			Login:         ac.Login,
			Contributions: ac.Contributions,
		}
	}

	return contributors, nil
}

// GetFileContent はファイルの内容を取得する。
func (c *Client) GetFileContent(ctx context.Context, repo domain.Repository, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s",
		c.baseURL,
		repo.Owner,
		repo.Name,
		path,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var content apiContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(content.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return decoded, nil
}

// GetPullRequests はプルリクエスト一覧を取得する。
func (c *Client) GetPullRequests(ctx context.Context, repo domain.Repository, state string) ([]analyze.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s&per_page=100",
		c.baseURL,
		repo.Owner,
		repo.Name,
		state,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pull requests: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var apiPRs []apiPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&apiPRs); err != nil {
		return nil, fmt.Errorf("failed to decode pull requests: %w", err)
	}

	prs := make([]analyze.PullRequest, len(apiPRs))
	for i, ap := range apiPRs {
		prs[i] = analyze.PullRequest{
			Number:    ap.Number,
			Title:     ap.Title,
			Author:    ap.User.Login,
			CreatedAt: ap.CreatedAt,
			MergedAt:  ap.MergedAt,
		}
	}

	return prs, nil
}

// API レスポンスの型定義

type apiCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

type apiContributor struct {
	Login         string `json:"login"`
	Contributions int    `json:"contributions"`
}

type apiContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type apiPullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	MergedAt  *time.Time `json:"merged_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}
