package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"rabbit-mq-with-go/internal/rabbitmq"
	"rabbit-mq-with-go/internal/schema"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var registry *schema.SchemaRegistry
var templates *template.Template

// RabbitMQ 연결 관리
var (
	rabbitConn *rabbitmq.Connection
	publishers map[string]*rabbitmq.Publisher
	pubMutex   sync.RWMutex
	rabbitURL  = "amqp://guest:guest@localhost:5672/"
)

// 메시지 발행 히스토리
type PublishRecord struct {
	ID         int       `json:"id"`
	SchemaName string    `json:"schema_name"`
	Exchange   string    `json:"exchange"`
	RoutingKey string    `json:"routing_key"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

var (
	publishHistory []PublishRecord
	historyMutex   sync.RWMutex
	publishID      int
)

func main() {
	// 스키마 레지스트리 초기화
	registry = schema.NewSchemaRegistry()
	initDefaultSchemas()

	// RabbitMQ 연결 초기화
	publishers = make(map[string]*rabbitmq.Publisher)
	publishHistory = make([]PublishRecord, 0)
	initRabbitMQ()

	// 템플릿 로드
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("템플릿 로드 실패: %v", err)
	}

	// 라우터 설정
	mux := http.NewServeMux()

	// 정적 파일
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// 페이지 라우트
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/schemas", handleSchemaList)
	mux.HandleFunc("/schemas/", handleSchemaDetail)
	mux.HandleFunc("/validate", handleValidatePage)
	mux.HandleFunc("/publish", handlePublishPage)

	// API 라우트
	mux.HandleFunc("/api/schemas", handleAPISchemas)
	mux.HandleFunc("/api/schemas/", handleAPISchemaByName)
	mux.HandleFunc("/api/validate", handleAPIValidate)
	mux.HandleFunc("/api/stats", handleAPIStats)
	mux.HandleFunc("/api/publish", handleAPIPublish)
	mux.HandleFunc("/api/publish/history", handleAPIPublishHistory)
	mux.HandleFunc("/api/rabbitmq/status", handleAPIRabbitMQStatus)

	port := ":8080"
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         Schema Registry Web Dashboard                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n  서버 시작: http://localhost%s\n\n", port)
	fmt.Println("  페이지:")
	fmt.Println("    • 대시보드:     http://localhost:8080/")
	fmt.Println("    • 스키마 목록:  http://localhost:8080/schemas")
	fmt.Println("    • 검증 도구:    http://localhost:8080/validate")
	fmt.Println("\n  API:")
	fmt.Println("    • GET  /api/schemas        - 스키마 목록")
	fmt.Println("    • GET  /api/schemas/:name  - 스키마 조회")
	fmt.Println("    • POST /api/schemas        - 스키마 등록")
	fmt.Println("    • POST /api/validate       - 데이터 검증")
	fmt.Println("    • POST /api/publish        - RabbitMQ에 메시지 게시")
	fmt.Println("    • GET  /api/publish/history - 게시 히스토리")
	fmt.Println("    • GET  /api/rabbitmq/status - RabbitMQ 연결 상태")
	fmt.Println("    • GET  /api/stats          - 통계")
	fmt.Println("\n  Ctrl+C로 종료")

	log.Fatal(http.ListenAndServe(port, mux))
}

func initDefaultSchemas() {
	registry.Register("OrderEvent", schema.SchemaTypeJSON, schema.OrderEventSchema, &schema.RegisterOptions{
		Description: "주문 생성, 결제, 배송 등 주문 관련 이벤트",
	})
	registry.Register("NotificationEvent", schema.SchemaTypeJSON, schema.NotificationEventSchema, &schema.RegisterOptions{
		Description: "이메일, SMS, 푸시 알림 이벤트",
	})
	registry.Register("UserEvent", schema.SchemaTypeJSON, schema.UserEventSchema, &schema.RegisterOptions{
		Description: "사용자 로그인, 로그아웃, 가입 이벤트",
	})
	registry.Register("PaymentEvent", schema.SchemaTypeJSON, schema.PaymentEventSchema, &schema.RegisterOptions{
		Description: "결제 처리 및 환불 이벤트",
	})
}

// ============================================
// 페이지 핸들러
// ============================================

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	stats := registry.GetStats()
	schemas := registry.List()

	data := map[string]interface{}{
		"Title":         "Dashboard",
		"Stats":         stats,
		"Schemas":       schemas,
		"TotalSchemas":  stats["total_schemas"],
		"TotalVersions": stats["total_versions"],
	}

	renderTemplate(w, "dashboard.html", data)
}

func handleSchemaList(w http.ResponseWriter, r *http.Request) {
	schemas := registry.List()
	data := map[string]interface{}{
		"Title":   "Schemas",
		"Schemas": schemas,
	}
	renderTemplate(w, "schemas.html", data)
}

func handleSchemaDetail(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/schemas/")
	if name == "" {
		http.Redirect(w, r, "/schemas", http.StatusFound)
		return
	}

	s, err := registry.Get(name)
	if err != nil {
		http.Error(w, "스키마를 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	versions, _ := registry.GetVersions(name)

	// JSON 예쁘게 포맷팅
	var prettySchema map[string]interface{}
	json.Unmarshal(s.Schema, &prettySchema)
	formattedSchema, _ := json.MarshalIndent(prettySchema, "", "  ")

	data := map[string]interface{}{
		"Title":           s.Name,
		"Schema":          s,
		"Versions":        versions,
		"FormattedSchema": string(formattedSchema),
	}
	renderTemplate(w, "schema_detail.html", data)
}

func handleValidatePage(w http.ResponseWriter, r *http.Request) {
	schemas := registry.List()
	data := map[string]interface{}{
		"Title":   "Validate",
		"Schemas": schemas,
	}
	renderTemplate(w, "validate.html", data)
}

// ============================================
// API 핸들러
// ============================================

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func handleAPISchemas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		schemas := registry.List()
		json.NewEncoder(w).Encode(APIResponse{Success: true, Data: schemas})

	case http.MethodPost:
		var req struct {
			Name        string          `json:"name"`
			Type        string          `json:"type"`
			Schema      json.RawMessage `json:"schema"`
			Description string          `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
			return
		}

		schemaType := schema.SchemaTypeJSON
		if req.Type == "protobuf" {
			schemaType = schema.SchemaTypeProtobuf
		} else if req.Type == "avro" {
			schemaType = schema.SchemaTypeAvro
		}

		err := registry.Register(req.Name, schemaType, req.Schema, &schema.RegisterOptions{
			Description: req.Description,
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(APIResponse{Success: true, Data: "스키마가 등록되었습니다"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleAPISchemaByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := strings.TrimPrefix(r.URL.Path, "/api/schemas/")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "스키마 이름이 필요합니다"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		s, err := registry.Get(name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(APIResponse{Success: true, Data: s})

	case http.MethodDelete:
		err := registry.Delete(name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(APIResponse{Success: true, Data: "스키마가 삭제되었습니다"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleAPIValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SchemaName string          `json:"schema_name"`
		Data       json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	result := registry.Validate(req.SchemaName, req.Data)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: result})
}

func handleAPIStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := registry.GetStats()
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: stats})
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ============================================
// RabbitMQ 연결 및 발행 기능
// ============================================

func initRabbitMQ() {
	var err error
	rabbitConn, err = rabbitmq.NewConnection(rabbitURL)
	if err != nil {
		log.Printf("⚠️  RabbitMQ 연결 실패: %v", err)
		log.Println("   docker-compose up -d 로 RabbitMQ를 시작해주세요")
		return
	}
	log.Println("✅ RabbitMQ 연결 성공")
}

func getOrCreatePublisher(exchange string) (*rabbitmq.Publisher, error) {
	pubMutex.Lock()
	defer pubMutex.Unlock()

	if pub, ok := publishers[exchange]; ok {
		return pub, nil
	}

	if rabbitConn == nil {
		return nil, fmt.Errorf("RabbitMQ에 연결되어 있지 않습니다")
	}

	pub, err := rabbitmq.NewPublisher(rabbitConn, exchange, "topic")
	if err != nil {
		return nil, err
	}

	publishers[exchange] = pub
	return pub, nil
}

// Publish 페이지 핸들러
func handlePublishPage(w http.ResponseWriter, r *http.Request) {
	schemas := registry.List()

	historyMutex.RLock()
	history := make([]PublishRecord, len(publishHistory))
	copy(history, publishHistory)
	historyMutex.RUnlock()

	connected := rabbitConn != nil

	data := map[string]interface{}{
		"Title":     "Publish",
		"Schemas":   schemas,
		"History":   history,
		"Connected": connected,
	}
	renderTemplate(w, "publish.html", data)
}

// Publish API 핸들러
func handleAPIPublish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SchemaName string          `json:"schema_name"`
		Exchange   string          `json:"exchange"`
		RoutingKey string          `json:"routing_key"`
		Data       json.RawMessage `json:"data"`
		Validate   bool            `json:"validate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	// 기본값 설정
	if req.Exchange == "" {
		req.Exchange = "schema.events"
	}
	if req.RoutingKey == "" {
		req.RoutingKey = strings.ToLower(req.SchemaName)
	}

	// 검증 옵션이 켜져 있으면 스키마 검증 수행
	if req.Validate {
		result := registry.Validate(req.SchemaName, req.Data)
		if !result.Valid {
			record := addPublishRecord(req.SchemaName, req.Exchange, req.RoutingKey, false, "검증 실패: "+result.Message)
			json.NewEncoder(w).Encode(APIResponse{
				Success: false,
				Error:   "스키마 검증 실패",
				Data: map[string]interface{}{
					"validation": result,
					"record":     record,
				},
			})
			return
		}
	}

	// RabbitMQ에 발행
	pub, err := getOrCreatePublisher(req.Exchange)
	if err != nil {
		record := addPublishRecord(req.SchemaName, req.Exchange, req.RoutingKey, false, err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
			Data:    record,
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 메시지에 스키마 정보 헤더 추가
	var msgData map[string]interface{}
	json.Unmarshal(req.Data, &msgData)

	err = pub.PublishWithHeaders(ctx, req.RoutingKey, msgData, map[string]interface{}{
		"schema_name":    req.SchemaName,
		"schema_version": getSchemaVersion(req.SchemaName),
		"published_at":   time.Now().Format(time.RFC3339),
	})

	if err != nil {
		record := addPublishRecord(req.SchemaName, req.Exchange, req.RoutingKey, false, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
			Data:    record,
		})
		return
	}

	record := addPublishRecord(req.SchemaName, req.Exchange, req.RoutingKey, true, "")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "메시지가 성공적으로 게시되었습니다",
			"record":  record,
		},
	})
}

func addPublishRecord(schemaName, exchange, routingKey string, success bool, errMsg string) PublishRecord {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	publishID++
	record := PublishRecord{
		ID:         publishID,
		SchemaName: schemaName,
		Exchange:   exchange,
		RoutingKey: routingKey,
		Success:    success,
		Error:      errMsg,
		Timestamp:  time.Now(),
	}

	// 최신 것을 앞에 추가
	publishHistory = append([]PublishRecord{record}, publishHistory...)

	// 최대 100개만 유지
	if len(publishHistory) > 100 {
		publishHistory = publishHistory[:100]
	}

	return record
}

func getSchemaVersion(name string) int {
	s, err := registry.Get(name)
	if err != nil {
		return 0
	}
	return s.Version
}

// Publish History API
func handleAPIPublishHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	historyMutex.RLock()
	defer historyMutex.RUnlock()

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    publishHistory,
	})
}

// RabbitMQ Status API
func handleAPIRabbitMQStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"connected":  rabbitConn != nil,
		"url":        rabbitURL,
		"publishers": len(publishers),
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    status,
	})
}
