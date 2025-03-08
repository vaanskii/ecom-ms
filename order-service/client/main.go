package main

import (
	"context"
	"log"
	"time"

	pb "github.com/vaanskii/ecommerce-microservices/order-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect to product-service: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	req := &pb.CreateOrderRequest{
		ProductId: "123",
		Quantity: 2,
		CustomerName: "Giorgi Vanadze",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateOrder(ctx, req)
	if err != nil {
		log.Fatalf("Error creating order: %v", err)
	}

	log.Printf("Order created successfully: ID=%s, Status=%s", res.OrderId, res.Status)
}