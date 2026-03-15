package db

import (
	"context"
	"errors"

	"github.com/leshless/pet/cub/internal/model"
	"github.com/leshless/pet/cub/internal/telemetry"
)

type TxAdapter interface {
	Begin(ctx context.Context, options model.TxOptions) (context.Context, error)
	Commit(ctx context.Context) (context.Context, error)
	Rollback(ctx context.Context) (context.Context, error)
}

// @PublicPointerInstance
type txAdapter struct {
	telemetry.Telemetry
	client Client
}

var _ TxAdapter = (*txAdapter)(nil)

func (a *txAdapter) Begin(ctx context.Context, options model.TxOptions) (context.Context, error) {
	tx, err := a.client.BeginTx(ctx, txOptionsFromModel(options))
	if err != nil {
		a.Logger.Error(ctx, "failed to begin tx", telemetry.Error(err))
		return nil, err
	}

	return ContextWithTx(ctx, tx), nil
}

func (a *txAdapter) Commit(ctx context.Context) (context.Context, error) {
	tx, ok := TxFromContext(ctx)
	if !ok {
		a.Logger.Error(ctx, "no tx found in context")
		return nil, errors.New("no tx provided in context")
	}

	if err := tx.Commit(ctx); err != nil {
		a.Logger.Error(ctx, "failed to commit tx", telemetry.Error(err))
		return nil, err
	}

	return ContextWithoutTx(ctx), nil
}

func (a *txAdapter) Rollback(ctx context.Context) (context.Context, error) {
	tx, ok := TxFromContext(ctx)
	if !ok {
		a.Logger.Error(ctx, "no tx found in context")
		return nil, errors.New("no tx provided in context")
	}

	if err := tx.Rollback(ctx); err != nil {
		a.Logger.Error(ctx, "failed to rollback tx", telemetry.Error(err))
		return nil, err
	}

	return ContextWithoutTx(ctx), nil
}
