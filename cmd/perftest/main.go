package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"rabbitmq-perftest/internal/config"
	"rabbitmq-perftest/internal/consumer"
	"rabbitmq-perftest/internal/metrics"
	"rabbitmq-perftest/internal/producer"
)

func main() {
	var cfgFile string
	var cfg config.Config

	rootCmd := &cobra.Command{
		Use:   "perftest",
		Short: "RabbitMQ Performance Testing Tool",
		Run: func(cmd *cobra.Command, args []string) {
			runPerfTest(cfg)
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
	rootCmd.PersistentFlags().StringVar(&cfg.RabbitMQ.URI, "uri", "amqp://guest:guest@localhost:5672/", "RabbitMQ URI")
	rootCmd.PersistentFlags().StringVar(&cfg.RabbitMQ.Exchange, "exchange", "perf-test", "Exchange name")
	rootCmd.PersistentFlags().StringVar(&cfg.RabbitMQ.ExchangeType, "exchange-type", "direct", "Exchange type")
	rootCmd.PersistentFlags().StringVar(&cfg.RabbitMQ.Queue, "queue", "perf-test", "Queue name")
	rootCmd.PersistentFlags().StringVar(&cfg.RabbitMQ.RoutingKey, "routing-key", "perf-test", "Routing key")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.MessageSize, "size", 1000, "Message size in bytes")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.MessageCount, "count", 100000, "Number of messages to publish")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.ConsumerCount, "consumers", 1, "Number of consumers")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.ProducerCount, "producers", 1, "Number of producers")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.PublishRate, "rate", 0, "Publishing rate limit (msgs/sec), 0 for unlimited")
	rootCmd.PersistentFlags().BoolVar(&cfg.Test.Persistent, "persistent", false, "Use persistent messages")
	rootCmd.PersistentFlags().BoolVar(&cfg.Test.Confirm, "confirm", false, "Use publisher confirms")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.QoS, "qos", 0, "Consumer QoS prefetch count")
	rootCmd.PersistentFlags().IntVar(&cfg.Test.Duration, "duration", 0, "Test duration in seconds (0 for message count based)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runPerfTest(cfg config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup metrics
	metrics := metrics.New()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	var wg sync.WaitGroup

	// Start consumers
	for i := 0; i < cfg.Test.ConsumerCount; i++ {
		consumer, err := consumer.New(&cfg, metrics)
		if err != nil {
			log.Printf("Failed to create consumer %d: %v", i, err)
			continue
		}
		wg.Add(1)
		go consumer.Run(ctx, &wg)
	}

	// Start producers
	for i := 0; i < cfg.Test.ProducerCount; i++ {
		producer, err := producer.New(&cfg, metrics)
		if err != nil {
			log.Printf("Failed to create producer %d: %v", i, err)
			continue
		}
		wg.Add(1)
		go producer.Run(ctx, &wg)
	}

	// If duration is set, cancel after duration
	if cfg.Test.Duration > 0 {
		go func() {
			time.Sleep(time.Duration(cfg.Test.Duration) * time.Second)
			cancel()
		}()
	}

	wg.Wait()
} 