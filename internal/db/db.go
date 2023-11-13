package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/pavelborisofff/go-metrics/internal/storage"
	"go.uber.org/zap"

	"github.com/pavelborisofff/go-metrics/internal/logger"
)

var (
	DB  *pgx.Conn
	log = logger.GetLogger()
)

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

func CreateTables() error {
	ctx := context.Background()

	// Create table for GAUGES
	_, err := DB.Exec(ctx,
		`create table if not exists gauges (id serial primary key, name varchar(255) unique, value double precision)`)
	if err != nil {
		log.Error("Error creating metrics table", zap.Error(err))
	}

	// Create table for COUNTERS
	_, err = DB.Exec(ctx,
		`create table if not exists counters (id serial primary key, name varchar(255) unique, value double precision)`)
	if err != nil {
		log.Error("Error creating metrics table", zap.Error(err))
	}

	return nil
}

func ToDatabase(s *storage.MemStorage) error {
	if DB == nil {
		log.Fatal("DB is nil")
		return nil
	}

	ctx := context.Background()

	for k, v := range s.CounterStorage {
		_, err := DB.Exec(ctx, "insert into counters (name, value) values ($1, $2) on conflict (name) do update set value = $2", k, v)
		if err != nil {
			log.Error("Error inserting Counter metrics", zap.Error(err))
			return err
		}
	}

	for k, v := range s.GaugeStorage {
		_, err := DB.Exec(ctx, "insert into gauges (name, value) values ($1, $2) on conflict (name) do update set value = $2", k, v)
		if err != nil {
			log.Error("Error inserting Gauge metrics", zap.Error(err))
			return err
		}
	}

	return nil
}

func FromDatabase(s *storage.MemStorage) error {
	if DB == nil {
		log.Fatal("DB is nil")
		return nil
	}

	ctx := context.Background()
	counters, err := DB.Query(ctx, "select name, value from counters")

	if err != nil {
		log.Error("Error selecting Counter metrics", zap.Error(err))
		return err
	}

	for counters.Next() {
		row, err := counters.Values()
		if err != nil {
			log.Error("Error reading data", zap.Error(err))
			return err
		}
		s.CounterStorage[row[0].(string)] = storage.Counter(row[1].(float64))
	}

	gauges, err := DB.Query(ctx, "select name, value from gauges")

	if err != nil {
		log.Error("Error selecting Gauge metrics", zap.Error(err))
		return err
	}

	for gauges.Next() {
		row, err := gauges.Values()
		if err != nil {
			log.Error("Error reading data", zap.Error(err))
			return err
		}
		s.GaugeStorage[row[0].(string)] = storage.Gauge(row[1].(float64))
	}

	log.Debug("Metrics loaded from DB", zap.Any("metrics", s))
	return nil
}
