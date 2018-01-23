package main

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"

	pb "github.com/mhowto/go-example/helloworld"
	"github.com/mhowto/go-example/session"
	"golang.org/x/net/trace"

	timeNowPb "github.com/mhowto/go-example/timenow"
	traceUtil "github.com/mhowto/go-example/trace"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	timeClient timeNowPb.TimerClient
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	var tr trace.Trace
	var ok bool

	if traceUtil.IsTraceEnable() {
		if tr, ok = trace.FromContext(ctx); !ok {
			tr = nil
		}
	}

	if rand.Intn(10) < 2 {
		tr.LazyPrintf("greeter server: random error\n")
		return nil, grpc.Errorf(codes.Internal, "random error")
	}
	sessID := uuid.NewV4().String()
	if traceUtil.IsTraceEnable() {
		tr.LazyPrintf("Greeter: session-id ", sessID)
	}
	ctx = session.NewContextWithSessionId(ctx, sessID)
	t, err := s.timeClient.WhatsTimeNow(ctx, &timeNowPb.WhatsTimeNowRequest{})
	if err != nil {
		return nil, err
	}
	return &pb.HelloReply{Message: "Hello " + in.Name + ". It's " + t.Time + " now."}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// traceUtil.InitGlobalOpenTracer("Greeter", "test")
	go startHttpServer()

	conn, err := grpc.Dial(":50061", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := timeNowPb.NewTimerClient(conn)

	pb.RegisterGreeterServer(s, &server{timeClient: c})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func startHttpServer() {
	srv := &http.Server{
		Addr:        ":8899",
		ReadTimeout: 5 * time.Second,
		// WriteTimeout: 90 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Fail to start http server")
	}
}
