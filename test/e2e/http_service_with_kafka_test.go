/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package e2e

import (
	"fmt"
	"github.com/Shopify/sarama"
	"net/http"
	"testing"
	"time"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/substrate"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServiceWithKafka(t *testing.T) {
	vclusterContent, err := utils.ReadFileUpward(
		"./test_services/http_service_with_kafka/http_service_with_kafka.vcluster",
		true, /*verbose*/
	)
	assert.NoError(t, err)

	ast, err := parser.ParseVCluster(string(vclusterContent))
	assert.NoError(t, err)

	manager, err := substrate.NewManager("/tmp/foo.sqlite3", substrate.WithVerbose())
	assert.NoError(t, err)
	defer func(manager *substrate.Manager) {
		err := manager.Close()
		if err != nil {
			t.Fatalf("failed to close manager: %v", err)
		}
	}(manager)

	err = manager.AddWorkingDirectoryUpward(
		"http_service_with_kafka",
		"./test_services/http_service_with_kafka",
		true, /*verbose*/
	)
	if err != nil {
		t.Fatalf("failed to add working directory: %v", err)
	}

	err = manager.StartServicesAndDependencies(
		[]*parser.VClusterAST{ast},
	)
	assert.NoError(t, err)

	// Wait for a short period to allow the managed Kafka dependency to start
	kw := utils.NewKafkaWaiter("localhost:9095")
	err = kw.Wait()
	assert.NoError(t, err)

	time.Sleep(5 * time.Second)

	// Send GET to /ping without proxy
	endpoint := fmt.Sprintf("http://localhost:%d/ping", *ast.Services[0].ServicePort)
	resp, err := http.Get(endpoint)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Send GET to /ping with proxy
	endpoint = fmt.Sprintf("http://localhost:%d/ping", *ast.Services[0].ProxyPort)
	resp, err = http.Get(endpoint)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Send a POST request to the /kafka endpoint
	endpoint = fmt.Sprintf("http://localhost:%d/kafka", *ast.Services[0].ProxyPort)
	resp, err = http.Post(endpoint, "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Wait for a short period to allow the message to be sent to Kafka
	time.Sleep(2 * time.Second)

	// Check if the message was sent to the "my-topic" Kafka topic
	consumer, err := sarama.NewConsumer([]string{"localhost:9095"}, nil)
	assert.NoError(t, err)

	partitionConsumer, err := consumer.ConsumePartition("my-topic", 0, sarama.OffsetOldest)
	assert.NoError(t, err)

	msg := <-partitionConsumer.Messages()
	assert.Equal(t, "Message 1", string(msg.Value))

	manager.StopAllProcesses()
	err = manager.Close()
	assert.NoError(t, err)
}
