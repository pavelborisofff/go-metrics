package config

import (
	"flag"
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
	var defCfg Config

	log.Info("Loading config from flags")
	fset := flag.NewFlagSet("flags", flag.ContinueOnError)
	fset.StringVar(&_cfg.ServerAddr, "a", _cfg.ServerAddr, "address")
	fset.IntVar(&_cfg.Server.SaveInterval, "i", _cfg.Server.SaveInterval, "save interval")
	fset.StringVar(&_cfg.Server.FileStore, "f", _cfg.Server.FileStore, "file storage path")
	fset.BoolVar(&_cfg.Server.Restore, "r", _cfg.Server.Restore, "restore metrics")
	fset.StringVar(&_cfg.Server.DBConn, "d", _cfg.Server.DBConn, "database connection string")
	err := fset.Parse(os.Args[1:])
	if err != nil {
		log.Error("Error parsing flags", zap.Error(err))
		return nil, err
	}

	log.Info("Loading config from environment")
	err = cleanenv.ReadEnv(&_cfg)
	if err != nil {
		log.Warn("Can't load config from environment", zap.Error(err))
	}

	log.Info("Loading default config from file", zap.String("file", envFile))
	err = cleanenv.ReadConfig(envFile, &defCfg)
	if err != nil {
		log.Warn("Can't load config from file", zap.Error(err))
	}

	if _cfg.ServerAddr == "" {
		_cfg.ServerAddr = defCfg.ServerAddr
	}
	if _cfg.Server.SaveInterval == 0 {
		_cfg.Server.SaveInterval = defCfg.Server.SaveInterval
	}
	if _cfg.Server.FileStore == "" {
		_cfg.Server.FileStore = defCfg.Server.FileStore
	}
	if !_cfg.Server.Restore {
		_cfg.Server.Restore = defCfg.Server.Restore
	}
	if _cfg.Server.DBConn == "" {
		_cfg.Server.DBConn = defCfg.Server.DBConn
	}

	return &_cfg, nil
}
