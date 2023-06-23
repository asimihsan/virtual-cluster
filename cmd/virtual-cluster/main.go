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
	"os"
	"path/filepath"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	"github.com/urfave/cli/v2"
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
						},
						Action: func(c *cli.Context) error {
							configQueue := []string{}

							if c.String("config-dir") != "" {
								configQueue = append(configQueue, c.String("config-dir"))
							}

							if c.String("config-file") != "" {
								configQueue = append(configQueue, c.String("config-file"))
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

									fmt.Printf("AST for '%s':\n%+v\n", config, ast)
								}
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
