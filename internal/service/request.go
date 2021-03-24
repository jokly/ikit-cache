package service

import (
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

	rs.waitInProgress(requestURL)

	cachedResponse, err := rs.cacheSvc.GetResponse(requestURL)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("couldn't get response from cache for %s: %v", requestURL, err)
		}
	} else {
		log.Printf("got response from cache for %s", requestURL)

		if !cachedResponse.IsError {
			responses <- cachedResponse.Body
		}

		return
	}

	if err := rs.cacheSvc.SetInProgress(requestURL, requestTimeout); err != nil {
		log.Printf("couldn't set in progress for %s: %v", requestURL, err)
	}

	log.Printf("set in progress for %s", requestURL)

	response := Response{}
	body, err := rs.makeRequest(requestURL)
	if err != nil {
		response.Body = err.Error()
		response.IsError = true
	} else {
		response.Body = body
	}

	if err := rs.cacheSvc.SetResponse(requestURL, response, time.Duration(rs.getRandomExpiration())*time.Second); err != nil {
		log.Printf("couldn't set response to cache for %s: %v", requestURL, err)
	}

	log.Printf("set response for %s", requestURL)

	if err := rs.cacheSvc.UnsetInProgress(requestURL); err != nil {
		log.Printf("couldn't unset in progress for %s: %v", requestURL, err)
	}

	log.Printf("unset in progress for %s", requestURL)

	responses <- response.Body
}

func (rs *RequestService) waitInProgress(requestURL string) {
	isInProgress, err := rs.cacheSvc.IsInProgress(requestURL)
	if err != nil {
		log.Printf("couldn't get in progress status for %s: %v", requestURL, err)
	}

	log.Printf("for %s progress is: %t", requestURL, isInProgress)

	if !isInProgress {
		return
	}

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
			return
		case <-ticker.C:
			isInProgress, err := rs.cacheSvc.IsInProgress(requestURL)
			if err != nil {
				log.Printf("couldn't get in progress status for %s: %v", requestURL, err)
			}

			if !isInProgress {
				return
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

func (rs *RequestService) getRandomExpiration() int {
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(rs.config.MaxTimeout-rs.config.MinTimeout+1) + rs.config.MinTimeout
}

func (rs *RequestService) getRandomURL() string {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(rs.config.URLs))

	return rs.config.URLs[randomIndex]
}
