package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	Env         string `yaml:"env" default:"local"`
	HttpConfig  `yaml:"http_server"`
	RedisConfig `yaml:"redis"`
}

type HttpConfig struct {
	Port         int16         `yaml:"port" default:"8080"`
	WriteTimeout time.Duration `yaml:"write_timeout" default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" default:"10s"`
}

type RedisConfig struct {
	Addr        string        `yaml:"host" required:"true"`
	Db          int16         `yaml:"db" default:"0"`
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"10s"`
	Timeout     time.Duration `yaml:"timeout" default:"5s"`
	MaxRetries  int           `yaml:"max_retries" default:"1"`
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		return nil, fmt.Errorf("env variable CONFIG_PATH not set")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
