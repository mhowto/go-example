package main

import (
	"log"
	"net"
	"time"

	pb "github.com/mhowto/go-example/timenow"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50061"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	timeClient timeNowPb.TimerClient
}

// SayHello implements helloworld.GreeterServer
func (s *server) WhatsTimeNow(ctx context.Context, in *pb.WhatsTimeNowRequest) (*pb.WhatsTimeNowReply, error) {
	now := time.Now()

	return &pb.WhatsTimeNowReply{Time: now.Format("2006-01-02 15:04:05")}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTimerServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
