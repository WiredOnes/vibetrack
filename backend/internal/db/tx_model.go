package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/leshless/pet/cub/internal/model"
)

var txIsolationLevelFromModel = map[model.TxIsolationLevel]pgx.TxIsoLevel{
	model.TxIsolationLevelReadUncommitted: pgx.ReadUncommitted,
	model.TxIsolationLevelReadCommitted:   pgx.ReadCommitted,
	model.TxIsolationLevelRepeatableRead:  pgx.RepeatableRead,
	model.TxIsolationLevelSerializable:    pgx.Serializable,
}

func txOptionsFromModel(options model.TxOptions) pgx.TxOptions {
	isolationLevel, ok := txIsolationLevelFromModel[options.IsolationLevel]
	if !ok {
		isolationLevel = pgx.ReadCommitted
	}

	var accessMode pgx.TxAccessMode
	if options.ReadOnly {
		accessMode = pgx.ReadOnly
	} else {
		accessMode = pgx.ReadWrite
	}

	return pgx.TxOptions{
		IsoLevel:   isolationLevel,
		AccessMode: accessMode,
	}
}
