package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"rabbit-mq-with-go/internal/models"
	"rabbit-mq-with-go/internal/rabbitmq"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchangeName = "orders.exchange"
	exchangeType = "topic" // topic exchange로 유연한 라우팅
)

func main() {
	// RabbitMQ 연결
	conn, err := rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		log.Fatalf("연결 실패: %v", err)
	}
	defer conn.Close()

	// Publisher 생성
	pub, err := rabbitmq.NewPublisher(conn, exchangeName, exchangeType)
	if err != nil {
		log.Fatalf("Publisher 생성 실패: %v", err)
	}

	ctx := context.Background()

	// 다양한 주문 이벤트 발행
	orders := []models.OrderEvent{
		{OrderID: "ORD-001", CustomerID: "CUST-100", Amount: 150000, Status: "created", CreatedAt: time.Now()},
		{OrderID: "ORD-002", CustomerID: "CUST-101", Amount: 89000, Status: "created", CreatedAt: time.Now()},
		{OrderID: "ORD-003", CustomerID: "CUST-102", Amount: 0, Status: "created", CreatedAt: time.Now()}, // 에러 유발용 (금액 0)
		{OrderID: "ORD-004", CustomerID: "CUST-103", Amount: 250000, Status: "created", CreatedAt: time.Now()},
	}

	for _, order := range orders {
		// Topic Exchange 라우팅 키 예시: order.created, order.paid, order.shipped
		routingKey := fmt.Sprintf("order.%s", order.Status)

		err := pub.Publish(ctx, routingKey, order)
		if err != nil {
			log.Printf("[❌] 메시지 발행 실패: %v", err)
		} else {
			log.Printf("[📤] 메시지 발행 완료: %s (routing: %s)", order.OrderID, routingKey)
		}

		time.Sleep(500 * time.Millisecond) // 모니터링 확인용 딜레이
	}

	log.Println("[✅] 모든 메시지 발행 완료!")
}
