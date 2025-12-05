package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rabbit-mq-with-go/internal/models"
	"rabbit-mq-with-go/internal/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchangeName = "orders.exchange"
	queueName    = "orders.queue"
	routingKey   = "order.*" // 모든 order 이벤트 구독 (topic pattern)

	// DLQ 설정
	dlqExchange = "orders.dlx"
	dlqQueue    = "orders.dlq"
)

func main() {
	// RabbitMQ 연결
	conn, err := rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		log.Fatalf("연결 실패: %v", err)
	}
	defer conn.Close()

	// Exchange 선언 (Publisher와 동일해야 함)
	err = conn.Channel().ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Exchange 선언 실패: %v", err)
	}

	// Consumer 생성 (DLQ 설정 포함)
	consumer, err := rabbitmq.NewConsumer(conn, rabbitmq.ConsumerConfig{
		QueueName:     queueName,
		Exchange:      exchangeName,
		RoutingKey:    routingKey,
		DLQExchange:   dlqExchange,
		DLQQueue:      dlqQueue,
		PrefetchCount: 10,
	})
	if err != nil {
		log.Fatalf("Consumer 생성 실패: %v", err)
	}

	// Graceful Shutdown 설정
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n[🛑] 종료 신호 수신, Consumer 종료 중...")
		conn.Close()
		os.Exit(0)
	}()

	// 메시지 소비 시작
	log.Println("[🚀] Consumer 시작!")
	err = consumer.Consume(handleOrder)
	if err != nil {
		log.Fatalf("Consume 실패: %v", err)
	}
}

// handleOrder 주문 메시지 처리 핸들러
func handleOrder(delivery amqp.Delivery) error {
	var order models.OrderEvent
	if err := json.Unmarshal(delivery.Body, &order); err != nil {
		return err
	}

	log.Printf("[📦] 주문 처리 중: %s (고객: %s, 금액: %.0f원)",
		order.OrderID, order.CustomerID, order.Amount)

	// 비즈니스 로직 예시: 금액이 0이면 에러
	if order.Amount <= 0 {
		return errors.New("주문 금액이 유효하지 않습니다")
	}

	// 정상 처리
	log.Printf("[💰] 주문 %s 처리 완료!", order.OrderID)
	return nil
}
