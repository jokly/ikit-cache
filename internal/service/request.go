package service

import (
	"ikit-cache/internal/util"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	requestTimeout = 5
)

type RequestService struct {
	config *util.Config
	client *http.Client
}

func MakeRequestService(config *util.Config) *RequestService {
	client := &http.Client{
		Timeout: requestTimeout * time.Second,
	}

	return &RequestService{
		config: config,
		client: client,
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
		go rs.makeAsyncRequest(url, responses, wg)
	}

	wg.Wait()
	close(responses)
}

func (rs *RequestService) makeAsyncRequest(requestURL string, responses chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := rs.client.Get(requestURL)
	if err != nil {

		urlErr, ok := err.(*url.Error)
		if ok && urlErr.Timeout() {
			log.Printf("timeout error: %s", requestURL)
		} else {
			log.Printf("couldn't get response from %s: %v", requestURL, err)
		}

		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("couldn't read body of %s: %v", requestURL, err)
	}

	responses <- string(body)
}

func (rs *RequestService) getRandomURL() string {
	randomIndex := rand.Intn(len(rs.config.URLs))

	return rs.config.URLs[randomIndex]
}
