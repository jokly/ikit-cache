package transport

import (
	"ikit-cache/internal/transport/proto"

	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedRandomServiceServer
}

func (s *server) GetRandomDataStream(req *proto.GetRandomDataStreamRequest, stream proto.RandomService_GetRandomDataStreamServer) error {
	resp := &proto.GetRandomDataStreamResponse{
		Result: "Hello World",
	}

	_ = stream.Send(resp)

	return nil
}

func InitGRPCServer() *grpc.Server {
	grpcServer := grpc.NewServer()
	s := &server{}

	proto.RegisterRandomServiceServer(grpcServer, s)

	return grpcServer
}
