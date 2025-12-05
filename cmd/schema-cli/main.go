package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"rabbit-mq-with-go/internal/schema"
)

var registry *schema.SchemaRegistry

func main() {
	registry = schema.NewSchemaRegistry()

	// 기본 스키마 등록
	initDefaultSchemas()

	fmt.Println()
	printBanner()
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n[schema-registry] > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]

		switch command {
		case "help", "h", "?":
			printHelp()
		case "list", "ls":
			listSchemas()
		case "get", "show":
			if len(args) < 1 {
				fmt.Println("  사용법: get <스키마명>")
				continue
			}
			getSchema(args[0])
		case "versions":
			if len(args) < 1 {
				fmt.Println("  사용법: versions <스키마명>")
				continue
			}
			getVersions(args[0])
		case "validate":
			if len(args) < 1 {
				fmt.Println("  사용법: validate <스키마명>")
				continue
			}
			validateData(args[0], reader)
		case "register":
			registerNewSchema(reader)
		case "delete", "rm":
			if len(args) < 1 {
				fmt.Println("  사용법: delete <스키마명>")
				continue
			}
			deleteSchema(args[0])
		case "stats":
			showStats()
		case "demo":
			runDemo()
		case "clear", "cls":
			clearScreen()
		case "exit", "quit", "q":
			fmt.Println("\n  Goodbye! 👋")
			return
		default:
			fmt.Printf("  알 수 없는 명령: %s (help로 도움말 보기)\n", command)
		}
	}
}

func printBanner() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                   ║")
	fmt.Println("║    ███████╗ ██████╗██╗  ██╗███████╗███╗   ███╗ █████╗            ║")
	fmt.Println("║    ██╔════╝██╔════╝██║  ██║██╔════╝████╗ ████║██╔══██╗           ║")
	fmt.Println("║    ███████╗██║     ███████║█████╗  ██╔████╔██║███████║           ║")
	fmt.Println("║    ╚════██║██║     ██╔══██║██╔══╝  ██║╚██╔╝██║██╔══██║           ║")
	fmt.Println("║    ███████║╚██████╗██║  ██║███████╗██║ ╚═╝ ██║██║  ██║           ║")
	fmt.Println("║    ╚══════╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝           ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("║         ██████╗ ███████╗ ██████╗ ██╗███████╗████████╗██████╗ ██╗ ║")
	fmt.Println("║         ██╔══██╗██╔════╝██╔════╝ ██║██╔════╝╚══██╔══╝██╔══██╗╚██╗║")
	fmt.Println("║         ██████╔╝█████╗  ██║  ███╗██║███████╗   ██║   ██████╔╝ ██║║")
	fmt.Println("║         ██╔══██╗██╔══╝  ██║   ██║██║╚════██║   ██║   ██╔══██╗ ██║║")
	fmt.Println("║         ██║  ██║███████╗╚██████╔╝██║███████║   ██║   ██║  ██║██╔╝║")
	fmt.Println("║         ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝ ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("║                  RabbitMQ Schema Registry CLI                     ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  'help' 명령으로 사용 가능한 명령어를 확인하세요.")
}

func printHelp() {
	fmt.Println()
	fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
	fmt.Println("  │                      사용 가능한 명령어                       │")
	fmt.Println("  ├──────────────────────────────────────────────────────────────┤")
	fmt.Println("  │  list, ls           등록된 모든 스키마 목록                   │")
	fmt.Println("  │  get <name>         스키마 상세 정보 조회                     │")
	fmt.Println("  │  versions <name>    스키마 버전 히스토리                      │")
	fmt.Println("  │  validate <name>    데이터 검증 (대화형)                      │")
	fmt.Println("  │  register           새 스키마 등록 (대화형)                   │")
	fmt.Println("  │  delete <name>      스키마 삭제                               │")
	fmt.Println("  │  stats              레지스트리 통계                           │")
	fmt.Println("  │  demo               검증 데모 실행                            │")
	fmt.Println("  │  clear              화면 지우기                               │")
	fmt.Println("  │  exit, quit         종료                                      │")
	fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")
}

func initDefaultSchemas() {
	registry.Register("OrderEvent", schema.SchemaTypeJSON, schema.OrderEventSchema, &schema.RegisterOptions{
		Description: "주문 이벤트 스키마",
	})
	registry.Register("NotificationEvent", schema.SchemaTypeJSON, schema.NotificationEventSchema, &schema.RegisterOptions{
		Description: "알림 이벤트 스키마",
	})
	registry.Register("UserEvent", schema.SchemaTypeJSON, schema.UserEventSchema, &schema.RegisterOptions{
		Description: "사용자 이벤트 스키마",
	})
	registry.Register("PaymentEvent", schema.SchemaTypeJSON, schema.PaymentEventSchema, &schema.RegisterOptions{
		Description: "결제 이벤트 스키마",
	})
}

func listSchemas() {
	schemas := registry.List()
	if len(schemas) == 0 {
		fmt.Println("\n  등록된 스키마가 없습니다.")
		return
	}

	fmt.Println()
	fmt.Println("  ╭────┬─────────────────────┬─────────┬──────────┬────────────────────────────╮")
	fmt.Println("  │ ID │ Name                │ Version │ Type     │ Description                │")
	fmt.Println("  ├────┼─────────────────────┼─────────┼──────────┼────────────────────────────┤")
	for _, s := range schemas {
		desc := s.Description
		if len(desc) > 24 {
			desc = desc[:21] + "..."
		}
		fmt.Printf("  │ %2d │ %-19s │   v%-4d │ %-8s │ %-26s │\n",
			s.ID, truncate(s.Name, 19), s.Version, s.Type, desc)
	}
	fmt.Println("  ╰────┴─────────────────────┴─────────┴──────────┴────────────────────────────╯")
	fmt.Printf("\n  총 %d개의 스키마가 등록되어 있습니다.\n", len(schemas))
}

func getSchema(name string) {
	s, err := registry.Get(name)
	if err != nil {
		fmt.Printf("\n  ❌ %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
	fmt.Printf("  │  스키마: %-52s │\n", s.Name)
	fmt.Println("  ├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("  │  ID:          %-47d │\n", s.ID)
	fmt.Printf("  │  Version:     v%-46d │\n", s.Version)
	fmt.Printf("  │  Type:        %-47s │\n", s.Type)
	fmt.Printf("  │  Compat:      %-47s │\n", s.Compatibility)
	fmt.Printf("  │  Description: %-47s │\n", truncate(s.Description, 47))
	fmt.Printf("  │  Created:     %-47s │\n", s.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")

	fmt.Println("\n  📄 Schema Definition:")
	fmt.Println("  " + strings.Repeat("─", 60))

	var prettyJSON map[string]interface{}
	json.Unmarshal(s.Schema, &prettyJSON)
	formatted, _ := json.MarshalIndent(prettyJSON, "  ", "    ")
	fmt.Printf("  %s\n", formatted)
}

func getVersions(name string) {
	versions, err := registry.GetVersions(name)
	if err != nil {
		fmt.Printf("\n  ❌ %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("  📜 '%s' 버전 히스토리:\n", name)
	fmt.Println("  " + strings.Repeat("─", 50))

	for _, v := range versions {
		fmt.Printf("  • v%d - %s\n", v.Version, v.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}

func validateData(schemaName string, reader *bufio.Reader) {
	_, err := registry.Get(schemaName)
	if err != nil {
		fmt.Printf("\n  ❌ %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("  검증할 JSON 데이터를 입력하세요 (빈 줄로 종료):")
	fmt.Println("  " + strings.Repeat("─", 50))

	var jsonBuilder strings.Builder
	for {
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) == "" {
			break
		}
		jsonBuilder.WriteString(line)
	}

	jsonData := strings.TrimSpace(jsonBuilder.String())
	if jsonData == "" {
		fmt.Println("  ⚠️  입력된 데이터가 없습니다.")
		return
	}

	result := registry.Validate(schemaName, []byte(jsonData))

	fmt.Println()
	if result.Valid {
		fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
		fmt.Println("  │  ✅ 검증 성공!                                               │")
		fmt.Println("  │     데이터가 스키마와 일치합니다.                             │")
		fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")
	} else {
		fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
		fmt.Println("  │  ❌ 검증 실패!                                               │")
		fmt.Println("  ├──────────────────────────────────────────────────────────────┤")
		for _, e := range result.Errors {
			fmt.Printf("  │  • %-57s │\n", truncate(e, 57))
		}
		fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")
	}
}

func registerNewSchema(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("  새 스키마 등록")
	fmt.Println("  " + strings.Repeat("─", 50))

	fmt.Print("  스키마 이름: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		fmt.Println("  ⚠️  이름을 입력해주세요.")
		return
	}

	fmt.Print("  설명: ")
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)

	fmt.Println("  JSON 스키마를 입력하세요 (빈 줄로 종료):")
	var jsonBuilder strings.Builder
	for {
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) == "" {
			break
		}
		jsonBuilder.WriteString(line)
	}

	schemaJSON := strings.TrimSpace(jsonBuilder.String())
	if schemaJSON == "" {
		fmt.Println("  ⚠️  스키마를 입력해주세요.")
		return
	}

	err := registry.Register(name, schema.SchemaTypeJSON, json.RawMessage(schemaJSON), &schema.RegisterOptions{
		Description: desc,
	})

	if err != nil {
		fmt.Printf("\n  ❌ 등록 실패: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
	fmt.Printf("  │  ✅ 스키마 '%s' 등록 완료!                          \n", name)
	fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")
}

func deleteSchema(name string) {
	err := registry.Delete(name)
	if err != nil {
		fmt.Printf("\n  ❌ %v\n", err)
		return
	}
	fmt.Printf("\n  ✅ 스키마 '%s'가 삭제되었습니다.\n", name)
}

func showStats() {
	stats := registry.GetStats()

	fmt.Println()
	fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
	fmt.Println("  │                    📊 레지스트리 통계                        │")
	fmt.Println("  ├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("  │  총 스키마 수:     %-41v │\n", stats["total_schemas"])
	fmt.Printf("  │  총 버전 수:       %-41v │\n", stats["total_versions"])
	fmt.Println("  ├──────────────────────────────────────────────────────────────┤")
	fmt.Println("  │  타입별 스키마:                                              │")
	if byType, ok := stats["by_type"].(map[schema.SchemaType]int); ok {
		for t, count := range byType {
			fmt.Printf("  │    • %-10s: %-45d │\n", t, count)
		}
	}
	fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")
}

func runDemo() {
	fmt.Println()
	fmt.Println("  ╭──────────────────────────────────────────────────────────────╮")
	fmt.Println("  │                    🎯 스키마 검증 데모                        │")
	fmt.Println("  ╰──────────────────────────────────────────────────────────────╯")

	testCases := []struct {
		name    string
		schema  string
		data    string
		desc    string
	}{
		{
			name:   "OrderEvent",
			schema: "OrderEvent",
			data:   `{"order_id": "ORD-001", "customer_id": "CUST-100", "amount": 150000, "status": "created", "created_at": "2024-01-15T10:30:00Z"}`,
			desc:   "정상적인 주문 데이터",
		},
		{
			name:   "OrderEvent",
			schema: "OrderEvent",
			data:   `{"order_id": "ORD-002", "amount": -5000, "status": "invalid_status"}`,
			desc:   "필수 필드 누락 + 잘못된 값",
		},
		{
			name:   "PaymentEvent",
			schema: "PaymentEvent",
			data:   `{"payment_id": "PAY-001", "order_id": "ORD-001", "amount": 50000, "currency": "KRW", "status": "completed"}`,
			desc:   "정상적인 결제 데이터",
		},
		{
			name:   "UserEvent",
			schema: "UserEvent",
			data:   `{"user_id": "USR-001", "action": "login", "timestamp": "2024-01-15T09:00:00Z"}`,
			desc:   "정상적인 사용자 이벤트",
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n  ─── 테스트 %d: %s ───\n", i+1, tc.desc)
		fmt.Printf("  스키마: %s\n", tc.schema)
		fmt.Printf("  데이터: %s\n", truncate(tc.data, 60))

		result := registry.Validate(tc.schema, []byte(tc.data))

		if result.Valid {
			fmt.Println("  결과: ✅ 성공")
		} else {
			fmt.Println("  결과: ❌ 실패")
			for _, e := range result.Errors {
				fmt.Printf("         • %s\n", e)
			}
		}
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("  ─── 데모 완료 ───")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// 숫자를 문자열로 변환하는 헬퍼
func itoa(i int) string {
	return strconv.Itoa(i)
}
