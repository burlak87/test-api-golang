package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yml:"env" env-default:"development"`
	StoragePath string `yml:"storage_path" env-required:"true"`
  HTTPServer 				 `yml:"http_server"`
}

type HTTPServer struct {
  Address     string        `yml:"address" env-default:"0.0.0.0:8080"`
  Timeout     time.Duration `yml:"timeout" env-default:"5s"`
  IdleTimeout time.Duration `yml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}
	
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}