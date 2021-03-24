package main

import (
	"context"
	"flag"
	"fmt"
	"ikit-cache/internal/transport/proto"
	"io"
	"log"
	"sync"

	"google.golang.org/grpc"
)

func main() {
	host := flag.String("h", "127.0.0.1", "server host")
	port := flag.Int("p", 50051, "server port")
	numConsumers := flag.Int("c", 10, "number of consumers")

	flag.Parse()

	log.Println("before connect")
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *host, *port), grpc.WithInsecure(), grpc.WithBlock())
	log.Println("after connect")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewRandomServiceClient(conn)
	wg := &sync.WaitGroup{}

	wg.Add(*numConsumers)
	for i := 0; i < *numConsumers; i++ {
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
