package middleware

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return nil, errors.New("missing authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing methods")
		}
		return jwtSecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return handler(ctx, req)
}