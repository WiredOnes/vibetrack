package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type contextKey struct{}

func ContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, contextKey{}, tx)
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(contextKey{}).(pgx.Tx)
	if tx == nil {
		return nil, false
	}

	return tx, ok
}

func ContextWithoutTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, nil)
}
