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
	"encoding/json"
	"fmt"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"net/http"
	"testing"
	"time"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/substrate"
	"github.com/stretchr/testify/assert"
)

func TestHTTPService(t *testing.T) {
	vclusterContent, err := utils.ReadFileUpward("./test_services/http_service/http_service.vcluster", true /*verbose*/)
	assert.NoError(t, err)

	ast, err := parser.ParseVCluster(string(vclusterContent))
	assert.NoError(t, err)

	manager, err := substrate.NewManager(":memory:")
	assert.NoError(t, err)
	defer manager.Close()

	err = manager.AddWorkingDirectoryUpward("http_service", "./test_services/http_service", true /*verbose*/)
	if err != nil {
		t.Fatalf("failed to add working directory: %v", err)
	}

	err = manager.StartServicesAndDependencies(
		[]*parser.VClusterAST{ast},
	)
	assert.NoError(t, err)

	time.Sleep(5 * time.Second)

	endpoint := fmt.Sprintf("http://localhost:%d", *ast.Services[0].ProxyPort)
	resp, err := http.Get(endpoint)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	manager.StopAllProcesses()

	logs, err := manager.GetLogsForProcess("http_service", "stdout")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(logs))

	// if "bind: address already in use" then the previous test didn't clean up properly
	assert.NotContains(t, logs[0], "bind: address already in use")

	var logFields map[string]interface{}
	err = json.Unmarshal([]byte(logs[0]), &logFields)
	if err != nil {
		t.Fatalf("failed to decode log as JSON: %v", err)
	}

	// get the fields
	if _, ok := logFields["method"]; !ok {
		t.Fatalf("logFields did not have 'method' field")
	}
	method := logFields["method"].(string)

	if _, ok := logFields["uri"]; !ok {
		t.Fatalf("logFields did not have 'uri' field")
	}
	uri := logFields["uri"].(string)

	if _, ok := logFields["status"]; !ok {
		t.Fatalf("logFields did not have 'status' field")
	}
	status := int(logFields["status"].(float64))

	// assert the fields
	assert.Equal(t, "GET", method)
	assert.Equal(t, "/", uri)
	assert.Equal(t, 200, status)

	proxyRequests, err := manager.GetHTTPProxyRequestsForProcess("http_service")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(proxyRequests))
	assert.Equal(t, "GET", proxyRequests[0].Method)
	assert.Equal(t, "/", proxyRequests[0].URL)

	err = manager.Close()
	assert.NoError(t, err)
}
