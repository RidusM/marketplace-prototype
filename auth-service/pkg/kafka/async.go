package kafka

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

var _defaultAsyncRetryBackoff = 10 * time.Millisecond

type AsyncProducer struct {
	asyncProducer sarama.AsyncProducer
	topic         string
}

func NewAsyncProducer(brokers []string, topic string) (*AsyncProducer, error) {
	cfg := sarama.NewConfig()

	cfg.ClientID = "auth-service"

	cfg.Version = sarama.V3_3_2_0

	cfg.Metadata.AllowAutoTopicCreation = false

	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.Retry.Max = 30
	cfg.Producer.Retry.Backoff = _defaultAsyncRetryBackoff
	cfg.Producer.Compression = sarama.CompressionZSTD
	cfg.Producer.RequiredAcks = sarama.WaitForLocal

	cfg.Net.MaxOpenRequests = 1

	producer, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka.async.NewAsyncProducer: %w", err)
	}

	return &AsyncProducer{
		asyncProducer: producer,
		topic:         topic,
	}, nil
}

func (p *AsyncProducer) SendMessage(key, value string) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	p.asyncProducer.Input() <- msg
}

func (p *AsyncProducer) Close() error {
	if err := p.asyncProducer.Close(); err != nil {
		return fmt.Errorf("kafka.async.Close: %w", err)
	}

	return nil
}
