package main

import (
	"flag"
	"fmt"
	"ikit-cache/internal/service"
	"ikit-cache/internal/transport"
	"ikit-cache/internal/util"
	"log"
	"net"
)

func main() {
	port := flag.Int("p", 50051, "server port")
	configPath := flag.String("c", "./config/config.yaml", "path to config file")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	config, err := util.GetConfig(*configPath)
	if err != nil {
		log.Fatalf("couldn't read config: %v", err)
	}

	cacheSvc := service.MakeCacheService(config.RedisURL)
	requestSvc := service.MakeRequestService(config, cacheSvc)
	grpcServer := transport.InitGRPCServer(requestSvc)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
