package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App
	Log
	Postgres
	HTTP
	Token
}

type Token struct {
	TelegramBotToken string `env:"TOKEN_TELEGRAM"`
}

type App struct {
	Name       string `env:"APP_NAME" env-default:"ailingo-backend"`
	Version    string `env:"APP_VERSION" env-default:"0.1.0"`
	TargetSite string `env:"TARGET_SITE" env-default:"https://www.some-site.com"`
}

type Log struct {
	Level string `env:"LOG_LEVEL" env-default:"debug"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-default:"postgres"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Pass     string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Database string `env:"POSTGRES_DB" env-default:"bazacars"`
}

type HTTP struct {
	Port int    `env:"HTTP_PORT" env-default:"8080"`
	Host string `env:"HTTP_HOST" env-default:""`
}

// New returns app config
func New() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
