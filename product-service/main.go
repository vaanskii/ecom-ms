package main

import (
	"log"
	"net"

	"github.com/vaanskii/ecommerce-microservices/product-service/db"
	pb "github.com/vaanskii/ecommerce-microservices/product-service/proto"
	services "github.com/vaanskii/ecommerce-microservices/product-service/services"
	"google.golang.org/grpc"
)


func main() {
	db.SetupDatabase()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, &services.ProductServiceServer{})
	log.Println("Product gRPC Server running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}