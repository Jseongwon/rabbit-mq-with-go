package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn     *Connection
	exchange string
}

func NewPublisher(conn *Connection, exchange, exchangeType string) (*Publisher, error) {
	// Exchange 선언
	err := conn.Channel().ExchangeDeclare(
		exchange,     // name
		exchangeType, // type: direct, fanout, topic, headers
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("exchange 선언 실패: %w", err)
	}

	return &Publisher{
		conn:     conn,
		exchange: exchange,
	}, nil
}

// Publish 메시지 발행
func (p *Publisher) Publish(ctx context.Context, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("메시지 직렬화 실패: %w", err)
	}

	return p.conn.Channel().PublishWithContext(
		ctx,
		p.exchange,   // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // 메시지 영속성
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// PublishWithHeaders 헤더와 함께 메시지 발행
func (p *Publisher) PublishWithHeaders(ctx context.Context, routingKey string, message interface{}, headers map[string]interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("메시지 직렬화 실패: %w", err)
	}

	return p.conn.Channel().PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      headers,
			Body:         body,
		},
	)
}
