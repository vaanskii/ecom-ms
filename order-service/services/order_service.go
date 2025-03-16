package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	OrderDB "github.com/vaanskii/ecommerce-microservices/order-service/db"
	pb "github.com/vaanskii/ecommerce-microservices/order-service/proto"
	"github.com/vaanskii/ecommerce-microservices/order-service/utils"
	pbProduct "github.com/vaanskii/ecommerce-microservices/product-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
}

type Order struct {
	OrderID       	string  
	ProductID 		string
	CustomerName	string
	Quantity 		int32
	Status  		string
}

func SaveOrder(order Order) error {
	db := OrderDB.GetDBInstance()
	newOrder := OrderDB.Orders{
		OrderID:   order.OrderID,
		ProductID: order.ProductID,
		CustomerName: order.CustomerName,
		Quantity: order.Quantity,
		Status: order.Status,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&newOrder).Error; err != nil {
		log.Printf("failed to save order in database: %v", err)
	}
	log.Printf("order saved to the database: %v", newOrder)
	return nil
}

func GetProductByID(productID string) (*pbProduct.ProductResponse, error) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect to product-service: %v", err)
        return nil, err
	}
	defer conn.Close()

	client := pbProduct.NewProductServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second) 
	defer cancel()

	return client.GetProductByID(ctx, &pbProduct.ProductRequest{Id: productID})
}

func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req.ProductId == "" || req.Quantity <= 0 {
		return nil, fmt.Errorf("invalid input: product_id and quantity must be valid")
	}

	product, err := GetProductByID(req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("product not found: %v", err)
	}

	log.Printf("Ordering product: %s - $%.2f", product.Name, product.Price)

	order := Order{
		OrderID: uuid.New().String(),
		ProductID: req.ProductId,
		Quantity:  req.Quantity,
		CustomerName: req.CustomerName,
		Status: "Order Created",
	}
	if err := SaveOrder(order); err != nil {
        return nil, fmt.Errorf("failed to save order: %v", err)
    }
	
	err = utils.PublishMessage("order_created", order)
	if err != nil {
        log.Printf("Failed to publish order to RabbitMQ: %v", err)
    }

	return &pb.CreateOrderResponse{
		OrderId: order.OrderID,
		Status: order.Status,
	}, nil
}