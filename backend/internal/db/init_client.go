package db

import (
	"context"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/leshless/golibrary/graceful"
	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/environment"
	"github.com/leshless/pet/cub/internal/telemetry"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func InitClient(
	logger telemetry.Logger,
	tel telemetry.Telemetry,
	gracefulRegistrator graceful.Registrator,
	configHolder config.Holder,
	environmentHolder environment.Holder,
) (*client, error) {
	ctx := context.Background()

	logger.Info(ctx, "initializing db client")

	config := configHolder.Config().DB
	dbPassword := environmentHolder.Environment().DBPassword

	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		dbPassword,
		config.Database,
		config.SSLMode,
	)

	dbConn, err := pgx.Connect(ctx, connString)
	if err != nil {
		logger.Error(ctx, "failed to connect to db", telemetry.Error(err))
		return nil, fmt.Errorf("connecting to db: %w", err)
	}

	client := newClient(tel, dbConn)

	gracefulRegistrator.Register(client.Close)

	err = client.Ping(ctx)
	if err != nil {
		logger.Error(ctx, "failed to ping db", telemetry.Error(err))
		return nil, fmt.Errorf("pinging db: %w", err)
	}

	logger.Info(ctx, "checking for applied migrations")

	migrationSourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		logger.Error(ctx, "failed to create migration fs", telemetry.Error(err))
		return nil, fmt.Errorf("creating migration fs: %w", err)
	}

	migrationURL := fmt.Sprintf(
		"pgx5://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
	)

	migrator, err := migrate.NewWithSourceInstance("iofs", migrationSourceDriver, migrationURL)
	if err != nil {
		logger.Error(ctx, "failed to create migrate object", telemetry.Error(err))
		return nil, fmt.Errorf("creating migrate object: %w", err)
	}

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Error(ctx, "failed to apply migrations", telemetry.Error(err))
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info(ctx, "database schema is up to date")
	} else {
		logger.Info(ctx, "migrations applied successfully")
	}

	migrator.Close()

	logger.Info(ctx, "db client successfully initialized")

	return client, nil
}
