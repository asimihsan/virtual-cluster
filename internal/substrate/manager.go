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
	"context"
	"database/sql"
	"fmt"
	"github.com/asimihsan/virtual-cluster/internal/dependencies/kafka"
	"github.com/asimihsan/virtual-cluster/internal/dependencies/localstack"
	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/proxy"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Manager struct {
	dbPath             string
	db                 *sql.DB
	processes          []*ManagedProcess
	workingDirectories map[string]string
	verbose            bool
}

type ManagerOption func(*Manager)

func WithVerbose() ManagerOption {
	return func(m *Manager) {
		m.verbose = true
	}
}

func NewManager(dbPath string, opts ...ManagerOption) (*Manager, error) {
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

	manager := &Manager{
		dbPath:             dbPath,
		db:                 db,
		workingDirectories: workingDirectories,
	}

	for _, opt := range opts {
		opt(manager)
	}

	return manager, nil
}

func (m *Manager) AddWorkingDirectoryUpward(serviceName, path string, verbose bool) error {
	if _, ok := m.workingDirectories[serviceName]; ok {
		return fmt.Errorf("service name already exists: %s", serviceName)
	}
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	stat, foundPath, err := utils.StatUpward(path, verbose)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path: %s", path)
	}
	if !stat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", foundPath)
	}

	m.workingDirectories[serviceName] = foundPath
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
		for _, managedDependency := range ast.ManagedDependencies {
			if managedDependency.ManagedKafka != nil {
				err := m.StartManagedKafka(managedDependency.Name, managedDependency.ManagedKafka.Port)
				if err != nil {
					return errors.Wrapf(err, "failed to start managed kafka: %s", managedDependency.Name)
				}
			} else if managedDependency.ManagedLocalstack != nil {
				err := m.StartManagedLocalstack(managedDependency.Name, managedDependency.ManagedLocalstack.Port)
				if err != nil {
					return errors.Wrapf(err, "failed to start managed localstack: %s", managedDependency.Name)
				}
			} else {
				return fmt.Errorf("unknown managed dependency type: %s", managedDependency.Name)
			}

		}

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

			go runProcessAndStoreOutput(process, m.db, m.verbose)
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

func (m *Manager) StartManagedKafka(
	managedDependencyName string,
	port int,
) error {
	dir, err := os.MkdirTemp("", "kafka")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}

	dockerComposeFile, err := kafka.GenerateDockerComposeFile(port)
	if err != nil {
		return errors.Wrap(err, "failed to generate docker compose file")
	}

	composeFilePath := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(composeFilePath, []byte(dockerComposeFile), 0644); err != nil {
		return errors.Wrap(err, "failed to write docker compose file")
	}

	fmt.Printf("Docker compose file location: %s\n", composeFilePath)

	fmt.Println("Cleaning up containers")
	cleanupContainers("broker1245")
	cleanupContainers("kowl12345")
	cleanupNetworks("my_custom_network")

	fmt.Println("Starting managed dependency:", managedDependencyName)
	workingDirectory := filepath.Dir(composeFilePath)
	process := &ManagedProcess{
		Name:             managedDependencyName,
		RunCommands:      []string{"docker compose up --no-color --timestamps"},
		WorkingDirectory: workingDirectory,
		Stop:             make(chan struct{}, 1),
	}
	m.processes = append(m.processes, process)
	go runProcessAndStoreOutput(process, m.db, m.verbose)
	fmt.Println("Started managed dependency:", managedDependencyName)

	kw := utils.NewKafkaWaiter(fmt.Sprintf("localhost:%d", port))
	err = kw.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to wait for kafka")
	}

	return nil
}

func (m *Manager) StartManagedLocalstack(
	managedDependencyName string,
	port int,
) error {
	dir, err := os.MkdirTemp("", "localstack")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}

	dockerComposeFile, err := localstack.GenerateDockerComposeFile(port)
	if err != nil {
		return errors.Wrap(err, "failed to generate docker compose file")
	}

	composeFilePath := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(composeFilePath, []byte(dockerComposeFile), 0644); err != nil {
		return errors.Wrap(err, "failed to write docker compose file")
	}

	fmt.Printf("Docker compose file location: %s\n", composeFilePath)

	fmt.Println("Cleaning up containers")
	cleanupContainers("localstack_main")
	cleanupNetworks("localstack_default")

	fmt.Println("Starting managed dependency:", managedDependencyName)
	workingDirectory := filepath.Dir(composeFilePath)
	process := &ManagedProcess{
		Name:             managedDependencyName,
		RunCommands:      []string{"docker compose up --no-color --timestamps"},
		WorkingDirectory: workingDirectory,
		Stop:             make(chan struct{}, 1),
	}
	m.processes = append(m.processes, process)
	go runProcessAndStoreOutput(process, m.db, m.verbose)
	fmt.Println("Started managed dependency:", managedDependencyName)

	localstackWaiter := utils.NewLocalStackWaiter(fmt.Sprintf("http://localhost:%d", port))
	err = localstackWaiter.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to wait for localstack")
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
	var proxyOptions []proxy.ProxyOption
	if m.verbose {
		proxyOptions = append(proxyOptions, proxy.WithVerbose(true))
	}

	httpProxy, err := proxy.NewProxy(target, processName, m.db, proxyOptions...)
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

func cleanupContainers(containerName string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "failed to create docker client")
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return errors.Wrap(err, "failed to list containers")
	}

	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			fmt.Printf("Removing container %s\n", container.ID)
			err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				return errors.Wrap(err, "failed to remove container "+container.ID)
			}
		}
	}

	return nil
}

func cleanupNetworks(networkName string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "failed to create docker client")
	}

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list networks")
	}

	for _, network := range networks {
		if network.Name == networkName {
			fmt.Printf("Removing network %s\n", network.ID)
			err := cli.NetworkRemove(ctx, network.ID)
			if err != nil {
				return errors.Wrap(err, "failed to remove network "+network.ID)
			}
		}
	}

	return nil
}
