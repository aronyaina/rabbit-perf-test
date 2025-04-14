package consumer

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"rabbitmq-perftest/internal/config"
	"rabbitmq-perftest/internal/metrics"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.Config
	metrics *metrics.Metrics
}

func New(cfg *config.Config, metrics *metrics.Metrics) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitMQ.URI)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		cfg.RabbitMQ.Exchange,     // name
		cfg.RabbitMQ.ExchangeType, // type
		true,                      // durable
		cfg.Test.AutoDelete,       // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		cfg.RabbitMQ.Queue,  // name
		true,                // durable
		cfg.Test.AutoDelete, // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		cfg.RabbitMQ.Queue,      // queue name
		cfg.RabbitMQ.RoutingKey, // routing key
		cfg.RabbitMQ.Exchange,   // exchange
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if cfg.Test.QoS > 0 {
		if err := ch.Qos(cfg.Test.QoS, 0, false); err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		config:  cfg,
		metrics: metrics,
	}, nil
}

func (c *Consumer) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer c.cleanup()

	msgs, err := c.channel.Consume(
		c.config.RabbitMQ.Queue,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("Failed to start consuming: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-msgs:
			if !ok {
				return
			}
			start := time.Now()
			c.metrics.RecordConsumeLatency(time.Since(start))
			c.metrics.IncrementConsumed()
		}
	}
}

func (c *Consumer) cleanup() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

