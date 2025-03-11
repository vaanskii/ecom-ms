package main

import (
	"log"
	"net"

	"github.com/vaanskii/ecommerce-microservices/order-service/middleware"
	pb "github.com/vaanskii/ecommerce-microservices/order-service/proto"
	"github.com/vaanskii/ecommerce-microservices/order-service/services"
	"github.com/vaanskii/ecommerce-microservices/order-service/utils"
	"google.golang.org/grpc"
)

func main() {
	conn, ch, err := utils.ConnectToRabbitMQ()
	if err != nil {
		log.Fatalf("cannot connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()


	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on port 50052: %v", err)
	}

	grpcServer := grpc.NewServer(
	grpc.UnaryInterceptor(middleware.UnaryAuthInterceptor),
	)

	orderService := &services.OrderServiceServer{}
	pb.RegisterOrderServiceServer(grpcServer, orderService)

	log.Println("Order service is running on port 50052...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
