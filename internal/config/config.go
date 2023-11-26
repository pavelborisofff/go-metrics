package config

import "github.com/pavelborisofff/go-metrics/internal/logger"

var (
	envFile = "config/config.yaml"
	log     = logger.GetLogger()
)

type Config struct {
	ServerAddr string `yaml:"server_addr" env:"ADDRESS"`
	HashKey    string `yaml:"hash_key" env:"KEY"`
	UseHashKey bool
	Server     struct {
		SaveInterval int    `yaml:"save_interval" env:"SAVE_INTERVAL"`
		FileStore    string `yaml:"file_store" env:"FILE_STORAGE_PATH"`
		Restore      bool   `yaml:"restore" env:"RESTORE"`
		DBConn       string `yaml:"db_conn,omitempty" env:"DATABASE_DSN"`
	} `yaml:"server"`
	Agent struct {
		PollInterval   int `yaml:"poll_interval" env:"POLL_INTERVAL"`
		ReportInterval int `yaml:"report_interval" env:"REPORT_INTERVAL,"`
	} `yaml:"agent"`
}
