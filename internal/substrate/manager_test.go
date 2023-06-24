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
	"strings"
	"testing"
	"time"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/substrate"
	"github.com/stretchr/testify/assert"
)

func TestStartAndStopSingleService(t *testing.T) {
	// Create a new Manager with a temporary SQLite database
	manager, err := substrate.NewManager(":memory:")
	assert.NoError(t, err)

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
	outputFound := false
	logs, err := manager.GetLogsForProcess("test-service", "stdout")
	assert.NoError(t, err)
	for _, content := range logs {
		fmt.Println("Found log:", content)
		if strings.Contains(content, "Service started") {
			outputFound = true
			break
		}
	}
	assert.True(t, outputFound, "Expected 'Service started' message not found in the output")

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
