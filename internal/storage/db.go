package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var DB *pgx.Conn

func InitDB(ps string) error {
	connConfig, err := pgx.ParseConfig(ps)
	if err != nil {
		log.Error("Error parsing connection string", zap.Error(err))
		return err
	}

	db, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		log.Error("Error connecting to DB", zap.Error(err))
		return err
	}

	DB = db
	return nil
}
