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
	"log"
	"sync"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	_ "github.com/mattn/go-sqlite3"
)

type Manager struct {
	dbPath    string
	db        *sql.DB
	processes []*ManagedProcess
	wg        sync.WaitGroup
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
			process_name TEXT,
			output_type TEXT,
			content TEXT
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Manager{
		dbPath: dbPath,
		db:     db,
	}, nil
}

func (m *Manager) Close() error {
	m.StopAllProcesses()
	return m.db.Close()
}

// takes the combined slice of VClusterServiceDefinitionAST and VClusterDependencyDefinitionAST and
// performs a topological sort.
func topologicalSort(definitions []interface{}) ([]interface{}, error) {
	// Perform topological sort on the combined slice

	return definitions, nil
}

func (m *Manager) StartServicesAndDependencies(asts []*parser.VClusterAST) error {
	// Combine and check for duplicate names
	// Perform topological sort
	// Start services and dependencies one at a time

	for _, ast := range asts {
		for _, service := range ast.Services {
			fmt.Println("Starting service:", service.Name)
			process := &ManagedProcess{
				Name:        service.Name,
				RunCommands: service.RunCommands,
				Stop:        make(chan struct{}, 1),
			}
			m.processes = append(m.processes, process)
			m.wg.Add(1)

			go runProcessAndStoreOutput(process, m.db)
			fmt.Println("Started service:", service.Name)
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
