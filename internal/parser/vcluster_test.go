package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVCluster(t *testing.T) {
	input := `
	service my_service {
		dependency {
			name: my_dependency;
		}
		health_check {
			endpoint: "/health";
		}
		startup_sequence {
			command: "start.sh";
		}
	}
	`

	ast, err := ParseVCluster(input)
	assert.NoError(t, err)

	expected := &VClusterAST{
		Services: []VClusterDefinitionAST{
			{
				Name: "my_service",
				Dependencies: []VClusterDependency{
					{Name: "my_dependency"},
				},
				HealthChecks: VClusterHealthCheck{
					Endpoint: "/health",
				},
				StartupSeq: []VClusterStartupSequence{
					{Command: "start.sh"},
				},
			},
		},
	}

	assert.Equal(t, expected, ast)
}

func TestParseVCluster_Error(t *testing.T) {
	input := `
	service my_service {
		dependency {
			name: my_dependency
		}
		health_check {
			endpoint: "/health";
		}
		startup_sequence {
			command: "start.sh;
		}
	}
	`

	_, err := ParseVCluster(input)
	assert.Error(t, err)
}
