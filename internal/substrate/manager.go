/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package substrate

import (
	"database/sql"
	"fmt"
	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/proxy"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"sync"
)

type Manager struct {
	dbPath             string
	db                 *sql.DB
	processes          []*ManagedProcess
	workingDirectories map[string]string
	wg                 sync.WaitGroup
}

func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA synchronous = FULL")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			process_name TEXT,
			output_type TEXT,
			content TEXT
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS http_requests (
			id INTEGER PRIMARY KEY,
			process_name TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			method TEXT,
			url TEXT,
			headers TEXT,
			body TEXT
		)
	`)
	if err != nil {
		return nil, err
	}

	workingDirectories := make(map[string]string)

	return &Manager{
		dbPath:             dbPath,
		db:                 db,
		workingDirectories: workingDirectories,
	}, nil
}

func (m *Manager) AddWorkingDirectory(serviceName, path string) error {
	if _, ok := m.workingDirectories[serviceName]; ok {
		return fmt.Errorf("service name already exists: %s", serviceName)
	}
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	stat, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path: %s", path)
	}
	if !stat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	m.workingDirectories[serviceName] = path
	return nil
}

func (m *Manager) Close() error {
	m.StopAllProcesses()
	return m.db.Close()
}

// takes the combined slice of VClusterServiceDefinitionAST and VClusterManagedDependencyDefinitionAST and
// performs a topological sort.
func topologicalSort(definitions []interface{}) ([]interface{}, error) {
	// Perform topological sort on the combined slice

	return definitions, nil
}

func (m *Manager) StartServicesAndDependencies(
	asts []*parser.VClusterAST,
) error {
	// Combine and check for duplicate names
	// Perform topological sort
	// Start services and dependencies one at a time

	for _, ast := range asts {
		for _, service := range ast.Services {
			workingDirectory, ok := m.workingDirectories[service.Name]
			if !ok {
				workingDirectory = "."
			}

			fmt.Println("Starting service:", service.Name)
			process := &ManagedProcess{
				Name:             service.Name,
				RunCommands:      service.RunCommands,
				WorkingDirectory: workingDirectory,
				Stop:             make(chan struct{}, 1),
			}
			m.processes = append(m.processes, process)
			m.wg.Add(1)

			go runProcessAndStoreOutput(process, m.db)
			fmt.Println("Started service:", service.Name)

			if service.ServicePort != nil && service.ProxyPort != nil {
				fmt.Println("Starting HTTP proxy for service:", service.Name)
				err := m.RunHTTPProxy(
					fmt.Sprintf("http://localhost:%d", *service.ServicePort),
					fmt.Sprintf(":%d", *service.ProxyPort),
					service.Name)
				if err != nil {
					return err
				}
				fmt.Println("Started HTTP proxy for service:", service.Name)
			}
		}
	}

	return nil
}

func (m *Manager) StopAllProcesses() {
	for _, process := range m.processes {
		select {
		case process.Stop <- struct{}{}:
			// value sent successfully
			fmt.Println("Sent stop signal to process:", process.Name)
		default:
			// channel is not ready to receive, handle the situation accordingly
			fmt.Println("Channel not ready to receive for process:", process.Name)
		}
	}
}

func (m *Manager) GetLogsForProcess(processName string, outputType string) ([]string, error) {
	rows, err := m.db.Query("SELECT content FROM logs WHERE process_name = ? AND output_type = ?", processName, outputType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []string
	for rows.Next() {
		var content string
		err = rows.Scan(&content)
		if err != nil {
			return nil, err
		}
		logs = append(logs, content)
	}

	return logs, nil
}

type HTTPProxyRequest struct {
	Timestamp string
	Method    string
	URL       string
	Headers   string
	Body      string
}

func (m *Manager) GetHTTPProxyRequestsForProcess(
	processName string,
) ([]*HTTPProxyRequest, error) {
	rows, err := m.db.Query("SELECT timestamp, method, url, headers, body FROM http_requests WHERE process_name = ?", processName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*HTTPProxyRequest
	for rows.Next() {
		var request HTTPProxyRequest
		err = rows.Scan(&request.Timestamp, &request.Method, &request.URL, &request.Headers, &request.Body)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

func (m *Manager) RunHTTPProxy(
	target string,
	listenAddr string,
	processName string,
) error {
	httpProxy, err := proxy.NewProxy(target, processName, m.db)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("Starting HTTP proxy on %s", listenAddr)
		if err := http.ListenAndServe(listenAddr, httpProxy); err != nil {
			log.Printf("Error starting HTTP proxy: %v", err)
		}
	}()

	return nil
}
