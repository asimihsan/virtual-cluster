/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

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
