package service

import "fmt"

type CacheService struct {
}

func (cs *CacheService) GetRandomDataStream() <-chan string {
	responses := make(chan string)

	// TODO: Implement
	go func() {
		for i := 0; i < 5; i++ {
			responses <- fmt.Sprintf("hello: %d", i)
		}

		close(responses)
	}()

	return responses
}
