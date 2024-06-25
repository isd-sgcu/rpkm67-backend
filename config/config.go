package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port string
	Env  string
}

type DbConfig struct {
	Url string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type Config struct {
	App   AppConfig
	Db    DbConfig
	Redis RedisConfig
}

func LoadConfig() (*Config, error) {
	if os.Getenv("APP_ENV") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			return nil, err
		}
	}

	appConfig := AppConfig{
		Port: os.Getenv("APP_PORT"),
		Env:  os.Getenv("APP_ENV"),
	}

	dbConfig := DbConfig{
		Url: os.Getenv("DB_URL"),
	}

	redisPort, err := strconv.ParseInt(os.Getenv("REDIS_PORT"), 10, 64)
	if err != nil {
		return nil, err
	}

	redisConfig := RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     int(redisPort),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

	return &Config{
		App:   appConfig,
		Db:    dbConfig,
		Redis: redisConfig,
	}, nil
}

func (ac *AppConfig) IsDevelopment() bool {
	return ac.Env == "development"
}
