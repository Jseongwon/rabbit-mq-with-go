package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *Connection
	queueName string
}

type ConsumerConfig struct {
	QueueName    string
	Exchange     string
	RoutingKey   string
	DLQExchange  string // Dead Letter Exchange
	DLQQueue     string // Dead Letter Queue
	MaxRetries   int32  // 최대 재시도 횟수
	TTL          int32  // 메시지 TTL (밀리초)
	PrefetchCount int   // Consumer가 한 번에 가져올 메시지 수
}

func NewConsumer(conn *Connection, config ConsumerConfig) (*Consumer, error) {
	ch := conn.Channel()

	// Prefetch 설정 (한 번에 처리할 메시지 수 제한)
	if config.PrefetchCount > 0 {
		err := ch.Qos(config.PrefetchCount, 0, false)
		if err != nil {
			return nil, fmt.Errorf("QoS 설정 실패: %w", err)
		}
	}

	// DLQ 설정이 있으면 DLQ 먼저 생성
	args := make(amqp.Table)
	if config.DLQExchange != "" {
		// DLQ Exchange 선언
		err := ch.ExchangeDeclare(
			config.DLQExchange,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("DLQ exchange 선언 실패: %w", err)
		}

		// DLQ Queue 선언
		_, err = ch.QueueDeclare(
			config.DLQQueue,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("DLQ queue 선언 실패: %w", err)
		}

		// DLQ 바인딩
		err = ch.QueueBind(
			config.DLQQueue,
			config.QueueName, // DLQ routing key = 원본 큐 이름
			config.DLQExchange,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("DLQ 바인딩 실패: %w", err)
		}

		// 원본 큐에 DLQ 설정 추가
		args["x-dead-letter-exchange"] = config.DLQExchange
		args["x-dead-letter-routing-key"] = config.QueueName
	}

	// TTL 설정
	if config.TTL > 0 {
		args["x-message-ttl"] = config.TTL
	}

	// 메인 Queue 선언
	_, err := ch.QueueDeclare(
		config.QueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("queue 선언 실패: %w", err)
	}

	// Exchange에 Queue 바인딩
	if config.Exchange != "" {
		err = ch.QueueBind(
			config.QueueName,
			config.RoutingKey,
			config.Exchange,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("queue 바인딩 실패: %w", err)
		}
	}

	return &Consumer{
		conn:      conn,
		queueName: config.QueueName,
	}, nil
}

// MessageHandler 메시지 처리 함수 타입
type MessageHandler func(delivery amqp.Delivery) error

// Consume 메시지 소비 시작
func (c *Consumer) Consume(handler MessageHandler) error {
	msgs, err := c.conn.Channel().Consume(
		c.queueName,
		"",    // consumer tag
		false, // auto-ack (false = 수동 ACK)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume 시작 실패: %w", err)
	}

	log.Printf("[*] %s 큐에서 메시지 대기 중...", c.queueName)

	for msg := range msgs {
		log.Printf("[📩] 메시지 수신: %s", string(msg.Body))

		err := handler(msg)
		if err != nil {
			log.Printf("[❌] 메시지 처리 실패: %v", err)
			// NACK - 메시지를 DLQ로 보냄 (requeue=false)
			msg.Nack(false, false)
		} else {
			log.Printf("[✅] 메시지 처리 완료")
			msg.Ack(false)
		}
	}

	return nil
}
