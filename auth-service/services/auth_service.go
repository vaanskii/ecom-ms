package services

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	pb "github.com/vaanskii/ecommerce-microservices/auth-service/proto"
)

var jwtSecretKey []byte

func init() {
	if err := godotenv.Load("../.env"); err != nil {
        log.Fatalf("Error loading .env file %v", err)
	}

	jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
}

type Claims struct {
	Username     string     `json:"username"`
	jwt.RegisteredClaims
}

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthServiceServer) GenerateToken(ctx context.Context, req *pb.GenerateTokenRequest) (*pb.GenerateTokenResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username or password cannot be empty")
	}

	claims := &Claims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &pb.GenerateTokenResponse{Token: tokenString}, nil
}

func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	token, err := jwt.ParseWithClaims(req.Token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecretKey, nil
	})
	if err != nil || !token.Valid {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	claims, ok := token.Claims.(*Claims) 
	if !ok {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		Username: claims.Username,
	}, nil
}