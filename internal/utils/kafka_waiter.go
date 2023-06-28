package utils

import (
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"time"
)

type KafkaWaiter struct {
	BaseWaiter
	broker string
}

func NewKafkaWaiter(broker string, opts ...WaiterOption) *KafkaWaiter {
	kw := &KafkaWaiter{
		BaseWaiter: BaseWaiter{
			interval: 1 * time.Second,
			timeout:  10 * time.Second,
		},
		broker: broker,
	}

	for _, opt := range opts {
		opt(&kw.BaseWaiter)
	}

	return kw
}

func (kw *KafkaWaiter) Wait() error {
	return kw.BaseWaiter.Wait(kw)
}

func (kw *KafkaWaiter) CheckHealth() (bool, error) {
	config := sarama.NewConfig()
	config.Net.DialTimeout = kw.interval
	config.Net.ReadTimeout = kw.interval
	config.Net.WriteTimeout = kw.interval

	client, err := sarama.NewClient([]string{kw.broker}, config)
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
