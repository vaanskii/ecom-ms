package services

import (
	"context"
	"fmt"

	"github.com/vaanskii/ecommerce-microservices/product-service/db"
	pb "github.com/vaanskii/ecommerce-microservices/product-service/proto"
)

type ProductServiceServer struct {
	pb.UnimplementedProductServiceServer
}

func(s *ProductServiceServer) GetProductByID(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	product, err := db.GetProductByID(req.Id)
	if err != nil {
		return nil, fmt.Errorf("product not found: %v", err)
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