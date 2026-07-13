package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, routingKey string, body []byte) error

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	handler MessageHandler
}

func NewConsumer(url string, handler MessageHandler) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := declareBackendConsumerTopology(ch); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Consumer{conn: conn, channel: ch, handler: handler}, nil
}

func declareBackendConsumerTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(BackendExchangeName, ExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare backend exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(SMSWorkerExchangeName, ExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare sms worker exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(EmailWorkerExchangeName, ExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare email worker exchange: %w", err)
	}

	q, err := ch.QueueDeclare(BackendQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare backend queue: %w", err)
	}

	workerResultBindings := []struct {
		exchange   string
		routingKey string
	}{
		{SMSWorkerExchangeName, RoutingNotificationSMSProcessing},
		{SMSWorkerExchangeName, RoutingNotificationSMSSent},
		{SMSWorkerExchangeName, RoutingNotificationSMSFailed},
		{EmailWorkerExchangeName, RoutingNotificationEmailProcessing},
		{EmailWorkerExchangeName, RoutingNotificationEmailSent},
		{EmailWorkerExchangeName, RoutingNotificationEmailFailed},
	}

	for _, binding := range workerResultBindings {
		if err := ch.QueueBind(q.Name, binding.routingKey, binding.exchange, false, nil); err != nil {
			return fmt.Errorf("bind backend queue to %s: %w", binding.exchange, err)
		}
	}

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(BackendQueueName, "meetoria-backend", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume backend queue: %w", err)
	}

	slog.Info("backend rabbitmq consumer started", "queue", BackendQueueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			if err := c.handler(ctx, msg.RoutingKey, msg.Body); err != nil {
				slog.Error("handle rabbitmq message failed", "routing_key", msg.RoutingKey, "error", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
