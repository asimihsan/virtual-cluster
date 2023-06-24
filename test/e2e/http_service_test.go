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
	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/substrate"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestHTTPService(t *testing.T) {
	vclusterContent, err := os.ReadFile("./test_services/http_service/http_service.vcluster")
	assert.NoError(t, err)

	ast, err := parser.ParseVCluster(string(vclusterContent))
	assert.NoError(t, err)

	manager, err := substrate.NewManager(":memory:")
	assert.NoError(t, err)
	defer manager.Close()

	workingDirectories := make(map[string]string)
	workingDirectories["http_service"] = "./test_services/http_service"

	err = manager.StartServicesAndDependencies(
		[]*parser.VClusterAST{ast},
		workingDirectories,
	)
	assert.NoError(t, err)

	time.Sleep(5 * time.Second)

	resp, err := http.Get("http://localhost:1323")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	manager.StopAllProcesses()

	logs, err := manager.GetLogsForProcess("http_service", "stdout")
	assert.NoError(t, err)
	fmt.Println(logs)
}
