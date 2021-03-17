package main

import (
	"context"
	"ikit-cache/internal/transport/proto"
	"log"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewRandomServiceClient(conn)

	request(client)
}

func request(c proto.RandomServiceClient) {
	ctx := context.Background()
	req := &proto.GetRandomDataStreamRequest{}

	stream, err := c.GetRandomDataStream(ctx, req)
	if err != nil {
		log.Printf("couldn't get stream: %v\n", err)
		return
	}

	resp, err := stream.Recv()
	if err != nil {
		log.Printf("couldn't get response: %v\n", err)
		return
	}

	log.Println(resp.Result)
}
