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
	"log"
	"net/http"
	"strings"
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
			Number:     ap.Number,
			Title:      ap.Title,
			Author:     ap.User.Login,
			HeadBranch: ap.Head.Ref,
			CreatedAt:  ap.CreatedAt,
			MergedAt:   ap.MergedAt,
		}
		// Note: additions/deletions は一覧APIに含まれないため、
		// 必要なPRのみ getPRDetail で個別取得する（buildPRDetails参照）
	}

	return prs, nil
}

// GetPRDetail はPRの詳細（additions/deletions含む）を取得する。
func (c *Client) GetPRDetail(ctx context.Context, repo domain.Repository, prNumber int) (*analyze.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d",
		c.baseURL,
		repo.Owner,
		repo.Name,
		prNumber,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var ap apiPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&ap); err != nil {
		return nil, err
	}

	return &analyze.PullRequest{
		Number:     ap.Number,
		Title:      ap.Title,
		Author:     ap.User.Login,
		HeadBranch: ap.Head.Ref,
		CreatedAt:  ap.CreatedAt,
		MergedAt:   ap.MergedAt,
		Additions:  ap.Additions,
		Deletions:  ap.Deletions,
	}, nil
}

// GetFiles はリポジトリ内のファイル一覧を取得する。
func (c *Client) GetFiles(ctx context.Context, repo domain.Repository) ([]analyze.File, error) {
	// デフォルトブランチのツリーを取得（recursive=1で全階層）
	url := fmt.Sprintf("%s/repos/%s/%s/git/trees/HEAD?recursive=1",
		c.baseURL,
		repo.Owner,
		repo.Name,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tree: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var tree apiTree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("failed to decode tree: %w", err)
	}

	// blob（ファイル）のみを抽出
	var files []analyze.File
	for _, item := range tree.Tree {
		if item.Type == "blob" {
			files = append(files, analyze.File{
				Path: item.Path,
				Size: item.Size,
			})
		}
	}

	return files, nil
}

// GetDependencies は各種依存ファイルから依存情報を取得する。
func (c *Client) GetDependencies(ctx context.Context, repo domain.Repository) ([]analyze.Dependency, error) {
	var allDependencies []analyze.Dependency

	// npm (package.json)
	npmDeps, err := c.getNpmDependencies(ctx, repo)
	if err != nil {
		log.Printf("[debug] npm dependencies not found: %v", err)
	}
	allDependencies = append(allDependencies, npmDeps...)

	// Go (go.mod)
	goDeps, err := c.getGoDependencies(ctx, repo)
	if err != nil {
		log.Printf("[debug] go dependencies not found: %v", err)
	}
	allDependencies = append(allDependencies, goDeps...)

	// Python (requirements.txt)
	pyDeps, err := c.getPythonDependencies(ctx, repo)
	if err != nil {
		log.Printf("[debug] python dependencies not found: %v", err)
	}
	allDependencies = append(allDependencies, pyDeps...)

	// .NET (*.csproj)
	dotnetDeps, err := c.getDotNetDependencies(ctx, repo)
	if err != nil {
		log.Printf("[debug] dotnet dependencies not found: %v", err)
	}
	allDependencies = append(allDependencies, dotnetDeps...)

	return allDependencies, nil
}

// GetIssues はIssue一覧を取得する。
func (c *Client) GetIssues(ctx context.Context, repo domain.Repository, state string, since *time.Time) ([]analyze.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=%s&per_page=100",
		c.baseURL,
		repo.Owner,
		repo.Name,
		state,
	)

	if since != nil {
		url += "&since=" + since.Format(time.RFC3339)
	}

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var apiIssues []apiIssue
	if err := json.NewDecoder(resp.Body).Decode(&apiIssues); err != nil {
		return nil, fmt.Errorf("failed to decode issues: %w", err)
	}

	// PRを除外（GitHub APIではPRもIssueとして返される）
	var issues []analyze.Issue
	for _, ai := range apiIssues {
		if ai.PullRequest != nil {
			continue // PRは除外
		}
		labels := make([]string, len(ai.Labels))
		for j, l := range ai.Labels {
			labels[j] = l.Name
		}
		issues = append(issues, analyze.Issue{
			Number:    ai.Number,
			Title:     ai.Title,
			State:     ai.State,
			Labels:    labels,
			CreatedAt: ai.CreatedAt,
			ClosedAt:  ai.ClosedAt,
		})
	}

	return issues, nil
}

// GetPRReviews はPRのレビュー一覧を取得する。
func (c *Client) GetPRReviews(ctx context.Context, repo domain.Repository, prNumber int) ([]analyze.Review, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews?per_page=100",
		c.baseURL,
		repo.Owner,
		repo.Name,
		prNumber,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var apiReviews []apiReview
	if err := json.NewDecoder(resp.Body).Decode(&apiReviews); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}

	reviews := make([]analyze.Review, len(apiReviews))
	for i, ar := range apiReviews {
		reviews[i] = analyze.Review{
			ID:          ar.ID,
			Author:      ar.User.Login,
			State:       ar.State,
			SubmittedAt: ar.SubmittedAt,
		}
	}

	return reviews, nil
}

// GetReleases はリリース一覧を取得する。
func (c *Client) GetReleases(ctx context.Context, repo domain.Repository) ([]analyze.Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100",
		c.baseURL,
		repo.Owner,
		repo.Name,
	)

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var apiReleases []apiRelease
	if err := json.NewDecoder(resp.Body).Decode(&apiReleases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	releases := make([]analyze.Release, len(apiReleases))
	for i, ar := range apiReleases {
		releases[i] = analyze.Release{
			ID:          ar.ID,
			TagName:     ar.TagName,
			Name:        ar.Name,
			PublishedAt: ar.PublishedAt,
		}
	}

	return releases, nil
}

// getNpmDependencies はpackage.jsonから依存を取得する。
func (c *Client) getNpmDependencies(ctx context.Context, repo domain.Repository) ([]analyze.Dependency, error) {
	content, err := c.GetFileContent(ctx, repo, "package.json")
	if err != nil {
		return nil, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, err
	}

	allDeps := make(map[string]string)
	for name, version := range pkg.Dependencies {
		allDeps[name] = version
	}
	for name, version := range pkg.DevDependencies {
		allDeps[name] = version
	}

	var dependencies []analyze.Dependency

	for name, version := range allDeps {
		cleanVersion := strings.TrimLeft(version, "^~>=<")
		releasedAt, err := c.getNpmReleaseDate(ctx, name, cleanVersion)
		if err != nil {
			continue
		}
		dependencies = append(dependencies, analyze.Dependency{
			Name:        name,
			Version:     cleanVersion,
			ReleasedAt:  releasedAt,
			AgeMonths:   ageMonths(releasedAt),
			PackageType: "npm",
		})
	}

	return dependencies, nil
}

// getGoDependencies はgo.modから依存を取得する。
func (c *Client) getGoDependencies(ctx context.Context, repo domain.Repository) ([]analyze.Dependency, error) {
	content, err := c.GetFileContent(ctx, repo, "go.mod")
	if err != nil {
		return nil, err
	}

	var dependencies []analyze.Dependency

	lines := strings.Split(string(content), "\n")
	inRequire := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if line == ")" {
			inRequire = false
			continue
		}

		// require行をパース
		var moduleLine string
		if inRequire {
			moduleLine = line
		} else if strings.HasPrefix(line, "require ") {
			moduleLine = strings.TrimPrefix(line, "require ")
		} else {
			continue
		}

		parts := strings.Fields(moduleLine)
		if len(parts) < 2 {
			continue
		}

		modulePath := parts[0]
		version := strings.TrimPrefix(parts[1], "v")

		releasedAt, err := c.getGoReleaseDate(ctx, modulePath, parts[1])
		if err != nil {
			continue
		}

		dependencies = append(dependencies, analyze.Dependency{
			Name:        modulePath,
			Version:     version,
			ReleasedAt:  releasedAt,
			AgeMonths:   ageMonths(releasedAt),
			PackageType: "go",
		})
	}

	return dependencies, nil
}

// getPythonDependencies はrequirements.txtから依存を取得する。
func (c *Client) getPythonDependencies(ctx context.Context, repo domain.Repository) ([]analyze.Dependency, error) {
	content, err := c.GetFileContent(ctx, repo, "requirements.txt")
	if err != nil {
		return nil, err
	}

	var dependencies []analyze.Dependency

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// package==version 形式をパース
		var name, version string
		if strings.Contains(line, "==") {
			parts := strings.Split(line, "==")
			name = parts[0]
			version = parts[1]
		} else if strings.Contains(line, ">=") {
			parts := strings.Split(line, ">=")
			name = parts[0]
			version = parts[1]
		} else {
			continue
		}

		releasedAt, err := c.getPyPIReleaseDate(ctx, name, version)
		if err != nil {
			continue
		}

		dependencies = append(dependencies, analyze.Dependency{
			Name:        name,
			Version:     version,
			ReleasedAt:  releasedAt,
			AgeMonths:   ageMonths(releasedAt),
			PackageType: "python",
		})
	}

	return dependencies, nil
}

// getDotNetDependencies は.csprojから依存を取得する。
func (c *Client) getDotNetDependencies(ctx context.Context, repo domain.Repository) ([]analyze.Dependency, error) {
	// ファイル一覧から.csprojを探す
	files, err := c.GetFiles(ctx, repo)
	if err != nil {
		return nil, err
	}

	var dependencies []analyze.Dependency

	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".csproj") {
			continue
		}

		content, err := c.GetFileContent(ctx, repo, f.Path)
		if err != nil {
			continue
		}

		// 簡易的なXMLパース（PackageReferenceを抽出）
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.Contains(line, "PackageReference") {
				continue
			}

			// Include="..." と Version="..." を抽出
			name := extractAttribute(line, "Include")
			version := extractAttribute(line, "Version")
			if name == "" || version == "" {
				continue
			}

			releasedAt, err := c.getNuGetReleaseDate(ctx, name, version)
			if err != nil {
				continue
			}

			dependencies = append(dependencies, analyze.Dependency{
				Name:        name,
				Version:     version,
				ReleasedAt:  releasedAt,
				AgeMonths:   ageMonths(releasedAt),
				PackageType: "nuget",
			})
		}
	}

	return dependencies, nil
}

// extractAttribute はXML属性値を抽出する。
func extractAttribute(line, attr string) string {
	pattern := attr + `="`
	start := strings.Index(line, pattern)
	if start == -1 {
		return ""
	}
	start += len(pattern)
	end := strings.Index(line[start:], `"`)
	if end == -1 {
		return ""
	}
	return line[start : start+end]
}

// fetchJSON は外部APIにGETリクエストを送り、レスポンスをJSONデコードする。
func (c *Client) fetchJSON(ctx context.Context, url string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s: %s", resp.Status, url)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// ageMonths はリリース日から現在までの月数を計算する。
func ageMonths(releasedAt time.Time) int {
	return int(time.Since(releasedAt).Hours() / 24 / 30)
}

// getNpmReleaseDate はnpmレジストリから特定バージョンのリリース日を取得する。
func (c *Client) getNpmReleaseDate(ctx context.Context, packageName, version string) (time.Time, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s", packageName)

	var npmResp npmRegistryResponse
	if err := c.fetchJSON(ctx, url, &npmResp); err != nil {
		return time.Time{}, err
	}

	// 指定バージョンのリリース日を探す
	if releaseTime, ok := npmResp.Time[version]; ok {
		return releaseTime, nil
	}

	// 完全一致がない場合、部分一致を試す（1.0.0 が 1.0.0-beta などにマッチ）
	for v, t := range npmResp.Time {
		if strings.HasPrefix(v, version) {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("version %s not found", version)
}

// getGoReleaseDate はGo Proxyから特定バージョンのリリース日を取得する。
func (c *Client) getGoReleaseDate(ctx context.Context, modulePath, version string) (time.Time, error) {
	// モジュールパスをエスケープ（大文字を!小文字に変換）
	var escaped strings.Builder
	for _, r := range modulePath {
		if r >= 'A' && r <= 'Z' {
			escaped.WriteByte('!')
			escaped.WriteRune(r + ('a' - 'A'))
		} else {
			escaped.WriteRune(r)
		}
	}

	url := fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.info", escaped.String(), version)

	var goResp goProxyResponse
	if err := c.fetchJSON(ctx, url, &goResp); err != nil {
		return time.Time{}, err
	}

	return goResp.Time, nil
}

// getPyPIReleaseDate はPyPIから特定バージョンのリリース日を取得する。
func (c *Client) getPyPIReleaseDate(ctx context.Context, packageName, version string) (time.Time, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)

	var pypiResp pypiResponse
	if err := c.fetchJSON(ctx, url, &pypiResp); err != nil {
		return time.Time{}, err
	}

	if releases, ok := pypiResp.Releases[version]; ok && len(releases) > 0 {
		return releases[0].UploadTime, nil
	}

	return time.Time{}, fmt.Errorf("version %s not found", version)
}

// getNuGetReleaseDate はNuGetから特定バージョンのリリース日を取得する。
func (c *Client) getNuGetReleaseDate(ctx context.Context, packageName, version string) (time.Time, error) {
	url := fmt.Sprintf("https://api.nuget.org/v3/registration5-gz-semver2/%s/%s.json",
		strings.ToLower(packageName), strings.ToLower(version))

	var nugetResp nugetResponse
	if err := c.fetchJSON(ctx, url, &nugetResp); err != nil {
		return time.Time{}, err
	}

	return nugetResp.Published, nil
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
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	MergedAt  *time.Time `json:"merged_at"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"` // ブランチ名
	} `json:"head"`
}

type apiTree struct {
	Tree []apiTreeItem `json:"tree"`
}

type apiTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"` // "blob" or "tree"
	Size int    `json:"size"` // ファイルサイズ（blobのみ）
}

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type npmRegistryResponse struct {
	Time map[string]time.Time `json:"time"`
}

type goProxyResponse struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

type pypiResponse struct {
	Releases map[string][]pypiRelease `json:"releases"`
}

type pypiRelease struct {
	UploadTime time.Time `json:"upload_time_iso_8601"`
}

type nugetResponse struct {
	Published time.Time `json:"published"`
}

type apiIssue struct {
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	State       string     `json:"state"`
	CreatedAt   time.Time  `json:"created_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	PullRequest *struct{}  `json:"pull_request"` // PRかどうかの判定用（nilならIssue）
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

type apiRelease struct {
	ID          int       `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
}

type apiReview struct {
	ID          int       `json:"id"`
	State       string    `json:"state"`
	SubmittedAt time.Time `json:"submitted_at"`
	User        struct {
		Login string `json:"login"`
	} `json:"user"`
}
