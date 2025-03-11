package utils

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectToRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
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