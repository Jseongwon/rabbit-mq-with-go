package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rabbit-mq-with-go/internal/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchangeName = "schema.events" // 웹 대시보드와 동일한 Exchange
	queueName    = "schema.consumer.queue"
	routingKey   = "#" // 모든 메시지 수신 (topic wildcard)
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Schema Events Consumer                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Exchange:    %s\n", exchangeName)
	fmt.Printf("  Queue:       %s\n", queueName)
	fmt.Printf("  Routing Key: %s (모든 메시지)\n", routingKey)
	fmt.Println()

	// RabbitMQ 연결
	conn, err := rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		log.Fatalf("❌ 연결 실패: %v", err)
	}
	defer conn.Close()

	log.Println("✅ RabbitMQ 연결 성공")

	// Exchange 선언 (웹 대시보드와 동일)
	err = conn.Channel().ExchangeDeclare(
		exchangeName,
		"topic", // topic exchange
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Exchange 선언 실패: %v", err)
	}

	// Consumer 생성
	consumer, err := rabbitmq.NewConsumer(conn, rabbitmq.ConsumerConfig{
		QueueName:     queueName,
		Exchange:      exchangeName,
		RoutingKey:    routingKey,
		PrefetchCount: 10,
	})
	if err != nil {
		log.Fatalf("❌ Consumer 생성 실패: %v", err)
	}

	// Graceful Shutdown 설정
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n")
		log.Println("🛑 종료 신호 수신, Consumer 종료 중...")
		conn.Close()
		os.Exit(0)
	}()

	// 메시지 소비 시작
	log.Println("🚀 메시지 대기 중... (Ctrl+C로 종료)")
	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────────")

	err = consumer.Consume(handleSchemaMessage)
	if err != nil {
		log.Fatalf("❌ Consume 실패: %v", err)
	}
}

// handleSchemaMessage 스키마 기반 메시지 처리 핸들러
func handleSchemaMessage(delivery amqp.Delivery) error {
	fmt.Println()
	fmt.Println("╭────────────────────────────────────────────────────────────╮")
	fmt.Println("│  📩 새 메시지 수신                                          │")
	fmt.Println("╰────────────────────────────────────────────────────────────╯")

	// 헤더 정보 출력
	fmt.Println("  📋 헤더:")
	if schemaName, ok := delivery.Headers["schema_name"]; ok {
		fmt.Printf("     • schema_name: %v\n", schemaName)
	}
	if schemaVersion, ok := delivery.Headers["schema_version"]; ok {
		fmt.Printf("     • schema_version: %v\n", schemaVersion)
	}
	if publishedAt, ok := delivery.Headers["published_at"]; ok {
		fmt.Printf("     • published_at: %v\n", publishedAt)
	}

	fmt.Printf("  🔑 Routing Key: %s\n", delivery.RoutingKey)
	fmt.Printf("  📦 Exchange: %s\n", delivery.Exchange)
	fmt.Printf("  📄 Content-Type: %s\n", delivery.ContentType)

	// 메시지 본문 출력 (JSON 포맷팅)
	fmt.Println("  📝 메시지 본문:")
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(delivery.Body, &prettyJSON); err != nil {
		fmt.Printf("     (JSON 파싱 실패) %s\n", string(delivery.Body))
	} else {
		formatted, _ := json.MarshalIndent(prettyJSON, "     ", "  ")
		fmt.Printf("     %s\n", formatted)
	}

	fmt.Println("────────────────────────────────────────────────────────────")

	// 정상 처리
	log.Println("✅ 메시지 처리 완료!")
	return nil
}
