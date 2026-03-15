package db

import (
	"context"
)

type PingAdapter interface {
	Ping(ctx context.Context) error
}

// @PublicPointerInstance
type pingAdapter struct {
	client Client
}

var _ PingAdapter = (*pingAdapter)(nil)

func (a *pingAdapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx)
}
