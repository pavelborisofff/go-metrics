package main

import (
	"context"
	"flag"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pavelborisofff/go-metrics/internal/db"
	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/routers"
	"github.com/pavelborisofff/go-metrics/internal/storage"
)

const (
	serverAddrDef   = "localhost:8080"
	saveIntervalDef = 2
	fileStoreDef    = "/tmp/metrics-db.json"
	restoreDef      = true
	dbConnDef       = ""
)

var (
	ServerAddr   string
	SaveInterval time.Duration
	FileStore    string
	Restore      bool
	DBConn       string
	log          = logger.GetLogger()
)

func ParseFlags() {
	var (
		err              error
		serverAddrFlag   string
		saveIntervalFlag int
		fileStoreFlag    string
		restoreFlag      bool
		dbConnFlag       string
	)
	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.IntVar(&saveIntervalFlag, "i", saveIntervalDef, "Save to file interval (sec)")
	flag.StringVar(&fileStoreFlag, "f", fileStoreDef, "Server address")
	flag.BoolVar(&restoreFlag, "r", restoreDef, "Restore metrics from storage")
	flag.StringVar(&dbConnFlag, "d", dbConnDef, "Database connection string")
	flag.Parse()

	// Server address
	serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	if exists {
		serverAddrFlag = serverAddrEnv
	}

	ServerAddr = serverAddrFlag

	// Store interval
	saveIntervalEnv, exists := os.LookupEnv("STORE_INTERVAL")
	if exists {
		saveIntervalFlag, err = strconv.Atoi(saveIntervalEnv)
		if err != nil {
			log.Fatal("Error parsing STORE_INTERVAL", zap.Error(err))
		}
	}
	SaveInterval = time.Duration(saveIntervalFlag) * time.Second

	// Store file path
	fileStoreEnv, exists := os.LookupEnv("FILE_STORAGE_PATH")
	if exists {
		fileStoreFlag = fileStoreEnv
	}
	FileStore = fileStoreFlag

	// Restore metrics from storage
	restoreEnv, exists := os.LookupEnv("RESTORE")
	if exists {
		restoreFlag, err = strconv.ParseBool(restoreEnv)
		if err != nil {
			log.Fatal("Error parsing RESTORE", zap.Error(err))
		}
	}
	Restore = restoreFlag

	// Database connection
	dbConnEnv, exists := os.LookupEnv("DATABASE_DSN")
	if exists {
		dbConnFlag = dbConnEnv
	}
	DBConn = dbConnFlag

	log.Info("Server address",
		zap.String("Server address", ServerAddr),
		zap.Duration("Save interval", SaveInterval),
		zap.String("File store", FileStore),
		zap.Bool("Restore", Restore),
	)
}

func main() {
	defer log.Sync()

	s := storage.NewMemStorage()
	ParseFlags()

	switch DBConn {
	case "":
		if Restore {
			if err := s.FromFile(FileStore); err != nil {
				log.Fatal("Error restore metrics", zap.Error(err))
			}
			log.Info("Metrics restored from file")
		}
	default:
		if err := db.InitDB(DBConn); err != nil {
			log.Fatal("Error init DB", zap.Error(err))
			DBConn = ""
		}
		defer db.DB.Close(context.Background())

		if err := db.CreateTables(); err != nil {
			log.Error("Error create tables", zap.Error(err))
		}

		log.Debug("DB init")

		if Restore {
			if err := db.FromDatabase(s); err != nil {
				log.Fatal("Error restore metrics", zap.Error(err))
			}
			log.Info("Metrics restored from DB")
		}
	}

	go func() {
		if SaveInterval <= 0 {
			return
		}

		ticker := time.NewTicker(SaveInterval)
		defer ticker.Stop()

		for range ticker.C {
			switch DBConn {
			case "":
				if FileStore == "" {
					return
				}
				if err := s.ToFile(FileStore); err != nil {
					log.Fatal("Error saving metrics", zap.Error(err))
				}
				log.Debug("Metrics saved to file")
			default:
				if err := db.ToDatabase(s); err != nil {
					log.Error("Error saving metrics to Database", zap.Error(err))
				}
				log.Debug("Metrics saved to DB")
			}
		}
	}()

	r := routers.InitRouter()
	log.Fatal("Server error", zap.Error(http.ListenAndServe(ServerAddr, r)))
}
