package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           RabbitMQ Pub/Sub Learning Project                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("사용 가능한 명령:")
	fmt.Println()
	fmt.Println("  1. RabbitMQ 시작:")
	fmt.Println("     docker-compose up -d")
	fmt.Println()
	fmt.Println("  2. 모니터링 실행:")
	fmt.Println("     go run cmd/monitor/main.go")
	fmt.Println()
	fmt.Println("  3. Consumer 시작 (터미널 1):")
	fmt.Println("     go run cmd/consumer/main.go")
	fmt.Println()
	fmt.Println("  4. DLQ Consumer 시작 (터미널 2):")
	fmt.Println("     go run cmd/dlq-consumer/main.go")
	fmt.Println()
	fmt.Println("  5. Publisher 실행 (터미널 3):")
	fmt.Println("     go run cmd/publisher/main.go")
	fmt.Println()
	fmt.Println("  6. 스키마 레지스트리 CLI:")
	fmt.Println("     go run cmd/schema-cli/main.go")
	fmt.Println()
	fmt.Println("  7. 스키마 웹 대시보드:")
	fmt.Println("     go run cmd/schema-web/main.go")
	fmt.Println("     http://localhost:8080")
	fmt.Println()
	fmt.Println("  8. 스키마 Consumer (웹 대시보드 메시지 수신):")
	fmt.Println("     go run cmd/schema-consumer/main.go")
	fmt.Println()
	fmt.Println("  9. RabbitMQ Management UI:")
	fmt.Println("     http://localhost:15672 (guest/guest)")
	fmt.Println()

	if len(os.Args) > 1 && os.Args[1] == "--demo" {
		runQuickDemo()
	}
}

func runQuickDemo() {
	fmt.Println("데모를 시작하려면 위 명령들을 순서대로 실행하세요.")
}
