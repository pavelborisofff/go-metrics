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
	fset.IntVar(&_cfg.Agent.PollInterval, "p", _cfg.Agent.PollInterval, "poll interval")
	fset.IntVar(&_cfg.Agent.ReportInterval, "r", _cfg.Agent.ReportInterval, "report interval")
	err = fset.Parse(os.Args[1:])
	if err != nil {
		msg := fmt.Sprintf("Error parsing flags")
		log.Error(msg, zap.Error(err))
		return nil, fmt.Errorf(msg, err)
	}

	_cfg.ServerAddr = fmt.Sprintf("http://%s", _cfg.ServerAddr)
	return &_cfg, nil
}
