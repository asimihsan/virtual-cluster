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
