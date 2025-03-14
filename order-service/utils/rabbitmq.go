package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	mu  			sync.Mutex
	rabbitConn		*amqp.Connection
	rabbitChannel	*amqp.Channel
)


func ConnectToRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	var err error
	rabbitConn, rabbitChannel, err = establishConnection()
	if err != nil {
		log.Printf("Initial connection to RabbitMQ failed: %v. Starting reconnection attempts.", err)
		go reconnectToRabbitMQ()
		return nil, nil
	}

	go monitorCloseNotifications(rabbitConn, rabbitChannel)
	return rabbitConn, rabbitChannel
}

func establishConnection() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        return nil, nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, nil, err
    }

    log.Println("Connected to RabbitMQ and channel opened...")
    return conn, ch, nil
}

func monitorCloseNotifications(conn *amqp.Connection, ch *amqp.Channel) {
	connClose := conn.NotifyClose(make(chan *amqp.Error))
	chClose := ch.NotifyClose(make(chan *amqp.Error))

	select {
	case err := <- connClose:
		log.Printf("RabbitMQ connection closed: %v. Starting reconnection...", err)
		go reconnectToRabbitMQ()
	case err := <- chClose:
		log.Printf("RabbitMQ channel closed: %v. Starting reconnection...", err)
		go reconnectToRabbitMQ()
	}
}

func reconnectToRabbitMQ() {
    for {
        time.Sleep(5 * time.Second)

        mu.Lock()
        conn, ch, err := establishConnection()
        if err != nil {
            log.Printf("Reconnection to RabbitMQ failed: %v", err)
            mu.Unlock()
            continue
        }

        if rabbitConn != nil {
            rabbitConn.Close()
        }
        rabbitConn = conn
        rabbitChannel = ch

        log.Println("Successfully reconnected to RabbitMQ.")

        go func() {
            err := DeclareQueue("order_created")
            if err != nil {
                log.Printf("Failed to declare queue after reconnection: %v", err)
            }
        }()

        go monitorCloseNotifications(rabbitConn, rabbitChannel)
        mu.Unlock()
        break
    }
}


func DeclareQueue(queueName string) error {
    for {
        mu.Lock()

        if rabbitChannel == nil {
            mu.Unlock()
            log.Printf("RabbitMQ channel is not available. Retrying to declare queue '%s'...", queueName)
            time.Sleep(5 * time.Second) 
            continue
        }

        _, err := rabbitChannel.QueueDeclare(
            queueName,
            true, 
            false,
            false,
            false,
            nil,  
        )
        if err != nil {
            mu.Unlock()
            log.Printf("Failed to declare queue '%s': %v. Retrying...", queueName, err)
            time.Sleep(5 * time.Second) 
            continue
        }

        log.Printf("Queue '%s' is declared successfully.", queueName)
        mu.Unlock()
        break
    }
    return nil
}

func isRabbitMQActive() bool {
    mu.Lock()
    defer mu.Unlock()

    if rabbitConn == nil || rabbitConn.IsClosed() || rabbitChannel == nil {
        return false
    }
    return true
}

func PublishMessage(queueName string, message interface{}) error {
    for {
        if !isRabbitMQActive() {
            log.Println("RabbitMQ is not active. Waiting for reconnection before publishing message...")
            time.Sleep(5 * time.Second)
            continue
        }

        mu.Lock()

        body, err := json.Marshal(message)
        if err != nil {
            mu.Unlock()
            return fmt.Errorf("failed to serialize message: %v", err)
        }

        err = rabbitChannel.Publish(
            "",        
            queueName,
            false,    
            false,
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            },
        )
        if err != nil {
            log.Printf("Failed to publish message: %v. Retrying...", err)
            mu.Unlock()
            time.Sleep(5 * time.Second)
            continue
        }

        log.Printf("Message published to queue '%s': %s", queueName, string(body))
        mu.Unlock()
        break
    }

    return nil
}
