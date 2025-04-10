package config

type Config struct {
	RabbitMQ RabbitMQConfig
	Test     TestConfig
}

type RabbitMQConfig struct {
	URI          string
	Exchange     string
	ExchangeType string
	Queue        string
	RoutingKey   string
}

type TestConfig struct {
	MessageSize    int
	MessageCount   int
	ConsumerCount  int
	ProducerCount  int
	PublishRate    int // messages per second, 0 for unlimited
	Persistent     bool
	AutoDelete     bool
	Mandatory      bool
	Immediate      bool
	Confirm        bool
	QoS           int
	Duration      int // test duration in seconds, 0 for message count based
} 