package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"

	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/storage"
)

const (
	pollIntervalDef   = 2
	reportIntervalDef = 10
	serverAddrDef     = "localhost:8080"
)

var (
	pollInterval   time.Duration
	reportInterval time.Duration
	serverAddr     string
	log            = logger.Log
)

func ParseFlags() {
	var (
		err                error
		serverAddrFlag     string
		pollIntervalFlag   int
		reportIntervalFlag int
	)

	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.IntVar(&pollIntervalFlag, "p", pollIntervalDef, "Poll interval")
	flag.IntVar(&reportIntervalFlag, "r", reportIntervalDef, "Report interval")
	flag.Parse()

	//serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	if exists {
		serverAddrFlag = serverAddrEnv
	}
	serverAddr = fmt.Sprintf("http://%s", serverAddrFlag)

	pollIntervalEnv, exists := os.LookupEnv("POLL_INTERVAL")
	if exists {
		pollIntervalFlag, err = strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatal("Error parsing POLL_INTERVAL", zap.Error(err))
		}
	}
	pollInterval = time.Duration(pollIntervalFlag) * time.Second

	if pollInterval < time.Duration(1)*time.Second {
		log.Fatal("Poll interval must be >= 1s")
	}

	reportIntervalEnv, exists := os.LookupEnv("REPORT_INTERVAL")
	if exists {
		reportIntervalFlag, err = strconv.Atoi(reportIntervalEnv)
		if err != nil {
			log.Fatal("Error parsing REPORT_INTERVAL", zap.Error(err))
		}
	}
	reportInterval = time.Duration(reportIntervalFlag) * time.Second

	if reportInterval < time.Duration(1)*time.Second {
		log.Fatal("Report interval must be >= 1s")
	}

	msg := fmt.Sprintf("\nServer address: %s\nPoll interval: %v\nReport interval: %v", serverAddr, pollInterval, reportInterval)
	log.Info(msg)
}

func main() {
	err := logger.InitLogger()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer log.Sync()

	s := storage.NewAgentStorage()
	ParseFlags()

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			s.UpdateMetrics()
		case <-reportTicker.C:
			err := s.SendJSONMetrics(serverAddr)
			if err != nil {
				return
			}
		}
	}
}
