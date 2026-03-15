package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/WiredOnes/vibetrack/backend/internal/db"
	"github.com/WiredOnes/vibetrack/backend/internal/environment"
	"github.com/WiredOnes/vibetrack/backend/internal/model"
	"github.com/WiredOnes/vibetrack/backend/internal/state"
	"github.com/WiredOnes/vibetrack/backend/internal/telemetry"
)

type Controller interface {
	CheckHealth(ctx context.Context, arg CheckArg) (CheckRes, error)
	UpdateHealthStatus(ctx context.Context, arg UpdateHealthStatusArg) (UpdateHealthStatusRes, error)
	GetRepositories(ctx context.Context, arg GetRepositoriesArg) (GetRepositoriesRes, error)
	GetRepository(ctx context.Context, arg GetRepositoryArg) (GetRepositoryRes, error)
	AnalyzeRepository(ctx context.Context, arg AnalyzeRepositoryArg) (AnalyzeRepositoryRes, error)
	AnalyzeCommit(ctx context.Context, arg AnalyzeCommitArg) (AnalyzeCommitRes, error)
	ExchangeOAuthCode(ctx context.Context, arg ExchangeOAuthCodeArg) (ExchangeOAuthCodeRes, error)
}

// @PublicValueInstance
type GetRepositoriesArg struct {
	Token string
}

// @PublicValueInstance
type GetRepositoriesRes struct {
	Repositories []Repository
}

// @PublicValueInstance
type GetRepositoryArg struct {
	RepositoryID int
	Token        string
}

// @PublicValueInstance
type GetRepositoryRes struct {
	Repository Repository
}

// @PublicValueInstance
type AnalyzeRepositoryArg struct {
	RepositoryID int
	Token        string
}

// @PublicValueInstance
type AnalyzeRepositoryRes struct {
	Summary      string
	FilesChanged []string
	KeyChanges   []AnalyzeKeyChange
}

// @PublicValueInstance
type AnalyzeKeyChange struct {
	Type        string
	Description string
}

// @PublicValueInstance
type AnalyzeCommitArg struct {
	RepositoryID int
	CommitSHA    string
	Token        string
}

// @PublicValueInstance
type AnalyzeCommitRes struct {
	Summary      string
	FilesChanged []string
	KeyChanges   []AnalyzeKeyChange
}

// @PublicValueInstance
type Repository struct {
	ID            int64
	Name          string
	FullName      string
	Description   string
	Private       bool
	DefaultBranch string
	UpdatedAt     time.Time
}

// @PublicPointerInstance
type controller struct {
	telemetry.Telemetry
	environmentHolder environment.Holder
	healthState       state.HealthAdapter
	pingDB            db.PingAdapter
}

var _ Controller = (*controller)(nil)

// @PublicValueInstance
type CheckArg struct{}

// @PublicValueInstance
type CheckRes struct {
	Status model.HealthStatus
}

func (c *controller) CheckHealth(ctx context.Context, arg CheckArg) (CheckRes, error) {
	c.Logger.Info(ctx, "checking health status")

	status, err := c.healthState.GetStatus(ctx)
	if err != nil {
		c.Logger.Error(ctx, "failed to get health status from state", telemetry.Error(err))
		return CheckRes{}, err
	}

	c.Logger.Info(ctx, "health check completed successfully", telemetry.Any("status", status))
	return NewCheckRes(status), nil
}

// @PublicValueInstance
type UpdateHealthStatusArg struct{}

// @PublicValueInstance
type UpdateHealthStatusRes struct{}

func (c *controller) UpdateHealthStatus(ctx context.Context, arg UpdateHealthStatusArg) (UpdateHealthStatusRes, error) {
	c.Logger.Info(ctx, "updating health status")

	status := model.HealthStatusServing

	c.Logger.Debug(ctx, "pinging database to check health")
	err := c.pingDB.Ping(ctx)
	if err != nil {
		c.Logger.Error(ctx, "database ping failed, setting status to not serving", telemetry.Error(err))
		status = model.HealthStatusNotServing
	} else {
		c.Logger.Debug(ctx, "database ping successful")
	}

	err = c.healthState.SetStatus(ctx, status)
	if err != nil {
		c.Logger.Error(ctx, "failed to update health status in state", telemetry.Error(err))
		return NewUpdateHealthStatusRes(), err
	}

	c.Logger.Info(ctx, "health status updated successfully", telemetry.Any("status", status))
	return NewUpdateHealthStatusRes(), nil
}

func (c *controller) GetRepositories(ctx context.Context, arg GetRepositoriesArg) (GetRepositoriesRes, error) {
	c.Logger.Info(ctx, "fetching user repositories from GitHub")

	if arg.Token == "" {
		c.Logger.Warn(ctx, "missing GitHub token in request")
		return GetRepositoriesRes{}, model.NewUnauthenticatedError()
	}

	client := &http.Client{Timeout: 15 * time.Second}
	var repos []Repository
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/user/repos?per_page=100&page=%d", page)
		c.Logger.Debug(ctx, "fetching repositories page", telemetry.Any("page", page), telemetry.Any("url", url))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			c.Logger.Error(ctx, "failed to create HTTP request for GitHub API", telemetry.Error(err))
			return GetRepositoriesRes{}, model.NewInternalError()
		}
		req.Header.Set("Authorization", "Bearer "+arg.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := client.Do(req)
		if err != nil {
			c.Logger.Error(ctx, "failed to call GitHub API", telemetry.Error(err))
			return GetRepositoriesRes{}, model.NewInternalError()
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			c.Logger.Warn(ctx, "GitHub API returned unauthorized/forbidden", telemetry.Any("status_code", resp.StatusCode))
			return GetRepositoriesRes{}, model.NewUnauthenticatedError()
		}
		if resp.StatusCode >= 400 {
			c.Logger.Error(ctx, "GitHub API returned error", telemetry.Any("status_code", resp.StatusCode))
			return GetRepositoriesRes{}, model.NewInternalError()
		}

		var pageRepos []struct {
			ID            int64     `json:"id"`
			Name          string    `json:"name"`
			FullName      string    `json:"full_name"`
			Description   string    `json:"description"`
			Private       bool      `json:"private"`
			DefaultBranch string    `json:"default_branch"`
			UpdatedAt     time.Time `json:"updated_at"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&pageRepos); err != nil {
			c.Logger.Error(ctx, "failed to decode GitHub API response", telemetry.Error(err))
			return GetRepositoriesRes{}, model.NewInternalError()
		}

		c.Logger.Debug(ctx, "received repositories from page", telemetry.Any("page", page), telemetry.Any("count", len(pageRepos)))

		if len(pageRepos) == 0 {
			break
		}

		for _, r := range pageRepos {
			repos = append(repos, Repository{
				ID:            r.ID,
				Name:          r.Name,
				FullName:      r.FullName,
				Description:   r.Description,
				Private:       r.Private,
				DefaultBranch: r.DefaultBranch,
				UpdatedAt:     r.UpdatedAt,
			})
		}

		if len(pageRepos) < 100 {
			break
		}
		page++
	}

	c.Logger.Info(ctx, "successfully fetched repositories", telemetry.Any("total_count", len(repos)))
	return GetRepositoriesRes{Repositories: repos}, nil
}

func (c *controller) AnalyzeRepository(ctx context.Context, arg AnalyzeRepositoryArg) (AnalyzeRepositoryRes, error) {
	c.Logger.Info(ctx, "analyzing repository with AI")

	if arg.Token == "" {
		c.Logger.Warn(ctx, "missing GitHub token in request")
		return AnalyzeRepositoryRes{}, model.NewUnauthenticatedError()
	}

	repoFullName, defaultBranch, err := c.getRepoFullNameAndBranch(ctx, arg.RepositoryID, arg.Token)
	if err != nil {
		return AnalyzeRepositoryRes{}, err
	}

	tree, err := c.getRepoTree(ctx, repoFullName, defaultBranch, arg.Token)
	if err != nil {
		return AnalyzeRepositoryRes{}, err
	}

	promptContent := analyzeRepositoryPrompt + "\n\n" + tree
	resp, err := c.requestAI(ctx, promptContent)
	if err != nil {
		return AnalyzeRepositoryRes{}, err
	}

	analysis, err := parseRepositoryAnalysisResponse(resp)
	if err != nil {
		c.Logger.Error(ctx, "failed to parse AI response", telemetry.Error(err))
		return AnalyzeRepositoryRes{}, model.NewInternalError()
	}

	return analysis, nil
}

func (c *controller) AnalyzeCommit(ctx context.Context, arg AnalyzeCommitArg) (AnalyzeCommitRes, error) {
	c.Logger.Info(ctx, "analyzing commit with AI", telemetry.Any("commit", arg.CommitSHA))

	if arg.Token == "" {
		c.Logger.Warn(ctx, "missing GitHub token in request")
		return AnalyzeCommitRes{}, model.NewUnauthenticatedError()
	}

	repoFullName, _, err := c.getRepoFullNameAndBranch(ctx, arg.RepositoryID, arg.Token)
	if err != nil {
		return AnalyzeCommitRes{}, err
	}

	diff, err := c.getCommitDiff(ctx, repoFullName, arg.CommitSHA, arg.Token)
	if err != nil {
		return AnalyzeCommitRes{}, err
	}

	promptContent := analyzeCommitPrompt + "\n\n" + diff
	resp, err := c.requestAI(ctx, promptContent)
	if err != nil {
		return AnalyzeCommitRes{}, err
	}

	analysis, err := parseCommitAnalysisResponse(resp)
	if err != nil {
		c.Logger.Error(ctx, "failed to parse AI response", telemetry.Error(err))
		return AnalyzeCommitRes{}, model.NewInternalError()
	}

	return analysis, nil
}

// @PublicValueInstance
type ExchangeOAuthCodeArg struct {
	Code string
}

// @PublicValueInstance
type ExchangeOAuthCodeRes struct {
	Token string
}

func (c *controller) ExchangeOAuthCode(ctx context.Context, arg ExchangeOAuthCodeArg) (ExchangeOAuthCodeRes, error) {
	c.Logger.Info(ctx, "exchanging OAuth code for access token")

	if arg.Code == "" {
		c.Logger.Warn(ctx, "missing required parameters for OAuth exchange", telemetry.Any("has_code", arg.Code != ""))
		return ExchangeOAuthCodeRes{}, model.NewBadRequestError()
	}

	client := &http.Client{Timeout: 15 * time.Second}
	body := map[string]string{
		"client_id":     c.environmentHolder.Environment().GithubClientID,
		"client_secret": c.environmentHolder.Environment().GithubClientSecret,
		"code":          arg.Code,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		c.Logger.Error(ctx, "failed to marshal request body for OAuth exchange", telemetry.Error(err))
		return ExchangeOAuthCodeRes{}, model.NewInternalError()
	}

	c.Logger.Debug(ctx, "sending POST request to GitHub OAuth endpoint")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewReader(reqBody))
	if err != nil {
		c.Logger.Error(ctx, "failed to create HTTP request for OAuth exchange", telemetry.Error(err))
		return ExchangeOAuthCodeRes{}, model.NewInternalError()
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		c.Logger.Error(ctx, "failed to call GitHub OAuth endpoint", telemetry.Error(err))
		return ExchangeOAuthCodeRes{}, model.NewInternalError()
	}
	defer resp.Body.Close()

	c.Logger.Debug(ctx, "received response from GitHub OAuth", telemetry.Any("status_code", resp.StatusCode))

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.Logger.Warn(ctx, "GitHub OAuth returned unauthorized/forbidden", telemetry.Any("status_code", resp.StatusCode))
		return ExchangeOAuthCodeRes{}, model.NewUnauthenticatedError()
	}
	if resp.StatusCode >= 400 {
		c.Logger.Error(ctx, "GitHub OAuth returned error", telemetry.Any("status_code", resp.StatusCode))
		return ExchangeOAuthCodeRes{}, model.NewInternalError()
	}

	var data struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.Logger.Error(ctx, "failed to decode GitHub OAuth response", telemetry.Error(err))
		return ExchangeOAuthCodeRes{}, model.NewInternalError()
	}

	if data.Error != "" {
		c.Logger.Warn(ctx, "GitHub OAuth returned error in response", telemetry.Any("error", data.Error), telemetry.Any("error_description", data.ErrorDesc))
		return ExchangeOAuthCodeRes{}, model.NewUnauthenticatedError()
	}

	c.Logger.Info(ctx, "successfully exchanged OAuth code for access token")
	return ExchangeOAuthCodeRes{Token: data.AccessToken}, nil
}

func (c *controller) getRepoFullNameAndBranch(ctx context.Context, repoID int, token string) (string, string, error) {
	c.Logger.Debug(ctx, "fetching repository metadata", telemetry.Any("repo_id", repoID))

	client := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repositories/%d", repoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Error(ctx, "failed to create HTTP request for GitHub repository metadata", telemetry.Error(err))
		return "", "", model.NewInternalError()
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		c.Logger.Error(ctx, "failed to call GitHub API for repository metadata", telemetry.Error(err))
		return "", "", model.NewInternalError()
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.Logger.Warn(ctx, "GitHub API returned unauthorized/forbidden when fetching repository metadata", telemetry.Any("status_code", resp.StatusCode))
		return "", "", model.NewUnauthenticatedError()
	}
	if resp.StatusCode >= 400 {
		c.Logger.Error(ctx, "GitHub API returned error when fetching repository metadata", telemetry.Any("status_code", resp.StatusCode))
		return "", "", model.NewInternalError()
	}

	var repo struct {
		FullName      string `json:"full_name"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		c.Logger.Error(ctx, "failed to decode GitHub repository metadata response", telemetry.Error(err))
		return "", "", model.NewInternalError()
	}

	return repo.FullName, repo.DefaultBranch, nil
}

func (c *controller) getRepoTree(ctx context.Context, fullName, branch, token string) (string, error) {
	c.Logger.Debug(ctx, "fetching repository tree", telemetry.Any("repo", fullName), telemetry.Any("branch", branch))

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", fullName, branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Error(ctx, "failed to create HTTP request for GitHub tree", telemetry.Error(err))
		return "", model.NewInternalError()
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		c.Logger.Error(ctx, "failed to call GitHub API for tree", telemetry.Error(err))
		return "", model.NewInternalError()
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.Logger.Warn(ctx, "GitHub API returned unauthorized/forbidden when fetching tree", telemetry.Any("status_code", resp.StatusCode))
		return "", model.NewUnauthenticatedError()
	}
	if resp.StatusCode >= 400 {
		c.Logger.Error(ctx, "GitHub API returned error when fetching tree", telemetry.Any("status_code", resp.StatusCode))
		return "", model.NewInternalError()
	}

	var treeResp struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		c.Logger.Error(ctx, "failed to decode GitHub tree response", telemetry.Error(err))
		return "", model.NewInternalError()
	}

	var b strings.Builder
	for _, item := range treeResp.Tree {
		b.WriteString(item.Path)
		b.WriteString("\n")
	}

	return b.String(), nil
}

func (c *controller) getCommitDiff(ctx context.Context, fullName, sha, token string) (string, error) {
	c.Logger.Debug(ctx, "fetching commit diff", telemetry.Any("repo", fullName), telemetry.Any("commit", sha))

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", fullName, sha)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Error(ctx, "failed to create HTTP request for GitHub commit", telemetry.Error(err))
		return "", model.NewInternalError()
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		c.Logger.Error(ctx, "failed to call GitHub API for commit", telemetry.Error(err))
		return "", model.NewInternalError()
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.Logger.Warn(ctx, "GitHub API returned unauthorized/forbidden when fetching commit", telemetry.Any("status_code", resp.StatusCode))
		return "", model.NewUnauthenticatedError()
	}
	if resp.StatusCode >= 400 {
		c.Logger.Error(ctx, "GitHub API returned error when fetching commit", telemetry.Any("status_code", resp.StatusCode))
		return "", model.NewInternalError()
	}

	var commitResp struct {
		Files []struct {
			Filename string `json:"filename"`
			Patch    string `json:"patch"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commitResp); err != nil {
		c.Logger.Error(ctx, "failed to decode GitHub commit response", telemetry.Error(err))
		return "", model.NewInternalError()
	}

	var b strings.Builder
	for _, f := range commitResp.Files {
		b.WriteString("FILE: ")
		b.WriteString(f.Filename)
		b.WriteString("\n")
		b.WriteString(f.Patch)
		b.WriteString("\n\n")
	}

	return b.String(), nil
}

func (c *controller) requestAI(ctx context.Context, promptContent string) (string, error) {
	c.Logger.Debug(ctx, "requesting AI analysis")

	env := c.environmentHolder.Environment()
	client := NewGigaChatClient(env.GigaChatClientID, env.GigaChatClientSecret)
	resp, err := client.SendChatRequest(ctx, []Message{{Role: "user", Content: promptContent}})
	if err != nil {
		c.Logger.Error(ctx, "failed to request AI service", telemetry.Error(err))
		return "", model.NewInternalError()
	}

	if len(resp.Choices) == 0 {
		c.Logger.Error(ctx, "AI service returned empty response")
		return "", model.NewInternalError()
	}

	return resp.Choices[0].Message.Content, nil
}

func parseRepositoryAnalysisResponse(resp string) (AnalyzeRepositoryRes, error) {
	result, err := parseAIResponse(resp)
	if err != nil {
		return AnalyzeRepositoryRes{}, err
	}
	return result, nil
}

func parseCommitAnalysisResponse(resp string) (AnalyzeCommitRes, error) {
	result, err := parseAIResponse(resp)
	if err != nil {
		return AnalyzeCommitRes{}, err
	}
	return AnalyzeCommitRes(result), nil
}

func parseAIResponse(resp string) (AnalyzeRepositoryRes, error) {
	trimmed := strings.TrimSpace(resp)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || start >= end {
		return AnalyzeRepositoryRes{}, fmt.Errorf("failed to locate JSON object in AI response")
	}

	jsonPart := trimmed[start : end+1]

	type rawAnalysis struct {
		Summary      string   `json:"summary"`
		FilesChanged []string `json:"files_changed"`
		KeyChanges   []struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"key_changes"`
	}

	var raw rawAnalysis
	if err := json.Unmarshal([]byte(jsonPart), &raw); err != nil {
		return AnalyzeRepositoryRes{}, fmt.Errorf("failed to parse AI response JSON: %w", err)
	}

	result := AnalyzeRepositoryRes{
		Summary:      raw.Summary,
		FilesChanged: raw.FilesChanged,
	}

	for _, kc := range raw.KeyChanges {
		result.KeyChanges = append(result.KeyChanges, AnalyzeKeyChange{Type: kc.Type, Description: kc.Description})
	}

	return result, nil
}
