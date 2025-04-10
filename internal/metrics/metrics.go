package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	publishedMessages  prometheus.Counter
	consumedMessages  prometheus.Counter
	publishLatency    prometheus.Histogram
	consumeLatency    prometheus.Histogram
}

func New() *Metrics {
	return &Metrics{
		publishedMessages: promauto.NewCounter(prometheus.CounterOpts{
			Name: "rabbitmq_perftest_published_messages_total",
			Help: "The total number of published messages",
		}),
		consumedMessages: promauto.NewCounter(prometheus.CounterOpts{
			Name: "rabbitmq_perftest_consumed_messages_total",
			Help: "The total number of consumed messages",
		}),
		publishLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "rabbitmq_perftest_publish_latency_seconds",
			Help:    "Publish latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		consumeLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "rabbitmq_perftest_consume_latency_seconds",
			Help:    "Consume latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}
}

func (m *Metrics) IncrementPublished() {
	m.publishedMessages.Inc()
}

func (m *Metrics) IncrementConsumed() {
	m.consumedMessages.Inc()
}

func (m *Metrics) RecordPublishLatency(d time.Duration) {
	m.publishLatency.Observe(d.Seconds())
}

func (m *Metrics) RecordConsumeLatency(d time.Duration) {
	m.consumeLatency.Observe(d.Seconds())
} 