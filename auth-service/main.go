package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/vaanskii/ecommerce-microservices/auth-service/proto"
	"github.com/vaanskii/ecommerce-microservices/auth-service/services"
)

func main() {
	authService := &services.AuthServiceServer{}

	req := &pb.GenerateTokenRequest{
		Username: "Giorgi",
		Password: "2144",
	}

	tokenRes, err := authService.GenerateToken(context.Background(), req)
	if err != nil {
		log.Fatalf("error generating token: %s", err)
	}

	fmt.Println("Generated Token", tokenRes.Token)

	valReq := &pb.ValidateTokenRequest{
		Token: tokenRes.Token,
	}

	valRes, err := authService.ValidateToken(context.Background(), valReq)
	if err != nil {
		log.Fatalf("cannot validate token: %s", err)
	}

	fmt.Println("Token Valid:", valRes.Valid)
	fmt.Println("Username from Token:", valRes.Username)
}
