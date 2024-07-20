package database

import (
	"errors"
	"fmt"

	"github.com/isd-sgcu/rpkm67-backend/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(conf *config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	cache := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: conf.Password,
	})

	if cache == nil {
		return nil, errors.New("failed to connect to redis server")
	}

	return cache, nil
}
