package producer

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"rabbitmq-perftest/internal/config"
	"rabbitmq-perftest/internal/metrics"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.Config
	metrics *metrics.Metrics
}

func New(cfg *config.Config, metrics *metrics.Metrics) (*Producer, error) {
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
		cfg.RabbitMQ.Queue, // name
		true,               // durable
		cfg.Test.AutoDelete,// auto-delete
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
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

	if cfg.Test.Confirm {
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}
	}

	return &Producer{
		conn:    conn,
		channel: ch,
		config:  cfg,
		metrics: metrics,
	}, nil
}

func (p *Producer) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer p.cleanup()

	confirms := make(chan amqp.Confirmation, 1)
	if p.config.Test.Confirm {
		p.channel.NotifyPublish(confirms)
	}

	payload := make([]byte, p.config.Test.MessageSize)
	publishing := amqp.Publishing{
		DeliveryMode: amqp.Transient,
		Body:        payload,
	}
	if p.config.Test.Persistent {
		publishing.DeliveryMode = amqp.Persistent
	}

	ticker := time.NewTicker(time.Second / time.Duration(p.config.Test.PublishRate))
	defer ticker.Stop()

	for i := 0; i < p.config.Test.MessageCount; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			err := p.channel.PublishWithContext(ctx,
				p.config.RabbitMQ.Exchange,
				p.config.RabbitMQ.RoutingKey,
				p.config.Test.Mandatory,
				p.config.Test.Immediate,
				publishing,
			)
			if err != nil {
				log.Printf("Failed to publish message: %v", err)
				continue
			}

			if p.config.Test.Confirm {
				if confirmed := <-confirms; confirmed.Ack {
					p.metrics.RecordPublishLatency(time.Since(start))
					p.metrics.IncrementPublished()
				}
			} else {
				p.metrics.RecordPublishLatency(time.Since(start))
				p.metrics.IncrementPublished()
			}
		}
	}
}

func (p *Producer) cleanup() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
} 