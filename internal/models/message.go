package models

import "time"

// OrderEvent 주문 이벤트 메시지
type OrderEvent struct {
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationEvent 알림 이벤트 메시지
type NotificationEvent struct {
	Type      string                 `json:"type"`       // email, sms, push
	Recipient string                 `json:"recipient"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// LogEvent 로그 이벤트 메시지
type LogEvent struct {
	Level     string                 `json:"level"`   // info, warn, error
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
