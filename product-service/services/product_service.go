package services

import (
	"context"
	"fmt"

	"github.com/vaanskii/ecom-ms/product-service/db"
	pb "github.com/vaanskii/ecom-ms/product-service/proto"
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