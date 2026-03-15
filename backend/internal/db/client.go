package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leshless/pet/cub/internal/telemetry"
)

// Client is a general handy wrapper around pgx DB conection
type Client interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error

	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type client struct {
	telemetry.Telemetry
	db *pgx.Conn
}

func newClient(
	tel telemetry.Telemetry,
	db *pgx.Conn,
) *client {
	return &client{
		Telemetry: tel,
		db:        db,
	}
}

var _ Client = (*client)(nil)

func (c *client) Ping(ctx context.Context) error {
	err := c.db.Ping(ctx)

	c.Registry.Counter(
		ctx,
		telemetry.DBRequestsTotal,
		telemetry.Successful(err == nil),
	).Inc()

	return err
}

func (c *client) Close(ctx context.Context) error {
	err := c.db.Close(ctx)

	c.Registry.Counter(
		ctx,
		telemetry.DBRequestsTotal,
		telemetry.Successful(err == nil),
	).Inc()

	return err
}

func (c *client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	var (
		tag pgconn.CommandTag
		err error
	)
	if tx, ok := TxFromContext(ctx); ok {
		tag, err = tx.Exec(ctx, query, args)
	} else {
		tag, err = c.db.Exec(ctx, query, args...)
	}

	c.Registry.Counter(
		ctx,
		telemetry.DBRequestsTotal,
		telemetry.Successful(err == nil),
	).Inc()

	return tag, err
}

func (c *client) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if tx, ok := TxFromContext(ctx); ok {
		rows, err = tx.Query(ctx, query, args)
	} else {
		rows, err = c.db.Query(ctx, query, args...)
	}

	c.Registry.Counter(
		ctx,
		telemetry.DBRequestsTotal,
		telemetry.Successful(err == nil),
	).Inc()

	return rows, err
}

func (c *client) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	// JiC, this probably wont be used at all
	c.Registry.Counter(ctx, telemetry.DBRequestsTotal, telemetry.Successful(true))

	if tx, ok := TxFromContext(ctx); ok {
		return tx.QueryRow(ctx, query, args)
	}

	return c.db.QueryRow(ctx, query, args...)
}

func (c *client) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	if _, ok := TxFromContext(ctx); ok {
		return nil, fmt.Errorf("tx is already in progress")
	}

	tx, err := c.db.BeginTx(ctx, txOptions)

	c.Registry.Counter(
		ctx,
		telemetry.DBRequestsTotal,
		telemetry.Successful(err == nil),
	).Inc()

	return tx, err
}
