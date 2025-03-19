package utils

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	mu 				sync.Mutex
	rabbitConn		*amqp.Connection
	rabbitChannel	*amqp.Channel
)

type Order struct {
	OrderID      string `json:"OrderID"`
    ProductID    string `json:"ProductID"`
    Quantity     int32  `json:"Quantity"`
    CustomerName string `json:"CustomerName"`
    Status       string `json:"Status"`
}

func ConnectToRabbitMQ(processFunc func(Order)) (*amqp.Connection, *amqp.Channel, error) {
	var err error
	rabbitConn, rabbitChannel, err = establishConnection()
	if err != nil {
		log.Printf("Initial connection to RabbitMQ failed: %v. Starting reconnection...", err)
		go reconnectToRabbitMQ(processFunc)
		return nil, nil, err
	}

	go monitorCloseNotifications(rabbitConn, rabbitChannel, processFunc)
	return rabbitConn, rabbitChannel, nil
}

func establishConnection() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	log.Println("Connected to RabbitMQ and channel opened...")
	return conn, ch, nil
}

func isRabbitMQActive() bool {
	if rabbitConn == nil || rabbitConn.IsClosed() || rabbitChannel == nil {
		return false
	} 

	return true
}

func ConsumeOrders(queueName string, processFunc func(Order)) {
	for {
		if !isRabbitMQActive() {
			log.Println("RabbitMQ is not active. waiting for reconnection...")
			time.Sleep(5 * time.Second)
			continue
		}

		mu.Lock()
		defer mu.Unlock()

		q, err := rabbitChannel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("Failed to declare a queue: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
	
		msgs, err := rabbitChannel.Consume(
			q.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("Failed to register a consumer: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		
		log.Printf("Listening for messages on queue: %s...", queueName)
	
		go func() {
			for d := range msgs{
				var order Order
				if err := json.Unmarshal(d.Body, &order); err != nil {
					log.Printf("Error decoding message: %v", err)
					continue
				}
				processFunc(order)
			}
		}()

		break
	}
}

func CloseRabbitMQ() {
    if rabbitChannel != nil {
        rabbitChannel.Close()
    }
    if rabbitConn != nil {
        rabbitConn.Close()
    }
    log.Println("RabbitMQ connection closed.")
}

func monitorCloseNotifications(conn *amqp.Connection, ch *amqp.Channel, processFunc func(Order)) {
	connClose := conn.NotifyClose(make(chan *amqp.Error))
	chClose := ch.NotifyClose(make(chan *amqp.Error))

	select {
	case err := <- connClose:
		log.Printf("RabbitMQ connection closed: %v. Starting reconnection...", err)
		go reconnectToRabbitMQ(processFunc)
	case err := <- chClose:
		log.Printf("Channel connection closed: %v, Starting reconnection...", err)
		go reconnectToRabbitMQ(processFunc)
	}
}

func reconnectToRabbitMQ(processFunc func(Order)) {
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

		go monitorCloseNotifications(rabbitConn, rabbitChannel, processFunc)

		go ConsumeOrders("order_created", processFunc)

		mu.Unlock()
		break
	}
}