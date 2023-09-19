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

func (s *RedisDB) SetPodIPBySubjectAndIP(subject, ip string) error {
	err := s.RedisClient.Set(context.Background(), subject+ip, ip, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisDB) GetPodIPBySubjectAndIP(subject, ip string) (string, error) {
	if cmd := s.RedisClient.Get(context.Background(), subject); cmd.Err() != redis.Nil {
		ip, err := cmd.Result()
		if err != nil {
			return "", redis.TxFailedErr
		}
		return ip, nil
	}
	return "", redis.Nil
}

func (s *RedisDB) GetPodsIPBySubject(subject string) ([]string, error) {
	var ips []string
	keys, err := s.RedisClient.Keys(context.Background(), subject+"*").Result()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		ip, err := s.RedisClient.Get(context.Background(), key).Result()
		if err != nil {
			continue
		}
		ips = append(ips, ip)
	}
	return ips, nil
}
