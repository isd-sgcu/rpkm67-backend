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

type PinConfig struct {
	WorkshopCode  string
	WorkshopCount int
	LandmarkCode  string
	LandmarkCount int
}
type Config struct {
	App   AppConfig
	Db    DbConfig
	Redis RedisConfig
	Pin   PinConfig
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

	workshopCount, err := strconv.ParseInt(os.Getenv("PIN_WORKSHOP_COUNT"), 10, 64)
	if err != nil {
		return nil, err
	}
	landmarkCount, err := strconv.ParseInt(os.Getenv("PIN_LANDMARK_COUNT"), 10, 64)
	if err != nil {
		return nil, err
	}

	pinConfig := PinConfig{
		WorkshopCode:  os.Getenv("PIN_WORKSHOP_CODE"),
		WorkshopCount: int(workshopCount),
		LandmarkCode:  os.Getenv("PIN_LANDMARK_CODE"),
		LandmarkCount: int(landmarkCount),
	}

	return &Config{
		App:   appConfig,
		Db:    dbConfig,
		Redis: redisConfig,
		Pin:   pinConfig,
	}, nil
}

func (ac *AppConfig) IsDevelopment() bool {
	return ac.Env == "development"
}
