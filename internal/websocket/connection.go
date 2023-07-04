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
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type ClientOption func(*Client)

func WithSendChannel(send chan []byte) ClientOption {
	return func(c *Client) {
		c.send = send
	}
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte),
	}
}

func (c *Client) Send(message []byte) {
	c.send <- message
}

func (c *Client) ReadPump() {
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("ReadPump: error closing websocket connection: %v", err)
		}
	}(c.conn)
	for {
		var msg []byte
		err := websocket.JSON.Receive(c.conn, &msg)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		// For now, we just log any message we receive. Later, we can add more logic here if needed.
		log.Printf("received: %s", msg)
	}
}

func (c *Client) WritePump() {
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("WritePump: error closing websocket connection: %v", err)
		}
	}(c.conn)
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The channel has been closed.
				return
			}

			err := websocket.JSON.Send(c.conn, message)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
		}
	}
}
