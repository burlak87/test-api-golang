package config

import (
	"gosmol/pkg/logging"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yml:"env" env-default:"development"`
	StorageConfig
}

// type HTTPServer struct {
//   Address     string        `yml:"address" env-default:"0.0.0.0:8080"`
//   Timeout     time.Duration `yml:"timeout" env-default:"5s"`
//   IdleTimeout time.Duration `yml:"idle_timeout" env-default:"60s"`
// }

type StorageConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var instance *Config
var once sync.Once

// func MustLoad() *Config {
// 	configPath := os.Getenv("CONFIG_PATH")

// 	if configPath == "" {
// 		log.Fatal("CONFIG_PATH environment variable is not set")
// 	}

// 	if _, err := os.Stat(configPath); err != nil {
// 		log.Fatalf("error opening config file: %s", err)
// 	}
	
// 	var cfg Config

// 	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
// 		log.Fatalf("error reading config file: %s", err)
// 	}

// 	return &cfg
// }

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application configuration")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}