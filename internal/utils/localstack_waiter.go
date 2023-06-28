package utils

import (
	"net/http"
	"time"
)

type LocalStackWaiter struct {
	BaseWaiter
	endpoint string
}

func NewLocalStackWaiter(endpoint string, opts ...WaiterOption) *LocalStackWaiter {
	lw := &LocalStackWaiter{
		BaseWaiter: BaseWaiter{
			interval: 1 * time.Second,
			timeout:  10 * time.Second,
		},
		endpoint: endpoint,
	}

	for _, opt := range opts {
		opt(&lw.BaseWaiter)
	}

	return lw
}

func (lw *LocalStackWaiter) Wait() error {
	return lw.BaseWaiter.Wait(lw)
}

func (lw *LocalStackWaiter) CheckHealth() (bool, error) {
	resp, err := http.Get(lw.endpoint)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
