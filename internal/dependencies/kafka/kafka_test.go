/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package kafka_test

import (
	"testing"

	"github.com/asimihsan/virtual-cluster/internal/dependencies/kafka"
	"github.com/stretchr/testify/assert"
)

func TestGenerateDockerComposeFile(t *testing.T) {
	tests := []struct {
		name       string
		kafkaPort  int
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "valid port",
			kafkaPort:  9092,
			wantOutput: "---\nversion: '2'\nservices:\n\n  broker:\n    image: confluentinc/cp-kafka:7.4.0\n    hostname: broker\n    container_name: broker\n    ports:\n      - \"9092:9092\"\n    environment:\n      KAFKA_NODE_ID: 1\n      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: 'CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT'\n      KAFKA_ADVERTISED_LISTENERS: 'PLAINTEXT://broker:29092,PLAINTEXT_HOST://localhost:9092'\n      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1\n      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0\n      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1\n      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1\n      KAFKA_PROCESS_ROLES: 'broker,controller'\n      KAFKA_CONTROLLER_QUORUM_VOTERS: '1@broker:29093'\n      KAFKA_LISTENERS: 'PLAINTEXT://broker:29092,CONTROLLER://broker:29093,PLAINTEXT_HOST://0.0.0.0:9092'\n      KAFKA_INTER_BROKER_LISTENER_NAME: 'PLAINTEXT'\n      KAFKA_CONTROLLER_LISTENER_NAMES: 'CONTROLLER'\n      KAFKA_LOG_DIRS: '/tmp/kraft-combined-logs'\n      KAFKA_LOG4J_LOGGERS: \"org.apache.kafka.image.loader.MetadataLoader=ERROR\"\n\n      # Replace CLUSTER_ID with a unique base64 UUID using \"bin/kafka-storage.sh random-uuid\"\n      # See https://docs.confluent.io/kafka/operations-tools/kafka-tools.html#kafka-storage-sh\n      CLUSTER_ID: '4GexY0RCRziZFDQu6KAXeQ'\n\n  kowl:\n    image: quay.io/cloudhut/kowl:v1.5.0\n    container_name: kowl\n    restart: always\n    ports:\n      - \"8080:8080\"\n    depends_on:\n      - broker\n    environment:\n      - KAFKA_BROKERS=broker:29092\n",
			wantErr:    false,
		},
		{
			name:       "invalid port",
			kafkaPort:  -1,
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, err := kafka.GenerateDockerComposeFile(tt.kafkaPort)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, gotOutput)
		})
	}
}
