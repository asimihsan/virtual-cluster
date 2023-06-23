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
