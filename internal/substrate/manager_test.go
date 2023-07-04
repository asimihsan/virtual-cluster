/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package substrate_test

import (
	"fmt"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"golang.org/x/net/websocket"
	"testing"
	"time"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/substrate"
	"github.com/stretchr/testify/assert"
)

type LogMessage struct {
	Content     string `json:"content"`
	OutputType  string `json:"output_type"`
	ProcessName string `json:"process_name"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
}

func TestStartAndStopSingleService(t *testing.T) {
	// Create a new Manager with a temporary SQLite database
	manager, err := substrate.NewManager(
		":memory:",
		substrate.WithHTTPPort(1371),
	)
	assert.NoError(t, err)

	// Connect to Websocket broadcast
	ws, err := websocket.Dial("ws://localhost:1371/ws", "", "http://localhost/")
	assert.NoError(t, err)
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			t.Fatalf("failed to close websocket: %v", err)
		}
	}(ws)

	// Gather broadcasted messages into a slice
	broadcastedMessages := make([]*LogMessage, 0)
	go func() {
		for {
			var message LogMessage
			err := websocket.JSON.Receive(ws, &message)
			if err != nil {
				return
			}
			broadcastedMessages = append(broadcastedMessages, &message)
		}
	}()

	// Define a simple service that runs forever using a bash command
	service := &parser.VClusterServiceDefinitionAST{
		Name: "test-service",
		RunCommands: []string{
			"echo 'Service started'; sleep 60",
		},
	}

	// Start the service
	fmt.Println("Starting service")
	err = manager.StartServicesAndDependencies([]*parser.VClusterAST{
		{
			Services: []parser.VClusterServiceDefinitionAST{*service},
		},
	})
	assert.NoError(t, err)
	fmt.Println("Service started")

	// Wait for a short period to allow the service to start
	time.Sleep(2 * time.Second)

	// Check if the service is actually running by looking for the "Service started" message in the output
	fmt.Println("Checking if service is running")
	logs, err := manager.GetLogsForProcess("test-service", "stdout")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, "Service started\n", logs[0])

	assert.Equal(t, 1, len(broadcastedMessages))
	assert.Equal(t, "Service started\n", broadcastedMessages[0].Content)
	assert.Equal(t, "stdout", broadcastedMessages[0].OutputType)
	assert.Equal(t, "test-service", broadcastedMessages[0].ProcessName)
	assert.Equal(t, "log", broadcastedMessages[0].Type)

	// Stop the service
	fmt.Println("Stopping service")
	manager.StopAllProcesses()
	fmt.Println("Service stopped")

	// Close the manager and clean up resources
	fmt.Println("Closing manager")
	err = manager.Close()
	assert.NoError(t, err)
	fmt.Println("Manager closed")
}

// virtual-cluster will never just start a single managed dependency, but having this test here is useful for ensuring
// that the manager can start and stop Kafka and Kafka actually works.
func TestStartAndStopManagedKafka(t *testing.T) {
	// Create a new Manager with a temporary SQLite database
	manager, err := substrate.NewManager(":memory:")
	assert.NoError(t, err)

	// Define managed Kafka dependency.
	managedKafka := &parser.VClusterManagedDependencyDefinitionAST{
		Name:         "kafka",
		ManagedKafka: &parser.ManagedKafka{Port: 9095},
	}

	// Start the managed Kafka dependency
	fmt.Println("Starting managed Kafka dependency")
	err = manager.StartServicesAndDependencies([]*parser.VClusterAST{
		{
			ManagedDependencies: []parser.VClusterManagedDependencyDefinitionAST{*managedKafka},
		},
	})
	assert.NoError(t, err)
	fmt.Println("Managed Kafka dependency started")

	// Wait for a short period to allow the managed Kafka dependency to start
	kw := utils.NewKafkaWaiter("localhost:9095")
	err = kw.Wait()
	assert.NoError(t, err)

	// Stop the managed Kafka dependency
	fmt.Println("Stopping managed Kafka dependency")
	manager.StopAllProcesses()
	fmt.Println("Managed Kafka dependency stopped")

	logs, err := manager.GetLogsForProcess("kafka", "stdout")
	assert.NoError(t, err)
	for _, content := range logs {
		fmt.Println("Found log:", content)
	}

	// Close the manager and clean up resources
	fmt.Println("Closing manager")
	err = manager.Close()
	assert.NoError(t, err)
	fmt.Println("Manager closed")
}
