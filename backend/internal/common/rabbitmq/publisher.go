package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ExchangeName is the backend outbound exchange.
const ExchangeName = BackendExchangeName

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(BackendExchangeName, ExchangeType, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	return p.channel.PublishWithContext(ctx, BackendExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
	})
}

func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
