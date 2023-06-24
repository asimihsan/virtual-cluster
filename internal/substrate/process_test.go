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
	"bytes"
	"testing"
)

func TestRunShellCommand(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		expectedOutput string
		expectedError  error
	}{
		{
			name:           "simple command",
			command:        "echo hello world",
			expectedOutput: "hello world\n",
			expectedError:  nil,
		},
		{
			name:           "multiple lines",
			command:        "echo hello world && echo goodbye world",
			expectedOutput: "hello world\ngoodbye world\n",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			var errOutput bytes.Buffer

			err := runShellCommand(tt.command, func(line string) {
				output.WriteString(line)
			}, func(line string) {
				errOutput.WriteString(line)
			})

			if err != nil && tt.expectedError == nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedError != nil {
				t.Fatalf("expected error: %v, but got nil", tt.expectedError)
			}
			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Fatalf("unexpected error: got %v, want %v", err, tt.expectedError)
			}

			if output.String() != tt.expectedOutput {
				t.Fatalf("unexpected output: got %q, want %q", output.String(), tt.expectedOutput)
			}
			if errOutput.String() != "" {
				t.Fatalf("unexpected error output: got %q, want \"\"", errOutput.String())
			}
		})
	}
}
