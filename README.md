# Ecommerce Microservices Project

This project implements a microservices architecture for an ecommerce platform, consisting of **Product**, **Order**, **Auth**, and services. The services are implemented in Go, using **gRPC** for communication, **RabbitMQ** for message brokering, and **Redis** for caching.

---

## **Project Structure**

```text
ecommerce-microservices/
├── auth-service/
├── docker-compose.yml
├── go.mod
├── go.sum
├── order-service/
├── product-service/
└── README.md
```
---

Prerequisites To work with this project, ensure the following tools are installed on your system:

[Go](https://go.dev/): Minimum version 1.18.

[Docker](https://docs.docker.com/): Required to build and run services in containers.

Protocol Buffers Compiler (protoc): Used to generate [gRPC](https://grpc.io/docs/languages/go/quickstart/) code for each service.

---

### Proto Compilation
Each service uses .proto files for defining gRPC communication. Use the following commands to generate the required Go files:

Each service uses .proto files for defining gRPC communication. Use the following commands to generate the required Go files:

#### Product Service
Navigate to the product-service directory and run:
```bash
protoc --go_out=. --go-grpc_out=. proto/product.proto
```
#### Order Service
Navigate to the order-service directory and run:
```bash 
protoc --go_out=. --go-grpc_out=. proto/order.proto
```

#### Auth Service
Navigate to the auth-service directory and run:
```bash
protoc --go_out=. --go-grpc_out=. proto/auth.proto
```

### Using Docker to Build and Run
The project includes a docker-compose.yml file that simplifies running the services and their dependencies.

#### Start All Services
Run the following command to start Redis, RabbitMQ, and all Go services:
```bash 
docker-compose up --build
```
---
### Running Servers Locally
For development and debugging purposes, you can run the services locally without Docker. Open separate terminal tabs for each service.


#### Run Product Service
```bash
cd product-service
go run main.go
```

#### After runing u should get back:
``` text 
➜  product-service: go run main.go
Database connected successfully!
Connected to Redis successfully!
Initial Redis connection established.
Connected to RabbitMQ and channel opened...
Product gRPC Server running on port 50051...
Listening for messages on queue: order_created...
```
---

#### Run Order Service
```bash
cd order-service
go run main.go
```
#### After runing u should get back:
``` text 
➜  order-service: go run main.go
Database connected successfully!
Connected to RabbitMQ and channel opened...
Order service is running on port 50052...
Queue 'order_created' is declared successfully.

```
---

#### Run Auth Service
```bash
cd auth-service
go run main.go
```
#### After runing u should get back:
``` text 
auth-service: go run main.go
Generated Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Ikdpb3JnaSIsImV4cCI6MTc0MjQ3NDk2OX0.QBNfn4VuYZkv2oDn3ElQdV7W1oMpeb-G7a5o96mBIJU
Token Valid: true
Username from Token: Username
```
---

### Message Broker: RabbitMQ
```text 
RabbitMQ is used for message brokering:

Exchange: order_events

Queue: order_created

RabbitMQ is exposed on:

Port 5672 for internal service communication.

Port 15672 for management UI (accessible at http://localhost:15672).
```


### Cache: Redis
```text
Redis is used for caching frequently accessed product data in the product-service. It is exposed on port 6379.
```

### Testing
#### Product-Service:

Test Redis caching with product queries.

Verify cache hits and misses in the logs.

#### Order-Service:

Use order-client to test gRPC communication with the product-service and order-service.
```bash 
cd order-service/client
go run main.go
```

#### After runing u should get back:
```text
order-service: ✗ go run main.go
2025/03/19 16:54:28 Order created successfully: ID=6dde6258-c70b-4426-a578-cd77959f0945, Status=Order Created
```
#### in product-service terminal you should get:
```text
Cache hit: Returning product from Redis {"ID":"2","Name":"TV","Price":799.99,"Quantity":72}

2025/03/19 16:55:33 Processing Order: {OrderID:a4d8d2f5-d9f0-4b44-9387-a49959cfdaca ProductID:2 Quantity:1 CustomerName:Jack Marston Status:Order Created}. 
Product: &{ID:2 Name:TV Price:799.99 Quantity:70}
```

#### in order-service terminal you should get:
```text
Ordering product: TV - $799.99

2025/03/19 16:55:33 order saved to the database: 

{a4d8d2f5-d9f0-4b44-9387-a49959cfdaca 2 Jack Marston 1 Order Created 2025-03-19 16:55:33.047487 +0400 +04 m=+424.348445460}

2025/03/19 16:55:33 Message published to queue 'order_created': {"OrderID":"a4d8d2f5-d9f0-4b44-9387-a49959cfdaca","ProductID":"2","CustomerName":"Jack Marston","Quantity":1,"Status":"Order Created"}
```