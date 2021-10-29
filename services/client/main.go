package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/alexeib-ds/grpc-test/services/proto"
	"google.golang.org/grpc"
)

type credsProvider struct {
	token string
}

func (c credsProvider) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %v", c.token),
	}, nil
}

func (c credsProvider) RequireTransportSecurity() bool {
	return false
}

func main() {
	creds := credsProvider{
		token: "JoinTheDarkSide",
	}
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithPerRPCCredentials(creds), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewGreeterClient(conn)
	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "Darth Vader",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response received: %v\n", resp.GetMessage())
}
