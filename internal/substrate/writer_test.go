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

func TestLineWriter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"one char", "a", ""},
		{"one line without line ending", "hello world", ""},
		{"one line with line ending", "hello world\n", "hello world\n"},
		{"two lines without line ending on second line", "hello\nworld", "hello\n"},
		{"two lines with line ending on second line", "hello\nworld\n", "hello\nworld\n"},
		{"multiple lines", "hello\nworld\nhow\nare\nyou\n", "hello\nworld\nhow\nare\nyou\n"},
		{"last line without line ending", "hello\nworld\nhow\nare\nyou", "hello\nworld\nhow\nare\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			writer := NewLineWriter(func(line string) {
				output.WriteString(line)
				output.WriteByte('\n')
			})

			n, err := writer.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n != len(tt.input) {
				t.Fatalf("unexpected length: got %d, want %d", n, len(tt.input))
			}

			if output.String() != tt.want {
				t.Fatalf("unexpected output: got %q, want %q", output.String(), tt.want)
			}
		})
	}
}
