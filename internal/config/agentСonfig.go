package config

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	cfgAgent    *Config
	cfgAgentErr error
	onceAgent   sync.Once
)

func GetAgentConfig() (*Config, error) {
	onceAgent.Do(func() {
		cfgAgent, cfgAgentErr = loadAgentConfig()
	})
	return cfgAgent, cfgAgentErr
}

func loadAgentConfig() (*Config, error) {
	var _cfg Config
	var defCfg Config

	log.Info("Loading config from flags")
	fset := flag.NewFlagSet("flags", flag.ContinueOnError)
	fset.StringVar(&_cfg.ServerAddr, "a", _cfg.ServerAddr, "address")
	fset.IntVar(&_cfg.Agent.PollInterval, "p", _cfg.Agent.PollInterval, "poll interval")
	fset.IntVar(&_cfg.Agent.ReportInterval, "r", _cfg.Agent.ReportInterval, "report interval")
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
	if _cfg.Agent.PollInterval == 0 {
		_cfg.Agent.PollInterval = defCfg.Agent.PollInterval
	}
	if _cfg.Agent.ReportInterval == 0 {
		_cfg.Agent.ReportInterval = defCfg.Agent.ReportInterval
	}

	_cfg.ServerAddr = fmt.Sprintf("http://%s", _cfg.ServerAddr)
	return &_cfg, nil
}
