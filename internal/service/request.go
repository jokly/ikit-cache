package service

import (
	"encoding/base64"
	"errors"
	"ikit-cache/internal/util"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	requestTimeout = 5 * time.Second
)

type RequestService struct {
	config   *util.Config
	client   *http.Client
	cacheSvc *CacheService
}

func MakeRequestService(config *util.Config, cacheSvc *CacheService) *RequestService {
	client := &http.Client{
		Timeout: requestTimeout,
	}

	return &RequestService{
		config:   config,
		client:   client,
		cacheSvc: cacheSvc,
	}
}

func (rs *RequestService) GetRandomDataStream() <-chan string {
	responses := make(chan string)

	go rs.makeAsyncRequests(responses)

	return responses
}

func (rs *RequestService) makeAsyncRequests(responses chan<- string) {
	wg := &sync.WaitGroup{}

	wg.Add(rs.config.NumberOfRequests)
	for i := 0; i < rs.config.NumberOfRequests; i++ {
		url := rs.getRandomURL()
		go rs.makeAsyncRequestWithCache(url, responses, wg)
	}

	wg.Wait()
	close(responses)
}

func (rs *RequestService) makeAsyncRequestWithCache(requestURL string, responses chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// read response from cache
		resp, err := rs.cacheSvc.GetResponse(requestURL)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				log.Printf("couldn't get response from cache for %s: %v", requestURL, err)
			}
		} else {
			log.Printf("get response from cache for %s", requestURL)
			if !resp.IsError {
				responses <- resp.Body
			}

			return
		}

		// try get lock
		isTakeLock := false
		lockValue, err := rs.getRandomLockValue()
		if err != nil {
			log.Println("couldn't generate random lock value")
		} else {
			isTakeLock, err = rs.cacheSvc.Lock(requestURL, lockValue, requestTimeout)
			if err != nil {
				log.Printf("couldn't take lock for %s: %v", requestURL, err)
			}
		}

		// wait lock
		if !isTakeLock {
			log.Printf("wait lock for %s", requestURL)
			isGetUnlock := rs.waitLock(requestURL)

			if isGetUnlock {
				log.Printf("get unlock for %s", requestURL)
				continue
			}
		}

		// make HTTP request
		log.Printf("make HTTP request for %s", requestURL)
		response := Response{}
		body, err := rs.makeRequest(requestURL)
		if err != nil {
			response.Body = err.Error()
			response.IsError = true
		} else {
			response.Body = body
		}

		// send response to channel
		if !response.IsError {
			responses <- response.Body
		}

		if isTakeLock {
			// set response to cache
			if err := rs.cacheSvc.SetResponse(requestURL, response, rs.getRandomExpiration()); err != nil {
				log.Printf("couldn't set response to cache for %s: %v", requestURL, err)
			}

			// delete lock
			if err := rs.cacheSvc.Unlock(requestURL, lockValue); err != nil {
				log.Printf("couldn't delete lock for %s: %v", requestURL, err)
			}
		}
	}
}

// true - no lock
// false - don't wait until unlock
func (rs *RequestService) waitLock(requestURL string) bool {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(requestTimeout)
		done <- true
	}()

	for {
		select {
		case <-done:
			return false
		case <-ticker.C:
			isLock, err := rs.cacheSvc.IsLock(requestURL)
			if err != nil {
				log.Printf("couldn't get in progress status for %s: %v", requestURL, err)
			}

			if !isLock {
				return true
			}
		}
	}
}

func (rs *RequestService) makeRequest(requestURL string) (string, error) {
	resp, err := rs.client.Get(requestURL)
	if err != nil {
		urlErr, ok := err.(*url.Error)
		if ok && urlErr.Timeout() {
			log.Printf("timeout error: %s", requestURL)
		} else {
			log.Printf("couldn't get response from %s: %v", requestURL, err)
		}

		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("couldn't read body of %s: %v", requestURL, err)
		return "", err
	}

	return string(body), nil
}

func (rs *RequestService) getRandomLockValue() (string, error) {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func (rs *RequestService) getRandomExpiration() time.Duration {
	rand.Seed(time.Now().UnixNano())

	return time.Duration(rand.Intn(rs.config.MaxTimeout-rs.config.MinTimeout+1)+rs.config.MinTimeout) * time.Second
}

func (rs *RequestService) getRandomURL() string {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(rs.config.URLs))

	return rs.config.URLs[randomIndex]
}
