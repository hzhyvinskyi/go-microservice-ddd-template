package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"github.com/hzhyvinskyi/go-microservice-template/internal/app/application/pb"
)

const gRPCPort = ":9000"

func main() {
	gRPCConn, err := grpc.Dial(gRPCPort, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial: %s\n", gRPCPort)
	}
	defer gRPCConn.Close()

	templateServiceClient := pb.NewTemplateServiceClient(gRPCConn)

	response, err := templateServiceClient.Get(context.Background(), &pb.GetTemplateReq{Id: "1x"})
	if err != nil {
		log.Fatalf("Failed to get template: %s\n", "1x")
	}

	log.Printf("RESP TEMPLATE: %v\n", response.Template)
}
