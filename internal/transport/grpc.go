package transport

import (
	"ikit-cache/internal/service"
	"ikit-cache/internal/transport/proto"

	"google.golang.org/grpc"
)

type server struct {
	requestSvc *service.RequestService
	proto.UnimplementedRandomServiceServer
}

func (s *server) GetRandomDataStream(req *proto.GetRandomDataStreamRequest, stream proto.RandomService_GetRandomDataStreamServer) error {
	for resultString := range s.requestSvc.GetRandomDataStream() {
		resp := &proto.GetRandomDataStreamResponse{
			Result: resultString,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func InitGRPCServer(requestSvc *service.RequestService) *grpc.Server {
	grpcServer := grpc.NewServer()
	s := &server{
		requestSvc: requestSvc,
	}

	proto.RegisterRandomServiceServer(grpcServer, s)

	return grpcServer
}
