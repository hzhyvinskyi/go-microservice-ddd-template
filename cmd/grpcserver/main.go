package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcopentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	appgrpc "github.com/hzhyvinskyi/go-microservice-template/internal/app/application/grpc"
	"github.com/hzhyvinskyi/go-microservice-template/internal/app/application/pb"
	"github.com/hzhyvinskyi/go-microservice-template/internal/app/infrastructure/persistence/dynamo"
)

const (
	awsRegion       = "us-east-1"
	grpcPort        = ":9000"
	certFilePath    = "cert/service.pem"
	keyFilePath     = "cert/service.key"
	shutdownTimeout = time.Second * 10
)

func main() {
	ctx := context.Background()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %s\n", err.Error())
	}
	defer logger.Sync()

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.Fatal("Failed to listen", zap.String("gRPC Port", grpcPort), zap.Error(err))
	}

	dynamoDB := dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})))
	dynamoRepository := dynamo.NewRepository(dynamoDB)
	templateServiceServer := appgrpc.NewTemplateServiceServer(dynamoRepository)

	creds, err := credentials.NewServerTLSFromFile(certFilePath, keyFilePath)
	if err != nil {
		logger.Fatal("Failed to construct TLS credentials for server from given cert and key files", zap.Error(err))
	}

	gRPCServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StreamInterceptor(
			grpcmiddleware.ChainStreamServer(
				grpcrecovery.StreamServerInterceptor(),
				grpcctxtags.StreamServerInterceptor(),
				grpcopentracing.StreamServerInterceptor(),
				grpcprometheus.StreamServerInterceptor,
				grpczap.StreamServerInterceptor(logger),
			),
		),
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				grpcrecovery.UnaryServerInterceptor(),
				grpcctxtags.UnaryServerInterceptor(),
				grpcopentracing.UnaryServerInterceptor(),
				grpcprometheus.UnaryServerInterceptor,
				grpczap.UnaryServerInterceptor(logger),
			),
		),
	)

	pb.RegisterTemplateServiceServer(gRPCServer, templateServiceServer)

	grpcprometheus.Register(gRPCServer)
	http.Handle("/metrics", promhttp.Handler())

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err = gRPCServer.Serve(lis)
		if err != nil {
			logger.Fatal("FailedToServe", zap.String("Listener Addr", lis.Addr().String()), zap.Error(err))
		}
	}()

	logger.Info("gRPC Server is running", zap.String("gRPC Port", grpcPort))

	sig := <-sigC
	logger.Info("OS signal was received. Server is gracefully shutting down...", zap.String("Signal", sig.String()))

	signal.Stop(sigC)

	_, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	gRPCServer.GracefulStop()
}
