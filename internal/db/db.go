package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/retrying"
	"github.com/pavelborisofff/go-metrics/internal/storage"
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

//go:embed queries/create_table_gauges.sql
var createTableGaugesSQL string

//go:embed queries/create_table_counters.sql
var createTableCountersSQL string

func CreateTables() error {
	ctx := context.Background()

	// Create table for GAUGES
	_, err := DB.Exec(ctx, createTableGaugesSQL)
	if err != nil {
		log.Error("Error creating metrics table", zap.Error(err))
		return err
	}

	// Create table for COUNTERS
	_, err = DB.Exec(ctx, createTableCountersSQL)
	if err != nil {
		log.Error("Error creating metrics table", zap.Error(err))
		return err
	}

	return nil
}

//go:embed queries/insert_counter.sql
var insertCounterSQL string

//go:embed queries/insert_gauge.sql
var insertGaugeSQL string

func Write(s *storage.MemStorage) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if DB == nil {
		log.Fatal("DB is nil")
		return nil
	}

	ctx := context.Background()

	// Prepare data for counters
	var i = 1
	counterValues := make([]string, 0, len(s.CounterStorage))

	if len(s.CounterStorage) != 0 {
		counterArgs := make([]interface{}, 0, len(s.CounterStorage)*2)
		for k, v := range s.CounterStorage {
			counterValues = append(counterValues, fmt.Sprintf("($%d, $%d)", i, i+1))
			counterArgs = append(counterArgs, string(k), float64(v))
			i += 2
		}

		// Prepare query for counters
		counterQuery := strings.Replace(insertCounterSQL, "%s", strings.Join(counterValues, ", "), 1)
		log.Debug("Counter query", zap.String("query", counterQuery))

		// Execute query for counters
		err := retrying.DBOperation(func() error {
			_, err := DB.Exec(ctx, counterQuery, counterArgs...)
			if err != nil {
				var pqErr *pgconn.PgError
				if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
					log.Error("UniqueViolation Counter", zap.Error(err))
					return nil
				}
				return err
			}
			return nil
		})

		if err != nil {
			return err
		}
	} else {
		log.Debug("No counters to insert")
	}

	// Prepare data for gauges
	i = 1
	gaugeValues := make([]string, 0, len(s.GaugeStorage))

	if len(s.GaugeStorage) != 0 {
		gaugeArgs := make([]interface{}, 0, len(s.GaugeStorage)*2)

		for k, v := range s.GaugeStorage {
			gaugeValues = append(gaugeValues, fmt.Sprintf("($%d, $%d)", i, i+1))
			gaugeArgs = append(gaugeArgs, string(k), float64(v))
			i += 2
		}

		// Prepare query for gauges
		gaugeQuery := strings.Replace(insertGaugeSQL, "%s", strings.Join(gaugeValues, ", "), 1)
		log.Debug("Gauge query", zap.String("query", gaugeQuery))
		// Execute query for gauges
		err := retrying.DBOperation(func() error {
			_, err := DB.Exec(ctx, gaugeQuery, gaugeArgs...)
			if err != nil {
				var pqErr *pgconn.PgError
				if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
					log.Error("UniqueViolation Gauge", zap.Error(err))
					return nil
				}
				return err
			}
			return nil
		})

		if err != nil {
			return err
		}
	} else {
		log.Debug("No gauges to insert")
	}

	return nil
}

//go:embed queries/read_counter.sql
var readCounterSQL string

//go:embed queries/read_gauge.sql
var readGaugeSQL string

func Read(s *storage.MemStorage) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if DB == nil {
		log.Fatal("DB is nil")
		return nil
	}

	ctx := context.Background()

	err := retrying.DBOperation(func() error {
		counters, err := DB.Query(ctx, readCounterSQL)
		if err != nil {
			log.Error("Error selecting Counter metrics", zap.Error(err))
			return err
		}

		for counters.Next() {
			var k string
			var v float64

			err = counters.Scan(&k, &v)
			if err != nil {
				log.Error("Error reading data", zap.Error(err))
				return err
			}
			s.CounterStorage[k] = storage.Counter(v)
		}

		if err = counters.Err(); err != nil {
			log.Error("Error after reading Counter metrics", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = retrying.DBOperation(func() error {
		gauges, err := DB.Query(ctx, readGaugeSQL)
		if err != nil {
			log.Error("Error selecting Gauge metrics", zap.Error(err))
			return err
		}

		for gauges.Next() {
			var k string
			var v float64

			err = gauges.Scan(&k, &v)
			if err != nil {
				log.Error("Error reading data", zap.Error(err))
				return err
			}
			s.GaugeStorage[k] = storage.Gauge(v)
		}

		if err = gauges.Err(); err != nil {
			log.Error("Error after reading Gauge metrics", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Debug("Metrics loaded from DB", zap.Any("metrics", s))
	return nil
}
