package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	pb "github.com/alexeib-ds/grpc-test/services/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello, %v!", in.GetName()),
	}, nil
}

func validateToken(token string) bool {
	token = strings.TrimPrefix(token, "Bearer ")
	return token == "JoinTheDarkSide"
}

func checkAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	auth := md["authorization"]
	if len(auth) < 1 {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	if !validateToken(auth[0]) {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return handler(ctx, req)
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(checkAuth))
	pb.RegisterGreeterServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	mux := runtime.NewServeMux()
	pb.RegisterGreeterHandler(context.Background(), mux, conn)
	gwServer := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	gwServer.ListenAndServe()
}
