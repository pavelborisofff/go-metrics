package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var DB *pgx.Conn

func InitDB(ps string) (*pgx.Conn, error) {
	ctx := context.Background()
	connConfig, err := pgx.ParseConfig(ps)
	if err != nil {
		log.Error("Error parsing connection string", zap.Error(err))
		return nil, err
	}

	db, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		log.Error("Error connecting to DB", zap.Error(err))
		return nil, err
	}

	//DB = db
	return db, nil
}

//func (db *DB) Close() {
//	db.Conn.Close(context.Background())
//}

//func SetDB(db *pgx.Conn) {
//	DB = db
//}
