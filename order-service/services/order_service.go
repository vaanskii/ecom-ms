package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/vaanskii/ecommerce-microservices/order-service/proto"
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
	Quantity 		int32
	CustomerName	string
	Status  		string
}

var (
	orders = make(map[string]Order)
	mu sync.Mutex
)

func SaveOrder(order Order) {
	mu.Lock()
	defer mu.Unlock()
	orders[order.OrderID] = order
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

func GetOrderByID(orderID string) (*Order, error) {
	mu.Lock()
	defer mu.Unlock()

	order, exists := orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}
	return &order, nil
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

	SaveOrder(order)

	return &pb.CreateOrderResponse{
		OrderId: order.OrderID,
		Status: order.Status,
	}, nil
}