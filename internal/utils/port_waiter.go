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
	"github.com/pkg/errors"
	"net"
	"time"
)

type PortWaiter struct {
	BaseWaiter
	port string
}

func NewPortWaiter(port string, options ...WaiterOption) *PortWaiter {
	pw := &PortWaiter{
		BaseWaiter: BaseWaiter{
			interval: 1 * time.Second,
			timeout:  5 * time.Second,
		},
		port: port,
	}

	for _, option := range options {
		option(&pw.BaseWaiter)
	}

	return pw
}

func (pw *PortWaiter) Wait() error {
	return pw.BaseWaiter.Wait(pw)
}

func (pw *PortWaiter) CheckHealth() (bool, error) {
	listener, err := net.Listen("tcp", ":"+pw.port)
	if err != nil {
		return false, nil
	}

	err = listener.Close()
	if err != nil {
		return false, errors.Wrapf(err, "failed to close listener")
	}
	return true, nil
}
