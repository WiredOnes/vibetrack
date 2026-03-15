package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/WiredOnes/vibetrack/backend/internal/db"
	"github.com/WiredOnes/vibetrack/backend/internal/model"
	"github.com/WiredOnes/vibetrack/backend/internal/state"
	"github.com/WiredOnes/vibetrack/backend/internal/telemetry"
)

type Controller interface {
	CheckHealth(ctx context.Context, arg CheckArg) (CheckRes, error)
	UpdateHealthStatus(ctx context.Context, arg UpdateHealthStatusArg) (UpdateHealthStatusRes, error)
	GetRepositories(ctx context.Context, arg GetRepositoriesArg) (GetRepositoriesRes, error)
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
	healthState state.HealthAdapter
	pingDB      db.PingAdapter
}

var _ Controller = (*controller)(nil)

// @PublicValueInstance
type CheckArg struct{}

// @PublicValueInstance
type CheckRes struct {
	Status model.HealthStatus
}

func (c *controller) CheckHealth(ctx context.Context, arg CheckArg) (CheckRes, error) {
	status, err := c.healthState.GetStatus(ctx)
	if err != nil {
		c.Logger.Error(ctx, "failed to get health status state", telemetry.Error(err))
		return CheckRes{}, err
	}

	return NewCheckRes(status), nil
}

// @PublicValueInstance
type UpdateHealthStatusArg struct{}

// @PublicValueInstance
type UpdateHealthStatusRes struct{}

func (c *controller) UpdateHealthStatus(ctx context.Context, arg UpdateHealthStatusArg) (UpdateHealthStatusRes, error) {
	status := model.HealthStatusServing

	err := c.pingDB.Ping(ctx)
	if err != nil {
		c.Logger.Error(ctx, "failed to ping database", telemetry.Error(err))
		status = model.HealthStatusNotServing
	}

	err = c.healthState.SetStatus(ctx, status)
	if err != nil {
		c.Logger.Error(ctx, "failed to update health status state", telemetry.Error(err))
		return NewUpdateHealthStatusRes(), err
	}

	return NewUpdateHealthStatusRes(), nil
}

func (c *controller) GetRepositories(ctx context.Context, arg GetRepositoriesArg) (GetRepositoriesRes, error) {
	if arg.Token == "" {
		return GetRepositoriesRes{}, model.NewUnauthenticatedError()
	}

	client := &http.Client{Timeout: 15 * time.Second}
	var repos []Repository
	for page := 1; ; page++ {
		url := fmt.Sprintf("https://api.github.com/user/repos?per_page=100&page=%d", page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return GetRepositoriesRes{}, model.NewInternalError()
		}
		req.Header.Set("Authorization", "Bearer "+arg.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := client.Do(req)
		if err != nil {
			return GetRepositoriesRes{}, model.NewInternalError()
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return GetRepositoriesRes{}, model.NewUnauthenticatedError()
		}
		if resp.StatusCode >= 400 {
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
			return GetRepositoriesRes{}, model.NewInternalError()
		}

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

		// GitHub paginates by 100 items; stop if less than the page size.
		if len(pageRepos) < 100 {
			break
		}
	}

	return GetRepositoriesRes{Repositories: repos}, nil
}
