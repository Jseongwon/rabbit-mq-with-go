package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"rabbit-mq-with-go/internal/monitor"
	"rabbit-mq-with-go/internal/schema"
)

const (
	rabbitMQURL = "http://localhost:15672"
	username    = "guest"
	password    = "guest"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          RabbitMQ Monitor & Schema Registry Demo           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// 1. 스키마 레지스트리 데모
	demoSchemaRegistry()

	fmt.Println("\n" + strings.Repeat("─", 60))

	// 2. RabbitMQ 모니터링 데모
	demoRabbitMQMonitor()
}

func demoSchemaRegistry() {
	fmt.Println("\n📋 [스키마 레지스트리 데모]")
	fmt.Println(strings.Repeat("-", 40))

	registry := schema.NewSchemaRegistry()

	// 스키마 등록
	registry.Register("OrderEvent", schema.SchemaTypeJSON, schema.OrderEventSchema, nil)
	registry.Register("NotificationEvent", schema.SchemaTypeJSON, schema.NotificationEventSchema, nil)

	// 등록된 스키마 목록 출력
	fmt.Println("\n등록된 스키마 목록:")
	for _, s := range registry.List() {
		fmt.Printf("  • %s (v%d, type: %s)\n", s.Name, s.Version, s.Type)
	}

	// 스키마 상세 조회
	orderSchema, err := registry.Get("OrderEvent")
	if err != nil {
		log.Printf("스키마 조회 실패: %v", err)
		return
	}

	fmt.Println("\nOrderEvent 스키마:")
	var prettyJSON map[string]interface{}
	json.Unmarshal(orderSchema.Schema, &prettyJSON)
	formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
	fmt.Printf("  %s\n", formatted)
}

func demoRabbitMQMonitor() {
	fmt.Println("\n🔍 [RabbitMQ 모니터링 데모]")
	fmt.Println(strings.Repeat("-", 40))

	mon := monitor.NewRabbitMQMonitor(rabbitMQURL, username, password)

	// Overview 조회
	fmt.Println("\n1. 전체 개요 (Overview)")
	overview, err := mon.GetOverview()
	if err != nil {
		log.Printf("Overview 조회 실패 (RabbitMQ가 실행 중인지 확인하세요): %v", err)
		fmt.Println("   ⚠️  RabbitMQ Management UI에 접속할 수 없습니다.")
		fmt.Println("   다음 명령으로 RabbitMQ를 시작하세요:")
		fmt.Println("   $ docker-compose up -d")
		return
	}

	fmt.Printf("   RabbitMQ Version: %s\n", overview.RabbitMQVersion)
	fmt.Printf("   Cluster Name: %s\n", overview.ClusterName)
	fmt.Printf("   총 메시지: %d (대기: %d, 처리중: %d)\n",
		overview.QueueTotals.Messages,
		overview.QueueTotals.MessagesReady,
		overview.QueueTotals.MessagesUnacked)
	fmt.Printf("   Connections: %d, Channels: %d\n",
		overview.ObjectTotals.Connections,
		overview.ObjectTotals.Channels)
	fmt.Printf("   Exchanges: %d, Queues: %d, Consumers: %d\n",
		overview.ObjectTotals.Exchanges,
		overview.ObjectTotals.Queues,
		overview.ObjectTotals.Consumers)

	// Exchange 목록
	fmt.Println("\n2. Exchange 목록")
	exchanges, err := mon.ListExchanges()
	if err != nil {
		log.Printf("Exchange 조회 실패: %v", err)
	} else {
		for _, ex := range exchanges {
			if ex.Name == "" {
				continue // 기본 exchange 제외
			}
			if strings.HasPrefix(ex.Name, "amq.") {
				continue // 시스템 exchange 제외
			}
			fmt.Printf("   • %s (type: %s, durable: %v)\n", ex.Name, ex.Type, ex.Durable)
		}
	}

	// Queue 목록
	fmt.Println("\n3. Queue 목록")
	queues, err := mon.ListQueues()
	if err != nil {
		log.Printf("Queue 조회 실패: %v", err)
	} else {
		if len(queues) == 0 {
			fmt.Println("   (등록된 큐가 없습니다)")
		}
		for _, q := range queues {
			fmt.Printf("   • %s\n", q.Name)
			fmt.Printf("     Messages: %d (ready: %d, unacked: %d)\n",
				q.Messages, q.MessagesReady, q.MessagesUnacked)
			fmt.Printf("     Consumers: %d, State: %s\n", q.Consumers, q.State)
		}
	}

	// Binding 목록
	fmt.Println("\n4. Binding 목록")
	bindings, err := mon.ListBindings()
	if err != nil {
		log.Printf("Binding 조회 실패: %v", err)
	} else {
		for _, b := range bindings {
			if b.Source == "" {
				continue // 기본 바인딩 제외
			}
			fmt.Printf("   • %s → %s (routing: %s)\n",
				b.Source, b.Destination, b.RoutingKey)
		}
	}

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("💡 Management UI: http://localhost:15672 (guest/guest)")
	fmt.Printf("   현재 시간: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
