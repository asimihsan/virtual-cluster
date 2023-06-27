package utils

import (
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"time"
)

type KafkaWaiter struct {
	broker   string
	interval time.Duration
	timeout  time.Duration
}

type KafkaWaiterOption func(*KafkaWaiter)

func WithInterval(interval time.Duration) KafkaWaiterOption {
	return func(kw *KafkaWaiter) {
		kw.interval = interval
	}
}

func WithTimeout(timeout time.Duration) KafkaWaiterOption {
	return func(kw *KafkaWaiter) {
		kw.timeout = timeout
	}
}

func NewKafkaWaiter(broker string, opts ...KafkaWaiterOption) *KafkaWaiter {
	kw := &KafkaWaiter{
		broker:   broker,
		interval: 1 * time.Second,
		timeout:  5 * time.Second,
	}

	for _, opt := range opts {
		opt(kw)
	}

	return kw
}

func (kw *KafkaWaiter) Wait() error {
	timeoutTimer := time.NewTimer(kw.timeout)
	defer timeoutTimer.Stop()

	ticker := time.NewTicker(kw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return errors.New("kafka broker still unhealthy after timeout")
		case <-ticker.C:
			healthy, err := IsKafkaHealthy(kw.broker)
			if err != nil {
				return err
			}
			if healthy {
				return nil
			}
		}
	}
}

// IsKafkaHealthy checks if the Kafka broker is healthy.
func IsKafkaHealthy(
	broker string,
) (bool, error) {
	config := sarama.NewConfig()
	timeout := 5 * time.Second
	config.Net.DialTimeout = timeout
	config.Net.ReadTimeout = timeout
	config.Net.WriteTimeout = timeout

	client, err := sarama.NewClient([]string{broker}, config)
	if err != nil {
		return false, errors.Wrap(err, "failed to create kafka client")
	}
	defer func() {
		_ = client.Close()
	}()

	err = client.RefreshMetadata()
	if err != nil {
		return false, errors.Wrap(err, "failed to refresh kafka metadata")
	}

	return true, nil
}
