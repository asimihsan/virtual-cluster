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
	"log"
	"os/exec"
)

type ManagedProcess struct {
	Name             string
	RunCommands      []string
	WorkingDirectory string
	Stop             chan struct{}
}

func runProcessAndStoreOutput(
	process *ManagedProcess,
	db *sql.DB,
	verbose bool,
) {
	outputCallback := func(line string) {
		if verbose {
			fmt.Printf("%s: %s", process.Name, line)
		}
		_, err := db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stdout', ?)", process.Name, line)
		if err != nil {
			log.Fatal(err)
		}
	}

	errorCallback := func(line string) {
		if verbose {
			fmt.Printf("%s: %s", process.Name, line)
		}
		_, err := db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stderr', ?)", process.Name, line)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, cmdStr := range process.RunCommands {
		fmt.Println("Running command:", cmdStr)
		err := runShellCommand(
			process.Stop,
			cmdStr,
			process.WorkingDirectory,
			outputCallback,
			errorCallback,
		)
		if err != nil {
			fmt.Println("Error occurred while running command:", cmdStr, "Error:", err)
			break
		}
	}
}

type OutputCallback func(string)
type ErrorCallback func(string)

func runShellCommand(
	stop chan struct{},
	command string,
	workingDirectory string,
	outputCallback OutputCallback,
	errorCallback ErrorCallback,
) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = workingDirectory

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	outputScanner := bufio.NewScanner(stdout)
	go readStream(outputScanner, outputCallback, "stdout")

	errorScanner := bufio.NewScanner(stderr)
	go readStream(errorScanner, errorCallback, "stderr")

	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for the command to finish.
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Use select to listen to the Stop channel and pre-emptively stop the process
	// if a signal is received.
	select {
	case <-stop:
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill process: ", err)
		}
		log.Println("process killed as stop signal received")
		return nil
	case err := <-done:
		return err
	}
}

func readStream(
	scanner *bufio.Scanner,
	callback func(string),
	streamType string,
) {
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		callback(line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", streamType, err)
	}
}
