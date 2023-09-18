package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisDB struct {
	RedisClient *redis.Client
}

func NewModule() *RedisDB {
	return &RedisDB{
		RedisClient: redis.NewClient(&redis.Options{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
		}),
	}
}

func (s *RedisDB) SetPodIPBySubject(subject, ip string) error {
	err := s.RedisClient.Set(context.Background(), subject, ip, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisDB) GetPodIPBySubject(subject string) (string, error) {
	ip, err := s.RedisClient.Get(context.Background(), subject).Result()
	if err != nil {
		return "", err
	}
	return ip, nil
}
