/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVCluster_SingleServiceNoDependencies(t *testing.T) {
	input := `
    service my_service {
        health_check {
            endpoint: "/health";
        }
    }
    `

	ast, err := ParseVCluster(input)
	assert.NoError(t, err)

	expected := &VClusterAST{
		Services: []VClusterServiceDefinitionAST{
			{
				Name: "my_service",
				HealthChecks: HealthCheck{
					Endpoint: "/health",
				},
			},
		},
	}

	assert.Equal(t, expected, ast)
}

func TestParseVCluster_ServiceWithOneRunCommand(t *testing.T) {
	input := `
    service my_service {
        run_commands: ["make run"];
    }
    `

	ast, err := ParseVCluster(input)
	assert.NoError(t, err)

	expected := &VClusterAST{
		Services: []VClusterServiceDefinitionAST{
			{
				Name: "my_service",
				RunCommands: []string{
					"make run",
				},
			},
		},
	}

	assert.Equal(t, expected, ast)
}

func TestParseVCluster_ServiceWithTwoRunCommands(t *testing.T) {
	input := `
    service my_service {
        run_commands: ["make build", "make run"];
    }
    `

	ast, err := ParseVCluster(input)
	assert.NoError(t, err)

	expected := &VClusterAST{
		Services: []VClusterServiceDefinitionAST{
			{
				Name: "my_service",
				RunCommands: []string{
					"make build",
					"make run",
				},
			},
		},
	}

	assert.Equal(t, expected, ast)
}

func TestParseVCluster_ServiceWithTwoRunCommandsAnotherFormat(t *testing.T) {
	input := `
    service my_service {
        run_commands: [
			"make build",
			"make run",
		];
    }
    `

	ast, err := ParseVCluster(input)
	assert.NoError(t, err)

	expected := &VClusterAST{
		Services: []VClusterServiceDefinitionAST{
			{
				Name: "my_service",
				RunCommands: []string{
					"make build",
					"make run",
				},
			},
		},
	}

	assert.Equal(t, expected, ast)
}

func TestParseVCluster_ServiceWithNoRunCommands(t *testing.T) {
	input := `
    service my_service {
    }
    `

	_, err := ParseVCluster(input)
	assert.Error(t, err)
}
