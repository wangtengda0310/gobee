package logger

import (
	"io"
	"sync"
)

type SilentWriter struct {
	mu     sync.Mutex
	writer io.Writer
	closed bool
}

func NewSilentWriter(w io.Writer) *SilentWriter {
	return &SilentWriter{writer: w}
}

func (w *SilentWriter) Write(p []byte) (int, error) {
	defer func() {
		a := recover()
		if a != nil {
			w.closed = true
		}
	}()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return len(p), nil
	}
	n, err := w.writer.Write(p)
	if err != nil {
		w.closed = true
		return len(p), nil
	}
	return n, err
}
