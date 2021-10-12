package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	firebase "firebase.google.com/go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/langa-me/langame-djinn/internal/config"
	"github.com/langa-me/langame-djinn/internal/djinn"
	"github.com/langa-me/langame-djinn/internal/server"
	"google.golang.org/grpc"
)

func main() {
	err := config.InitConfig(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}
	log.Printf("Config loaded")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	grpcEndpoint := fmt.Sprintf(":%s", port)
	log.Printf("gRPC endpoint [%s]", grpcEndpoint)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(server.AuthFunc)),
	)
	// Override the default gRPC server with our unauthenticated server
	if len(os.Args) > 2 && os.Args[2] == "no-auth" {
		grpcServer = grpc.NewServer()
		log.Printf("Authentication is disabled")
	}
	ctx := context.Background()
	fb, err := firebase.NewApp(ctx, nil)
	server.App = fb
	if err != nil {
		panic(fmt.Sprintf("Failed to init firebase: %v", err))
	}
	djinn.RegisterConversationMagnifierServer(grpcServer, server.NewServer())

	listen, err := net.Listen("tcp", grpcEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting: gRPC Listener [%s]\n", grpcEndpoint)
	log.Fatal(grpcServer.Serve(listen))
}