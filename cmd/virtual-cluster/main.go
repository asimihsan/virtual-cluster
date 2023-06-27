/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package main

import (
	"fmt"
	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/asimihsan/virtual-cluster/internal/substrate"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	app := &cli.App{
		Name:  "vcluster",
		Usage: "Virtual cluster",
		Commands: []*cli.Command{
			{
				Name:  "substrate",
				Usage: "Substrate",
				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "Start substrate",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "config-dir",
								Aliases: []string{"d"},
								Usage:   "directory containing config files",
							},
							&cli.StringFlag{
								Name:    "config-file",
								Aliases: []string{"f"},
								Usage:   "config file",
							},
							&cli.StringFlag{
								Name:  "db-path",
								Usage: "filepath to database",
							},
							&cli.StringSliceFlag{
								Name:    "working-dir",
								Aliases: []string{"w"},
								Usage:   "working directory for a service, specified as key=value, e.g. service=~/service",
							},
						},
						Action: func(c *cli.Context) error {
							dbPath := c.String("db-path")
							if dbPath == "" {
								fmt.Fprintf(os.Stderr, "db-path is required\n")
								return nil
							}
							if fi, err := os.Stat(dbPath); err == nil && fi.IsDir() {
								fmt.Fprintf(os.Stderr, "db-path '%s' is a directory\n", dbPath)
								return nil
							}

							configQueue := []string{}
							var asts []*parser.VClusterAST

							if c.String("config-dir") != "" {
								configQueue = append(configQueue, c.String("config-dir"))
							}

							if c.String("config-file") != "" {
								configQueue = append(configQueue, c.String("config-file"))
							}

							workingDirsInput := c.StringSlice("working-dir")
							workingDirs := make(map[string]string)
							for _, wd := range workingDirsInput {
								kv := strings.SplitN(wd, "=", 2)
								if len(kv) != 2 {
									fmt.Fprintf(os.Stderr, "failed to parse working-dir '%s'\n", wd)
									return nil
								}

								serviceName := kv[0]
								serviceDir := strings.Trim(kv[1], "\"")
								workingDirs[serviceName] = serviceDir
							}

							for {
								if len(configQueue) == 0 {
									break
								}

								config := configQueue[0]
								configQueue = configQueue[1:]

								if _, err := os.Stat(config); os.IsNotExist(err) {
									fmt.Fprintf(os.Stderr, "directory or file '%s' does not exist\n", config)
									continue
								}

								if fi, err := os.Stat(config); err == nil && fi.IsDir() {
									err := filepath.Walk(config, func(path string, info os.FileInfo, err error) error {
										if err != nil {
											return err
										}

										if info.IsDir() {
											return nil
										}

										configQueue = append(configQueue, path)
										return nil
									})

									if err != nil {
										fmt.Fprintf(os.Stderr, "failed to walk directory '%s': %s\n", config, err)
									}
								} else {
									contents, err := os.ReadFile(config)
									if err != nil {
										fmt.Fprintf(os.Stderr, "failed to read '%s': %s\n", config, err)
										continue
									}

									ast, err := parser.ParseVCluster(string(contents))
									if err != nil {
										fmt.Fprintf(os.Stderr, "failed to parse '%s'\n", config)
										if err, ok := err.(parser.SyntaxErrors); ok {
											for _, e := range err.Errors {
												fmt.Fprintln(os.Stderr, e)
											}
										} else {
											fmt.Fprintf(os.Stderr, "%s\n", err)
										}

										continue
									}

									asts = append(asts, ast)
									fmt.Printf("AST for '%s':\n%+v\n", config, ast)
								}
							}

							// Start substrate with asts.
							manager, err := substrate.NewManager(dbPath)
							if err != nil {
								fmt.Fprintf(os.Stderr, "failed to create substrate manager: %s\n", err)
								return nil
							}
							defer func(manager *substrate.Manager) {
								err := manager.Close()
								if err != nil {
									fmt.Fprintf(os.Stderr, "failed to close substrate manager: %s\n", err)
								}
							}(manager)

							for service, wd := range workingDirs {
								err := manager.AddWorkingDirectoryUpward(service, wd, false /*verbose*/)
								if err != nil {
									fmt.Fprintf(os.Stderr, "failed to add working directory '%s' for service '%s': %s\n", wd, service, err)
									return nil
								}
							}

							err = manager.StartServicesAndDependencies(asts)
							if err != nil {
								fmt.Fprintf(os.Stderr, "failed to start services and dependencies: %s\n", err)
								return nil
							}

							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
