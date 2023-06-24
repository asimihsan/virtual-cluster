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
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os/exec"
	"sync"
)

type ManagedProcess struct {
	Name        string
	RunCommands []string
	Stop        chan struct{}
}

func runProcessAndStoreOutput(process *ManagedProcess, db *sql.DB) {
	outputCallback := func(line string) {
		fmt.Println("Writing line to DB:", line)
		_, err := db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stdout', ?)", process.Name, line)
		if err != nil {
			log.Fatal(err)
		}
	}

	errorCallback := func(line string) {
		fmt.Println("Writing error line to DB:", line)
		_, err := db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stderr', ?)", process.Name, line)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, cmdStr := range process.RunCommands {
		fmt.Println("Running command:", cmdStr)
		err := runShellCommand(cmdStr, outputCallback, errorCallback)
		if err != nil {
			fmt.Println("Error occurred while running command:", cmdStr, "Error:", err)
			break
		}
	}
}

type OutputCallback func(string)
type ErrorCallback func(string)

func runShellCommand(command string, outputCallback OutputCallback, errorCallback ErrorCallback) error {
	cmd := exec.Command("bash", "-c", command)

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	err := cmd.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 1)

	go func() {
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		err := cmd.Wait()
		if err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	go func() {
		readStream(stdoutReader, outputCallback, "stdout")
		wg.Done()
	}()

	go func() {
		readStream(stderrReader, errorCallback, "stderr")
		wg.Done()
	}()

	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func readStream(reader io.Reader, callback func(string), streamType string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		callback(line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", streamType, err)
	}
}
