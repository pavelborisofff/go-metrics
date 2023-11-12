package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func InitDB(ps string) error {
	connConfig, err := pgx.ParseConfig(ps)
	if err != nil {
		return err
	}

	conn, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return err
	}

	DB = conn
	return nil
}

func PingDB() error {
	err := DB.Ping(context.Background())

	if err != nil {
		return err
	}

	return nil
}
