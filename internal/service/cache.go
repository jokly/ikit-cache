package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Response struct {
	Body    string `json:"response"`
	IsError bool   `json:"is_error"`
}

const (
	inProgressKeySuffix = ":in_progress"
	inProgressStatus    = 1
)

var (
	ctx = context.Background()
)

type CacheService struct {
	rdb *redis.Client
}

func MakeCacheService(redisURL string) *CacheService {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("couldn't parse redis URL: %v", err)
	}

	return &CacheService{
		rdb: redis.NewClient(opt),
	}
}

func (cs *CacheService) IsInProgress(url string) (bool, error) {
	res, err := cs.rdb.Get(ctx, cs.getInProgressKey(url)).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, err
	}

	return res == inProgressStatus, nil
}

func (cs *CacheService) SetInProgress(url string, expiration time.Duration) error {
	cmd := cs.rdb.SetEX(ctx, cs.getInProgressKey(url), inProgressStatus, expiration)

	return cmd.Err()
}

func (cs *CacheService) UnsetInProgress(url string) error {
	cmd := cs.rdb.Del(ctx, cs.getInProgressKey(url))

	return cmd.Err()
}

func (cs *CacheService) GetResponse(url string) (Response, error) {
	respJSON, err := cs.rdb.Get(ctx, url).Result()

	if err != nil {
		return Response{}, err
	}

	resp := Response{}
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func (cs *CacheService) SetResponse(url string, response Response, expiration time.Duration) error {
	respJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cmd := cs.rdb.SetEX(ctx, url, respJSON, expiration)

	return cmd.Err()
}

func (cs *CacheService) getInProgressKey(url string) string {
	return url + inProgressKeySuffix
}
