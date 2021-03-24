package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

/* cache strategy (redlock)

(x) get response from cache:
	1. exists -> send response to channel
	2. not exists -> set lock (set nx)
		1. successfull -> http request -> send response to channel -> add cache -> unset lock
		2. fail (lock exist) -> wait until unlock -> go to (x)
*/

type Response struct {
	Body    string `json:"response"`
	IsError bool   `json:"is_error"`
}

const (
	lockKeySuffix = ":lock"
)

var (
	ctx = context.Background()

	deleteScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
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

func (cs *CacheService) IsLock(url string) (bool, error) {
	cmd := cs.rdb.Get(ctx, cs.getLockKey(url))
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return false, nil
		}

		return false, cmd.Err()
	}

	return true, nil
}

func (cs *CacheService) Lock(url, value string, expiration time.Duration) (bool, error) {
	cmd := cs.rdb.SetNX(ctx, cs.getLockKey(url), value, expiration)

	return cmd.Val(), cmd.Err()
}

func (cs *CacheService) Unlock(url, value string) error {
	cmd := cs.rdb.Eval(
		ctx,
		deleteScript,
		[]string{cs.getLockKey(url)},
		value,
	)

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

func (cs *CacheService) getLockKey(url string) string {
	return url + lockKeySuffix
}
