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
)

type LineWriter struct {
	callback func(line string)
	buffer   bytes.Buffer
}

func NewLineWriter(callback func(line string)) *LineWriter {
	return &LineWriter{
		callback: callback,
	}
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	for _, b := range p {
		if b == '\n' {
			w.callback(w.buffer.String())
			w.buffer.Reset()
		} else {
			w.buffer.WriteByte(b)
		}
	}
	return n, nil
}
