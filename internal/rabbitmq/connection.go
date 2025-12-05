package rabbitmq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConnection(url string) (*Connection, error) {
	var conn *amqp.Connection
	var err error

	// 재시도 로직 (RabbitMQ 시작 대기)
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ 연결 실패, 재시도 중... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("RabbitMQ 연결 실패: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("채널 생성 실패: %w", err)
	}

	return &Connection{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

func (c *Connection) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
