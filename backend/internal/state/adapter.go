package state

import (
	"context"
	"sync"

	"github.com/leshless/pet/cub/internal/model"
)

type HealthAdapter interface {
	SetStatus(ctx context.Context, status model.HealthStatus) error
	GetStatus(ctx context.Context) (model.HealthStatus, error)
}

type healthAdapter struct {
	mu     sync.RWMutex
	status model.HealthStatus
}

func NewHealthAdapter() *healthAdapter {
	return &healthAdapter{}
}

var _ HealthAdapter = (*healthAdapter)(nil)

func (a *healthAdapter) SetStatus(ctx context.Context, status model.HealthStatus) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.status = status
	return nil
}

func (a *healthAdapter) GetStatus(ctx context.Context) (model.HealthStatus, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.status, nil
}
