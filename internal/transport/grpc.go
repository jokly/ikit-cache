package transport

import (
	"ikit-cache/internal/service"
	"ikit-cache/internal/transport/proto"

	"google.golang.org/grpc"
)

type server struct {
	cacheSvc *service.CacheService
	proto.UnimplementedRandomServiceServer
}

func (s *server) GetRandomDataStream(req *proto.GetRandomDataStreamRequest, stream proto.RandomService_GetRandomDataStreamServer) error {
	for resultString := range s.cacheSvc.GetRandomDataStream() {
		resp := &proto.GetRandomDataStreamResponse{
			Result: resultString,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func InitGRPCServer(cacheSvc *service.CacheService) *grpc.Server {
	grpcServer := grpc.NewServer()
	s := &server{
		cacheSvc: cacheSvc,
	}

	proto.RegisterRandomServiceServer(grpcServer, s)

	return grpcServer
}
