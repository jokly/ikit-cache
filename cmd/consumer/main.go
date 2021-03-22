package main

import (
	"context"
	"ikit-cache/internal/transport/proto"
	"io"
	"log"
	"sync"

	"google.golang.org/grpc"
)

const (
	address      = "localhost:50051"
	numConsumers = 10
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewRandomServiceClient(conn)
	wg := &sync.WaitGroup{}

	wg.Add(numConsumers)
	for i := 0; i < numConsumers; i++ {
		go request(wg, client)
	}

	wg.Wait()
}

func request(wg *sync.WaitGroup, client proto.RandomServiceClient) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.GetRandomDataStream(ctx, &proto.GetRandomDataStreamRequest{})

	if err != nil {
		log.Printf("couldn't get stream: %v\n", err)
		return
	}

	i := 0
	for {
		_, err := stream.Recv()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("couldn't get response: %v\n", err)
			break
		}

		log.Println(i)
		i += 1
	}
}
