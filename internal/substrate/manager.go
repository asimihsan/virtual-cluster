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
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/asimihsan/virtual-cluster/internal/dependencies/kafka"
	"github.com/asimihsan/virtual-cluster/internal/dependencies/localstack"
	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/proxy"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"github.com/asimihsan/virtual-cluster/internal/websocket"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Manager struct {
	dbPath             string
	db                 *sql.DB
	processes          []*ManagedProcess
	workingDirectories map[string]string
	verbose            bool
	httpPort           int
	websocket          *websocket.Broadcaster
}

func (m *Manager) Websocket() *websocket.Broadcaster {
	return m.websocket
}

type ManagerOption func(*Manager)

func WithVerbose() ManagerOption {
	return func(m *Manager) {
		m.verbose = true
	}
}

func WithHTTPPort(port int) ManagerOption {
	return func(m *Manager) {
		m.httpPort = port
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
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			process_name TEXT,
			method TEXT,
			url TEXT,
			headers TEXT,
			body TEXT
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kafka_messages (
			id INTEGER PRIMARY KEY,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			broker_name TEXT,
			topic_name TEXT,
			message_key TEXT,
			message_value TEXT
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
		websocket:          websocket.NewBroadcaster(),
		httpPort:           1371,
	}

	for _, opt := range opts {
		opt(manager)
	}

	go func() {
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.GET("/ws", func(c echo.Context) error {
			websocket.WebSocketHandler(manager.Websocket()).ServeHTTP(c.Response(), c.Request())
			return nil
		})
		manager.BroadcastLogsAndRequests()
		err := e.Start(fmt.Sprintf(":%d", manager.httpPort))
		if err != nil {
			log.Printf("failed to start http server: %s", err)
		}
	}()

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

	go func() {
		err = m.ConsumeAndStoreKafkaMessages(managedDependencyName, port)
		if err != nil {
			fmt.Println("failed to consume and store kafka messages:", err)
		}
	}()

	return nil
}

func (m *Manager) ConsumeAndStoreKafkaMessages(brokerName string, port int) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Metadata.RefreshFrequency = 1 * time.Second

	// Connect to the Kafka broker
	broker := fmt.Sprintf("localhost:%d", port)
	kafkaClient, err := sarama.NewClient([]string{broker}, config)
	if err != nil {
		return err
	}
	defer func(kafkaClient sarama.Client) {
		err := kafkaClient.Close()
		if err != nil {
			fmt.Println("failed to close kafka client:", err)
		}
	}(kafkaClient)

	// Keep track of the topics we're already consuming
	consumingTopics := make(map[string]bool)

	for {
		topics, err := kafkaClient.Topics()
		if err != nil {
			return err
		}

		// For each topicName, if we're not already consuming it, start a consumer
		for _, topicName := range topics {
			if consumingTopics[topicName] {
				continue
			}

			consumingTopics[topicName] = true

			consumer, err := sarama.NewConsumerFromClient(kafkaClient)
			if err != nil {
				return err
			}

			partitionConsumer, err := consumer.ConsumePartition(topicName, 0, sarama.OffsetOldest)
			if err != nil {
				return err
			}

			topic := topicName
			go func() {
				fmt.Printf("Consuming messages from topic: %s\n", topic)
				for message := range partitionConsumer.Messages() {
					fmt.Printf("Consumed message from topic: %s\n", topic)
					fmt.Printf("Message: %s\n", string(message.Value))

					// For each message, store it in the SQLite database
					_, err := m.db.Exec("INSERT INTO kafka_messages (broker_name, topic_name, message_key, message_value, timestamp) VALUES (?, ?, ?, ?, ?)",
						brokerName, topic, string(message.Key), string(message.Value), message.Timestamp)
					if err != nil {
						log.Printf("Failed to insert message into database: %v", err)
					}
				}
			}()
		}

		// Wait for a bit before checking for new topics
		time.Sleep(1 * time.Second)
	}
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

func (m *Manager) BroadcastLogsAndRequests() {
	go func() {
		var lastLogID, lastRequestID, lastKafkaMessageID int
		for {
			// Query logs
			rows, err := m.db.Query(`SELECT id, timestamp, process_name, output_type, content FROM logs WHERE id > ? ORDER BY id ASC LIMIT 100`, lastLogID)
			if err != nil {
				log.Printf("error querying logs: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			for rows.Next() {
				var id int
				var timestamp time.Time
				var processName, outputType, content string
				err = rows.Scan(&id, &timestamp, &processName, &outputType, &content)
				if err != nil {
					log.Printf("error scanning log row: %v", err)
					continue
				}

				lastLogID = id
				message, _ := json.Marshal(map[string]interface{}{
					"type":         "log",
					"timestamp":    timestamp,
					"process_name": processName,
					"output_type":  outputType,
					"content":      content,
				})
				m.websocket.Broadcast(message)
			}
			err = rows.Close()
			if err != nil {
				log.Printf("error closing rows for logs: %v", err)
			}

			// Query HTTP requests
			rows, err = m.db.Query(`SELECT id, timestamp, process_name, method, url, headers, body FROM http_requests WHERE id > ? ORDER BY id ASC LIMIT 100`, lastRequestID)
			if err != nil {
				log.Printf("error querying http_requests: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			for rows.Next() {
				var id int
				var timestamp time.Time
				var processName, method, url, headers, body string
				err = rows.Scan(&id, &timestamp, &processName, &method, &url, &headers, &body)
				if err != nil {
					log.Printf("error scanning http_request row: %v", err)
					continue
				}

				lastRequestID = id
				message, _ := json.Marshal(map[string]interface{}{
					"type":         "http_request",
					"timestamp":    timestamp,
					"process_name": processName,
					"method":       method,
					"url":          url,
					"headers":      headers,
					"body":         body,
				})
				m.websocket.Broadcast(message)
			}
			err = rows.Close()
			if err != nil {
				log.Printf("error closing rows for http_requests: %v", err)
			}

			// Query Kafka messages
			rows, err = m.db.Query(`SELECT id, broker_name, topic_name, message_key, message_value, timestamp FROM kafka_messages WHERE id > ? ORDER BY id ASC LIMIT 100`, lastKafkaMessageID)
			if err != nil {
				log.Printf("error querying kafka_messages: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			for rows.Next() {
				var id int
				var timestamp time.Time
				var brokerName, topicName, messageKey, messageValue string
				err = rows.Scan(&id, &brokerName, &topicName, &messageKey, &messageValue, &timestamp)
				if err != nil {
					log.Printf("error scanning kafka_message row: %v", err)
					continue
				}

				lastKafkaMessageID = id
				messagePayload, _ := json.Marshal(map[string]interface{}{
					"type":          "kafka_message",
					"timestamp":     timestamp,
					"broker_name":   brokerName,
					"topic_name":    topicName,
					"message_key":   messageKey,
					"message_value": messageValue,
				})
				m.websocket.Broadcast(messagePayload)
			}
			err = rows.Close()
			if err != nil {
				log.Printf("error closing rows for kafka_messages: %v", err)
			}

			time.Sleep(1 * time.Second)
		}
	}()
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
