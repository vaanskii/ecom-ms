package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/vaanskii/ecommerce-microservices/product-service/db"
	pb "github.com/vaanskii/ecommerce-microservices/product-service/proto"
	services "github.com/vaanskii/ecommerce-microservices/product-service/services"
	"github.com/vaanskii/ecommerce-microservices/product-service/utils"
	"google.golang.org/grpc"
)


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func (){
		<-signalChan
		log.Println("Shutting down gracefully...")
        cancel()
	}()

	if err := db.SetupDatabase(); err != nil {
        log.Fatalf("Failed to set up database: %v", err)
    }
	defer db.CloseDatabase()

	if err := utils.InitRedis(ctx); err != nil {
		log.Fatalf("Failed to set up database: %v", err )
	}
	defer utils.CloseRedis()

	processFunc := func(order utils.Order) {
		product, err := db.GetProductByID(order.ProductID)
		if err != nil {
            log.Printf("Product not found for order: %v", err)
            return
        }
		log.Printf("Processing Order: %+v. Product: %+v", order, product)
	}

	utils.ConnectToRabbitMQ(processFunc)
	defer utils.CloseRabbitMQ()


	go utils.ConsumeOrders("order_created", processFunc)

	listener, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterProductServiceServer(grpcServer, &services.ProductServiceServer{})

    go func() {
        log.Println("Product gRPC Server running on port 50051...")
        if err := grpcServer.Serve(listener); err != nil {
            log.Printf("Failed to serve: %v", err)
            cancel()
        }
    }()
	<-ctx.Done()
	log.Println("Server has shut down.")
}