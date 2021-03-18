package main

import (
	"ikit-cache/internal/service"
	"ikit-cache/internal/transport"
	"log"
	"net"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cacheSvc := &service.CacheService{}
	grpcServer := transport.InitGRPCServer(cacheSvc)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
