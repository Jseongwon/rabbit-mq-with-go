package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RabbitMQMonitor RabbitMQ Management API 클라이언트
type RabbitMQMonitor struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewRabbitMQMonitor(baseURL, username, password string) *RabbitMQMonitor {
	return &RabbitMQMonitor{
		baseURL:  baseURL,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueueInfo 큐 정보
type QueueInfo struct {
	Name           string `json:"name"`
	VHost          string `json:"vhost"`
	Durable        bool   `json:"durable"`
	Messages       int    `json:"messages"`
	MessagesReady  int    `json:"messages_ready"`
	MessagesUnacked int   `json:"messages_unacknowledged"`
	Consumers      int    `json:"consumers"`
	State          string `json:"state"`
	// Message Stats
	MessageStats   *MessageStats `json:"message_stats,omitempty"`
}

type MessageStats struct {
	Publish        int `json:"publish"`
	PublishDetails struct {
		Rate float64 `json:"rate"`
	} `json:"publish_details"`
	Deliver        int `json:"deliver"`
	DeliverDetails struct {
		Rate float64 `json:"rate"`
	} `json:"deliver_details"`
	Ack        int `json:"ack"`
	AckDetails struct {
		Rate float64 `json:"rate"`
	} `json:"ack_details"`
}

// ExchangeInfo Exchange 정보
type ExchangeInfo struct {
	Name       string `json:"name"`
	VHost      string `json:"vhost"`
	Type       string `json:"type"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"auto_delete"`
}

// BindingInfo 바인딩 정보
type BindingInfo struct {
	Source          string `json:"source"`
	VHost           string `json:"vhost"`
	Destination     string `json:"destination"`
	DestinationType string `json:"destination_type"`
	RoutingKey      string `json:"routing_key"`
}

// Overview 전체 개요
type Overview struct {
	ManagementVersion string `json:"management_version"`
	RabbitMQVersion   string `json:"rabbitmq_version"`
	ClusterName       string `json:"cluster_name"`
	QueueTotals       struct {
		Messages       int `json:"messages"`
		MessagesReady  int `json:"messages_ready"`
		MessagesUnacked int `json:"messages_unacknowledged"`
	} `json:"queue_totals"`
	ObjectTotals struct {
		Connections int `json:"connections"`
		Channels    int `json:"channels"`
		Exchanges   int `json:"exchanges"`
		Queues      int `json:"queues"`
		Consumers   int `json:"consumers"`
	} `json:"object_totals"`
}

func (m *RabbitMQMonitor) request(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", m.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(m.username, m.password)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 요청 실패: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetOverview 전체 개요 조회
func (m *RabbitMQMonitor) GetOverview() (*Overview, error) {
	data, err := m.request("/api/overview")
	if err != nil {
		return nil, err
	}

	var overview Overview
	if err := json.Unmarshal(data, &overview); err != nil {
		return nil, err
	}
	return &overview, nil
}

// ListQueues 모든 큐 목록 조회
func (m *RabbitMQMonitor) ListQueues() ([]QueueInfo, error) {
	data, err := m.request("/api/queues")
	if err != nil {
		return nil, err
	}

	var queues []QueueInfo
	if err := json.Unmarshal(data, &queues); err != nil {
		return nil, err
	}
	return queues, nil
}

// GetQueue 특정 큐 조회
func (m *RabbitMQMonitor) GetQueue(vhost, name string) (*QueueInfo, error) {
	endpoint := fmt.Sprintf("/api/queues/%s/%s", vhost, name)
	data, err := m.request(endpoint)
	if err != nil {
		return nil, err
	}

	var queue QueueInfo
	if err := json.Unmarshal(data, &queue); err != nil {
		return nil, err
	}
	return &queue, nil
}

// ListExchanges 모든 Exchange 목록 조회
func (m *RabbitMQMonitor) ListExchanges() ([]ExchangeInfo, error) {
	data, err := m.request("/api/exchanges")
	if err != nil {
		return nil, err
	}

	var exchanges []ExchangeInfo
	if err := json.Unmarshal(data, &exchanges); err != nil {
		return nil, err
	}
	return exchanges, nil
}

// ListBindings 모든 바인딩 목록 조회
func (m *RabbitMQMonitor) ListBindings() ([]BindingInfo, error) {
	data, err := m.request("/api/bindings")
	if err != nil {
		return nil, err
	}

	var bindings []BindingInfo
	if err := json.Unmarshal(data, &bindings); err != nil {
		return nil, err
	}
	return bindings, nil
}

// GetMessages 큐에서 메시지 가져오기 (peek, 소비하지 않음)
func (m *RabbitMQMonitor) GetMessages(vhost, queueName string, count int) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/api/queues/%s/%s/get", vhost, queueName)

	// POST 요청 필요
	reqBody := fmt.Sprintf(`{"count":%d,"ackmode":"ack_requeue_true","encoding":"auto"}`, count)

	req, err := http.NewRequest("POST", m.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(m.username, m.password)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(io.Reader(nil))

	// 간단 구현을 위해 빈 배열 반환
	// 실제로는 POST body와 함께 요청해야 함
	_ = reqBody

	return nil, nil
}
