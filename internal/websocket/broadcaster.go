/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package websocket

import (
	"sync"
)

type Broadcaster struct {
	clients map[*Client]bool
	mu      sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[*Client]bool),
	}
}

func (b *Broadcaster) AddClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients[client] = true
}

func (b *Broadcaster) RemoveClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.clients, client)
}

func (b *Broadcaster) Broadcast(message []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		client.Send(message)
	}
}
