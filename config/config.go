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

type GroupConfig struct {
	Capacity int
	CacheTTL int
}

type SelectionConfig struct {
	CacheTTL int
}

type PinConfig struct {
	WorkshopCode  string
	WorkshopCount int
	LandmarkCode  string
	LandmarkCount int
}
type Config struct {
	App       AppConfig
	Db        DbConfig
	Redis     RedisConfig
	Group     GroupConfig
	Selection SelectionConfig
	Pin       PinConfig
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

	groupCapacity, err := strconv.ParseInt(os.Getenv("GROUP_CAPACITY"), 10, 64)
	if err != nil {
		return nil, err
	}
	groupCacheTTL, err := strconv.ParseInt(os.Getenv("GROUP_CACHE_TTL"), 10, 64)
	if err != nil {
		return nil, err
	}
	groupConfig := GroupConfig{
		Capacity: int(groupCapacity),
		CacheTTL: int(groupCacheTTL),
	}

	selectionCacheTTL, err := strconv.ParseInt(os.Getenv("SELECTION_CACHE_TTL"), 10, 64)
	if err != nil {
		return nil, err
	}
	selectionConfig := SelectionConfig{
		CacheTTL: int(selectionCacheTTL),
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
		App:       appConfig,
		Db:        dbConfig,
		Redis:     redisConfig,
		Group:     groupConfig,
		Selection: selectionConfig,
		Pin:       pinConfig,
	}, nil
}

func (ac *AppConfig) IsDevelopment() bool {
	return ac.Env == "development"
}
