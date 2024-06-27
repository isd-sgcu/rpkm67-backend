package pin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	SetPin(key string, code interface{}) error
	GetPin(key string, code interface{}) error
	DeletePin(key string) error
}

type repositoryImpl struct {
	client *redis.Client
}

func NewRepository(client *redis.Client) Repository {
	return &repositoryImpl{client: client}
}

func (r *repositoryImpl) SetPin(key string, pin interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	v, err := json.Marshal(pin)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, pinKey(key), v, 0).Err()
}

func (r *repositoryImpl) GetPin(key string, pin interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	v, err := r.client.Get(ctx, pinKey(key)).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(v), pin)
}

func (r *repositoryImpl) DeletePin(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Del(ctx, pinKey(key)).Err()
}

func pinKey(key string) string {
	return fmt.Sprintf("pin:%s", key)
}
