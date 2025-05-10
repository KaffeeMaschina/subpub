package main

import (
	"context"
	"flag"
	"github.com/KaffeeMaschina/subpub/config"
	"github.com/KaffeeMaschina/subpub/config/env"
	"github.com/KaffeeMaschina/subpub/mysubpub"
	desc "github.com/KaffeeMaschina/subpub/pkg/bus_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	sp mysubpub.SubPub
	desc.UnimplementedPubSubServer
}

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}
func (s *server) Subscribe(req *desc.SubscribeRequest, stream desc.PubSub_SubscribeServer) error {
	_, err := s.sp.Subscribe(req.Key, func(msg interface{}) {
		dataStr, ok := msg.(string)
		if !ok {
			log.Println("invalid message type")
			return
		}
		err := stream.Send(&desc.Event{Data: dataStr})
		if err != nil {
			log.Printf("stream send error: %v", err)
		}
	})
	if err != nil {
		return err
	}

	<-stream.Context().Done()
	return nil
}
func (s *server) Publish(ctx context.Context, req *desc.PublishRequest) (*emptypb.Empty, error) {
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	err := s.sp.Publish(req.Key, req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish: %v", err)
	}

	return &emptypb.Empty{}, nil
}
func setupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()
	return ctx
}
func main() {
	ctx := setupSignalHandler()

	flag.Parse()
	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	grpcConfig, err := env.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sp := mysubpub.NewSubPub()

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterPubSubServer(s, &server{
		sp: sp,
	})
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	s.GracefulStop()
	err = sp.Close(context.Background())
	if err != nil {
		log.Fatalf("failed to shutdown: %v", err)
	}
}
