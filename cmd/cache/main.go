package main

import (
	"ikit-cache/internal/service"
	"ikit-cache/internal/transport"
	"ikit-cache/internal/util"
	"log"
	"net"
)

const (
	port       = ":50051"
	configPath = "./config/config.yaml"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	config, err := util.GetConfig(configPath)
	if err != nil {
		log.Fatalf("couldn't read config: %v", err)
	}

	requestSvc := service.MakeRequestService(config)
	grpcServer := transport.InitGRPCServer(requestSvc)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
