package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// SchemaType 지원하는 스키마 타입
type SchemaType string

const (
	SchemaTypeJSON     SchemaType = "json"
	SchemaTypeProtobuf SchemaType = "protobuf"
	SchemaTypeAvro     SchemaType = "avro"
)

// CompatibilityMode 스키마 호환성 모드
type CompatibilityMode string

const (
	CompatibilityNone     CompatibilityMode = "NONE"
	CompatibilityBackward CompatibilityMode = "BACKWARD"
	CompatibilityForward  CompatibilityMode = "FORWARD"
	CompatibilityFull     CompatibilityMode = "FULL"
)

// SchemaInfo 스키마 정보
type SchemaInfo struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	Version       int             `json:"version"`
	Type          SchemaType      `json:"type"`
	Schema        json.RawMessage `json:"schema"`
	Compatibility CompatibilityMode `json:"compatibility"`
	Description   string          `json:"description,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// SchemaVersion 스키마 버전 히스토리
type SchemaVersion struct {
	Version   int             `json:"version"`
	Schema    json.RawMessage `json:"schema"`
	CreatedAt time.Time       `json:"created_at"`
}

// ValidationResult 검증 결과
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message"`
}

// SchemaRegistry 스키마 레지스트리
type SchemaRegistry struct {
	mu            sync.RWMutex
	schemas       map[string]SchemaInfo
	versions      map[string][]SchemaVersion // 스키마 이름별 버전 히스토리
	nextID        int
	defaultCompat CompatibilityMode
}

// NewSchemaRegistry 새 스키마 레지스트리 생성
func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas:       make(map[string]SchemaInfo),
		versions:      make(map[string][]SchemaVersion),
		nextID:        1,
		defaultCompat: CompatibilityBackward,
	}
}

// RegisterOptions 스키마 등록 옵션
type RegisterOptions struct {
	Description   string
	Compatibility CompatibilityMode
}

// Register 스키마 등록 (옵션 포함)
func (r *SchemaRegistry) Register(name string, schemaType SchemaType, schema json.RawMessage, opts *RegisterOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// JSON 유효성 검증
	if schemaType == SchemaTypeJSON {
		var parsed interface{}
		if err := json.Unmarshal(schema, &parsed); err != nil {
			return fmt.Errorf("잘못된 JSON 스키마: %w", err)
		}
	}

	now := time.Now()
	version := 1
	compatibility := r.defaultCompat
	description := ""

	if opts != nil {
		if opts.Compatibility != "" {
			compatibility = opts.Compatibility
		}
		description = opts.Description
	}

	// 기존 스키마가 있으면 버전 업
	if existing, ok := r.schemas[name]; ok {
		version = existing.Version + 1

		// 버전 히스토리에 추가
		r.versions[name] = append(r.versions[name], SchemaVersion{
			Version:   existing.Version,
			Schema:    existing.Schema,
			CreatedAt: existing.CreatedAt,
		})
	}

	r.schemas[name] = SchemaInfo{
		ID:            r.nextID,
		Name:          name,
		Version:       version,
		Type:          schemaType,
		Schema:        schema,
		Compatibility: compatibility,
		Description:   description,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	r.nextID++

	return nil
}

// Get 스키마 조회
func (r *SchemaRegistry) Get(name string) (SchemaInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, ok := r.schemas[name]
	if !ok {
		return SchemaInfo{}, fmt.Errorf("스키마를 찾을 수 없습니다: %s", name)
	}
	return schema, nil
}

// GetByID ID로 스키마 조회
func (r *SchemaRegistry) GetByID(id int) (SchemaInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, schema := range r.schemas {
		if schema.ID == id {
			return schema, nil
		}
	}
	return SchemaInfo{}, fmt.Errorf("스키마 ID를 찾을 수 없습니다: %d", id)
}

// GetVersion 특정 버전의 스키마 조회
func (r *SchemaRegistry) GetVersion(name string, version int) (SchemaVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 현재 버전 확인
	current, ok := r.schemas[name]
	if !ok {
		return SchemaVersion{}, fmt.Errorf("스키마를 찾을 수 없습니다: %s", name)
	}

	if current.Version == version {
		return SchemaVersion{
			Version:   current.Version,
			Schema:    current.Schema,
			CreatedAt: current.CreatedAt,
		}, nil
	}

	// 히스토리에서 조회
	versions := r.versions[name]
	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return SchemaVersion{}, fmt.Errorf("버전을 찾을 수 없습니다: %s v%d", name, version)
}

// GetVersions 스키마의 모든 버전 조회
func (r *SchemaRegistry) GetVersions(name string) ([]SchemaVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	current, ok := r.schemas[name]
	if !ok {
		return nil, fmt.Errorf("스키마를 찾을 수 없습니다: %s", name)
	}

	versions := make([]SchemaVersion, 0)
	versions = append(versions, r.versions[name]...)
	versions = append(versions, SchemaVersion{
		Version:   current.Version,
		Schema:    current.Schema,
		CreatedAt: current.CreatedAt,
	})

	return versions, nil
}

// List 모든 스키마 목록
func (r *SchemaRegistry) List() []SchemaInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]SchemaInfo, 0, len(r.schemas))
	for _, schema := range r.schemas {
		result = append(result, schema)
	}

	// ID 순으로 정렬
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// Delete 스키마 삭제
func (r *SchemaRegistry) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.schemas[name]; !ok {
		return fmt.Errorf("스키마를 찾을 수 없습니다: %s", name)
	}

	delete(r.schemas, name)
	delete(r.versions, name)
	return nil
}

// Count 등록된 스키마 수
func (r *SchemaRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.schemas)
}

// Validate 데이터 검증
func (r *SchemaRegistry) Validate(schemaName string, data []byte) ValidationResult {
	schema, err := r.Get(schemaName)
	if err != nil {
		return ValidationResult{
			Valid:   false,
			Errors:  []string{err.Error()},
			Message: "스키마 조회 실패",
		}
	}

	switch schema.Type {
	case SchemaTypeJSON:
		return r.validateJSON(schema, data)
	case SchemaTypeProtobuf:
		return ValidationResult{
			Valid:   false,
			Errors:  []string{"Protobuf 검증은 아직 지원하지 않습니다"},
			Message: "미지원 타입",
		}
	case SchemaTypeAvro:
		return ValidationResult{
			Valid:   false,
			Errors:  []string{"Avro 검증은 아직 지원하지 않습니다"},
			Message: "미지원 타입",
		}
	default:
		return ValidationResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("알 수 없는 스키마 타입: %s", schema.Type)},
			Message: "알 수 없는 타입",
		}
	}
}

// validateJSON JSON 스키마 검증
func (r *SchemaRegistry) validateJSON(schema SchemaInfo, data []byte) ValidationResult {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return ValidationResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("JSON 파싱 실패: %v", err)},
			Message: "잘못된 JSON 형식",
		}
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema.Schema, &schemaMap); err != nil {
		return ValidationResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("스키마 파싱 실패: %v", err)},
			Message: "잘못된 스키마 형식",
		}
	}

	var validationErrors []string

	// required 필드 검증
	if required, ok := schemaMap["required"].([]interface{}); ok {
		for _, field := range required {
			fieldName := field.(string)
			if _, exists := dataMap[fieldName]; !exists {
				validationErrors = append(validationErrors, fmt.Sprintf("필수 필드 누락: %s", fieldName))
			}
		}
	}

	// properties 타입 검증
	if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for fieldName, fieldValue := range dataMap {
			if propDef, exists := properties[fieldName]; exists {
				propMap := propDef.(map[string]interface{})
				if err := r.validateField(fieldName, fieldValue, propMap); err != nil {
					validationErrors = append(validationErrors, err.Error())
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		return ValidationResult{
			Valid:   false,
			Errors:  validationErrors,
			Message: fmt.Sprintf("%d개의 검증 오류", len(validationErrors)),
		}
	}

	return ValidationResult{
		Valid:   true,
		Message: "검증 성공",
	}
}

// validateField 필드 타입 검증
func (r *SchemaRegistry) validateField(name string, value interface{}, propDef map[string]interface{}) error {
	expectedType, _ := propDef["type"].(string)

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("필드 '%s'는 string 타입이어야 합니다", name)
		}
		// enum 검증
		if enum, ok := propDef["enum"].([]interface{}); ok {
			strVal := value.(string)
			valid := false
			for _, e := range enum {
				if e.(string) == strVal {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("필드 '%s'의 값 '%s'는 허용되지 않습니다", name, strVal)
			}
		}

	case "number":
		switch v := value.(type) {
		case float64:
			// minimum 검증
			if min, ok := propDef["minimum"].(float64); ok {
				if v < min {
					return fmt.Errorf("필드 '%s'의 값은 %v 이상이어야 합니다", name, min)
				}
			}
		case int, int64:
			// int도 허용
		default:
			return fmt.Errorf("필드 '%s'는 number 타입이어야 합니다", name)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("필드 '%s'는 boolean 타입이어야 합니다", name)
		}

	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("필드 '%s'는 object 타입이어야 합니다", name)
		}

	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("필드 '%s'는 array 타입이어야 합니다", name)
		}
	}

	return nil
}

// SetCompatibility 호환성 모드 설정
func (r *SchemaRegistry) SetCompatibility(name string, mode CompatibilityMode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	schema, ok := r.schemas[name]
	if !ok {
		return errors.New("스키마를 찾을 수 없습니다")
	}

	schema.Compatibility = mode
	schema.UpdatedAt = time.Now()
	r.schemas[name] = schema
	return nil
}

// GetStats 레지스트리 통계
func (r *SchemaRegistry) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typeCount := make(map[SchemaType]int)
	totalVersions := 0

	for name, schema := range r.schemas {
		typeCount[schema.Type]++
		totalVersions += len(r.versions[name]) + 1
	}

	return map[string]interface{}{
		"total_schemas":  len(r.schemas),
		"total_versions": totalVersions,
		"by_type":        typeCount,
	}
}

// ============================================
// 미리 정의된 JSON 스키마들
// ============================================

var OrderEventSchema = json.RawMessage(`{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["order_id", "customer_id", "amount", "status", "created_at"],
	"properties": {
		"order_id": {"type": "string"},
		"customer_id": {"type": "string"},
		"amount": {"type": "number", "minimum": 0},
		"status": {"type": "string", "enum": ["created", "paid", "shipped", "delivered", "cancelled"]},
		"created_at": {"type": "string", "format": "date-time"}
	}
}`)

var NotificationEventSchema = json.RawMessage(`{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["type", "recipient", "subject", "body"],
	"properties": {
		"type": {"type": "string", "enum": ["email", "sms", "push"]},
		"recipient": {"type": "string"},
		"subject": {"type": "string"},
		"body": {"type": "string"},
		"metadata": {"type": "object"},
		"created_at": {"type": "string", "format": "date-time"}
	}
}`)

var UserEventSchema = json.RawMessage(`{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["user_id", "action", "timestamp"],
	"properties": {
		"user_id": {"type": "string"},
		"action": {"type": "string", "enum": ["login", "logout", "signup", "delete"]},
		"ip_address": {"type": "string"},
		"user_agent": {"type": "string"},
		"timestamp": {"type": "string", "format": "date-time"}
	}
}`)

var PaymentEventSchema = json.RawMessage(`{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["payment_id", "order_id", "amount", "currency", "status"],
	"properties": {
		"payment_id": {"type": "string"},
		"order_id": {"type": "string"},
		"amount": {"type": "number", "minimum": 0},
		"currency": {"type": "string", "enum": ["KRW", "USD", "EUR", "JPY"]},
		"status": {"type": "string", "enum": ["pending", "completed", "failed", "refunded"]},
		"method": {"type": "string", "enum": ["card", "bank", "virtual", "point"]},
		"created_at": {"type": "string", "format": "date-time"}
	}
}`)
