package main

import (
	"context"
	"log"
	"time"

	pb "github.com/vaanskii/ecommerce-microservices/order-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect to product-service: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Ikdpb3JnaSIsImV4cCI6MTc0MTYxOTA2M30.B1YTJsp2XfuIXbCBdRxpzh3g00w3VcrV81rVESJ11i8"

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	req := &pb.CreateOrderRequest{
		ProductId: "456",
		Quantity: 2,
		CustomerName: "Giorgi Vanadze",
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := client.CreateOrder(ctx, req)
	if err != nil {
		log.Fatalf("Error creating order: %v", err)
	}

	log.Printf("Order created successfully: ID=%s, Status=%s", res.OrderId, res.Status)
}