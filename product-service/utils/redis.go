package utils

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
    RDB      *redis.Client
    rMu       sync.Mutex
    stopChan chan struct{}
)

func InitRedis(ctx context.Context) error {
    stopChan = make(chan struct{})
    if err := establishRedisConnection(); err != nil {
        log.Printf("Initial connection to Redis failed: %v", err)
        go monitorRedis(ctx)
        return err
    }
    log.Println("Initial Redis connection established.")
    go monitorRedis(ctx)
    return nil
}

func establishRedisConnection() error {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    _, err := client.Ping(context.Background()).Result()
    if err != nil {
        return err
    }

    rMu.Lock()
    RDB = client
    rMu.Unlock()

    log.Println("Connected to Redis successfully!")
    return nil
}

func monitorRedis(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("Stopping Redis monitor...")
            return
        case <-ticker.C:
            rMu.Lock()
            err := RDB.Ping(context.Background()).Err()
            rMu.Unlock()
            if err != nil {
                log.Printf("Redis connection lost: %v", err)
                if reconnectErr := establishRedisConnection(); reconnectErr != nil {
                    log.Printf("Reconnection to Redis failed: %v", reconnectErr)
                    continue
                }
                log.Println("Reconnected to Redis successfully!")
            }
        }
    }
}

func CloseRedis() {
    close(stopChan)
    rMu.Lock()
    if RDB != nil {
        if err := RDB.Close(); err != nil {
            log.Printf("Error closing Redis connection: %v", err)
        } else {
            log.Println("Redis connection closed.")
        }
    }
    rMu.Unlock()
}
