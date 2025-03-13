package utils

import (
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
		log.Printf("RabbitMQ connection closed: %v. Starting reconnection.", err)
		go reconnectToRabbitMQ()
	case err := <- chClose:
		log.Printf("RabbitMQ channel closed: %v. Starting reconnection.", err)
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
		go monitorCloseNotifications(rabbitConn, rabbitChannel) 
		mu.Unlock()
		break
	}
}