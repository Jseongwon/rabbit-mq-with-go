package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rabbit-mq-with-go/internal/models"
	"rabbit-mq-with-go/internal/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitURL = "amqp://guest:guest@localhost:5672/"
	dlqQueue  = "orders.dlq"
)

func main() {
	log.Println("[💀 DLQ Consumer 시작]")
	log.Println("실패한 메시지들을 처리합니다...")

	// RabbitMQ 연결
	conn, err := rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		log.Fatalf("연결 실패: %v", err)
	}
	defer conn.Close()

	// DLQ Consumer 생성 (DLQ는 이미 생성되어 있으므로 간단히 구성)
	consumer, err := rabbitmq.NewConsumer(conn, rabbitmq.ConsumerConfig{
		QueueName:     dlqQueue,
		PrefetchCount: 5,
	})
	if err != nil {
		log.Fatalf("DLQ Consumer 생성 실패: %v", err)
	}

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n[🛑] DLQ Consumer 종료 중...")
		conn.Close()
		os.Exit(0)
	}()

	// DLQ 메시지 처리
	err = consumer.Consume(handleDeadLetter)
	if err != nil {
		log.Fatalf("DLQ Consume 실패: %v", err)
	}
}

func handleDeadLetter(delivery amqp.Delivery) error {
	var order models.OrderEvent
	if err := json.Unmarshal(delivery.Body, &order); err != nil {
		log.Printf("[❌] 메시지 파싱 실패: %v", err)
		return nil // 파싱 실패는 재시도해도 의미 없으므로 ACK
	}

	log.Println("╔════════════════════════════════════════╗")
	log.Println("║         💀 Dead Letter 수신            ║")
	log.Println("╚════════════════════════════════════════╝")
	log.Printf("  Order ID: %s", order.OrderID)
	log.Printf("  Customer: %s", order.CustomerID)
	log.Printf("  Amount: %.0f", order.Amount)
	log.Printf("  Status: %s", order.Status)

	// x-death 헤더에서 실패 정보 추출
	if xDeath, ok := delivery.Headers["x-death"]; ok {
		deaths := xDeath.([]interface{})
		for _, death := range deaths {
			deathInfo := death.(amqp.Table)
			log.Printf("  실패 횟수: %d", deathInfo["count"])
			log.Printf("  원인: %s", deathInfo["reason"])
			log.Printf("  원본 큐: %s", deathInfo["queue"])
		}
	}

	log.Println("─────────────────────────────────────────")

	// 여기서 실패한 메시지에 대한 처리 수행:
	// 1. DB에 실패 기록 저장
	// 2. 알림 발송 (Slack, 이메일 등)
	// 3. 수동 재처리를 위한 대기열에 추가
	// 4. 메트릭 수집

	return nil
}
