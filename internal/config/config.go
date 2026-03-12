package config

import (
	"TestHitalent2/pkg/postgres"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Host     string `yaml:"host" env:"HOST" env-default:"0.0.0.0"`
	Port     string `yaml:"port" env:"PORT" env-default:"4047"`
	Postgres postgres.Config
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load(".env")

	var cfg Config
	err := cleanenv.ReadConfig("./config/config.yaml", &cfg)
	if err != nil {
		return nil, err
	}
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
