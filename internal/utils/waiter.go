package utils

import (
	"time"
)

type Waiter interface {
	Wait() error
	CheckHealth() (bool, error)
}

type WaiterOption func(*BaseWaiter)

func WithInterval(interval time.Duration) WaiterOption {
	return func(bw *BaseWaiter) {
		bw.interval = interval
	}
}

func WithTimeout(timeout time.Duration) WaiterOption {
	return func(bw *BaseWaiter) {
		bw.timeout = timeout
	}
}

type BaseWaiter struct {
	interval time.Duration
	timeout  time.Duration
}

func (bw *BaseWaiter) Wait(w Waiter) error {
	timeoutTimer := time.NewTimer(bw.timeout)
	defer timeoutTimer.Stop()

	ticker := time.NewTicker(bw.interval)
	defer ticker.Stop()

	var healthy bool
	var lastErr error

	for {
		select {
		case <-timeoutTimer.C:
			return lastErr
		case <-ticker.C:
			healthy, lastErr = w.CheckHealth()
			if lastErr == nil && healthy {
				return nil
			}
		}
	}
}
