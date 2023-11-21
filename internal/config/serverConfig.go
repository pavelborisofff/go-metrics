package config

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

var (
	cfgServer    *Config
	cfgServerErr error
	onceServer   sync.Once
)

func GetServerConfig() (*Config, error) {
	onceServer.Do(func() {
		cfgServer, cfgServerErr = loadServerConfig()
	})
	return cfgServer, cfgServerErr
}

func loadServerConfig() (*Config, error) {
	var _cfg Config

	log.Info("Loading default config from file", zap.String("file", envFile))
	err := cleanenv.ReadConfig(envFile, &_cfg)
	if err != nil {
		log.Warn("Can't load config from file", zap.Error(err))
	}

	log.Info("Loading config from environment")
	err = cleanenv.ReadEnv(&_cfg)
	if err != nil {
		log.Warn("Can't load config from environment", zap.Error(err))
	}

	log.Info("Loading config from flags")
	fset := flag.NewFlagSet("flags", flag.ContinueOnError)
	fset.StringVar(&_cfg.ServerAddr, "a", _cfg.ServerAddr, "address")
	flag.IntVar(&_cfg.Server.SaveInterval, "i", _cfg.Server.SaveInterval, "save interval")
	fset.StringVar(&_cfg.Server.FileStore, "f", _cfg.Server.FileStore, "file storage path")
	fset.BoolVar(&_cfg.Server.Restore, "r", _cfg.Server.Restore, "restore metrics")
	fset.StringVar(&_cfg.Server.DBConn, "d", _cfg.Server.DBConn, "database connection string")
	err = fset.Parse(os.Args[1:])
	if err != nil {
		msg := fmt.Sprintf("Error parsing flags")
		log.Error(msg, zap.Error(err))
		return nil, fmt.Errorf(msg, err)
	}

	return &_cfg, nil
}
