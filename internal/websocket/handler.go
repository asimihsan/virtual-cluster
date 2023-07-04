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
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

func WebSocketHandler(broadcaster *Broadcaster) http.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		defer func(conn *websocket.Conn) {
			err := conn.Close()
			if err != nil {
				log.Printf("error closing websocket connection: %v", err)
			}
		}(conn)

		client := NewClient(conn)
		broadcaster.AddClient(client)

		go client.WritePump()

		// This is a blocking call.
		client.ReadPump()

		broadcaster.RemoveClient(client)
	})
}
