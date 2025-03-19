package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vaanskii/ecommerce-microservices/product-service/db"
	pb "github.com/vaanskii/ecommerce-microservices/product-service/proto"
	"github.com/vaanskii/ecommerce-microservices/product-service/utils"
)

type ProductServiceServer struct {
	pb.UnimplementedProductServiceServer
}

func(s *ProductServiceServer) GetProductByID(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	cachedProduct, err := utils.RDB.Get(ctx, "product:"+req.Id).Result()
    if err == redis.Nil {
        log.Println("Cache miss: Key not found in Redis")
    } else if err != nil {
        return nil, fmt.Errorf("redis error: %v", err)
    } else {
        var product db.Product
        if err := json.Unmarshal([]byte(cachedProduct), &product); err != nil {
            return nil, fmt.Errorf("cache error: %v", err)
        }
        log.Println("Cache hit: Returning product from Redis", cachedProduct)
        return &pb.ProductResponse{
            Id:    product.ID,
            Name:  product.Name,
            Price: product.Price,
        }, nil
    }

	product, err := db.GetProductByID(req.Id)
	if err != nil {
		return nil, fmt.Errorf("product not found: %v", err)
	}

	productJSON, err := json.Marshal(product)
	if err == nil {
        err = utils.RDB.Set(ctx, "product:"+req.Id, productJSON, 10*time.Minute).Err()
        if err != nil {
            log.Printf("Failed to cache product in Redis: %v", err)
        } else {
            log.Println("Cache miss: Product stored in Redis")
        }
    }

	return &pb.ProductResponse{
		Id: product.ID,
		Name: product.Name,
		Price: product.Price,
	}, nil
}

func (s *ProductServiceServer) GetAllProducts(ctx context.Context, req *pb.EmptyRequest) (*pb.ProductListResponse, error) {
	products, err := db.GetAllProducts()
	if err != nil {
		return nil, fmt.Errorf("no products found: %v", err)
	}

	productResponse := []*pb.ProductResponse{}
	for _, product := range products {
		productResponse = append(productResponse, &pb.ProductResponse{
			Id: product.ID,
			Name: product.Name,
			Price: product.Price,
		})
	}

	return &pb.ProductListResponse{Products: productResponse}, nil
}

func (s *ProductServiceServer) UpdateProductQuantity(ctx context.Context, req *pb.UpdateQuantityRequest) (*pb.UpdateQuantityResponse, error) {
	DB := db.GetDBInstance()
	product, err := db.GetProductByID(req.Id)
	if err != nil {
		return nil, fmt.Errorf("product not found, %v", err)
	}

	if product.Quantity < req.Quantity {
		return &pb.UpdateQuantityResponse{Success: false}, nil
	}

	product.Quantity -= req.Quantity 
	if err := DB.Save(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to update product quantity: %v", err)
	}

	return &pb.UpdateQuantityResponse{Success: true}, nil
}