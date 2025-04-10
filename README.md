# RabbitMQ Performance Testing Tool

A Go-based performance testing tool for RabbitMQ, similar to the official PerfTest tool but written in Go.

## Features

- Multiple producers and consumers
- Configurable message size and count
- Rate limiting for publishers
- Publisher confirms support
- Consumer QoS settings
- Prometheus metrics
- Duration-based or message count-based tests
- Configurable via command line flags or config file

## Building

```bash
go build -o perftest ./cmd/perftest
```

## Usage

Basic usage with default settings:

```bash
./perftest
```

Custom configuration:

```bash
./perftest \
  --uri="amqp://guest:guest@localhost:5672/" \
  --exchange="perf-test" \
  --queue="perf-test" \
  --size=1000 \
  --count=100000 \
  --producers=1 \
  --consumers=1 \
  --rate=1000 \
  --persistent \
  --confirm \
  --qos=100
```

## Metrics

The tool exposes Prometheus metrics at `:2112/metrics` including:

- Total messages published
- Total messages consumed
- Publishing latency
- Consuming latency

## Configuration Options

- `--uri`: RabbitMQ connection URI (default: "amqp://guest:guest@localhost:5672/")
- `--exchange`: Exchange name (default: "perf-test")
- `--exchange-type`: Exchange type (default: "direct")
- `--queue`: Queue name (default: "perf-test")
- `--routing-key`: Routing key (default: "perf-test")
- `--size`: Message size in bytes (default: 1000)
- `--count`: Number of messages to publish (default: 100000)
- `--consumers`: Number of consumers (default: 1)
- `--producers`: Number of producers (default: 1)
- `--rate`: Publishing rate limit in msgs/sec, 0 for unlimited (default: 0)
- `--persistent`: Use persistent messages (default: false)
- `--confirm`: Use publisher confirms (default: false)
- `--qos`: Consumer QoS prefetch count (default: 0)
- `--duration`: Test duration in seconds, 0 for message count based (default: 0)

## Project Structure

```
.
├── cmd/
│   └── perftest/
│       └── main.go           # Main application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration structures
│   ├── consumer/
│   │   └── consumer.go      # Consumer implementation
│   ├── producer/
│   │   └── producer.go      # Producer implementation
│   └── metrics/
│       └── metrics.go       # Metrics collection
├── go.mod                   # Go module file
└── README.md               # This file
```

## License

MIT 