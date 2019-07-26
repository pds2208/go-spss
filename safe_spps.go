package spss

//Wraps around SafeCSVWriter and makes it thread safe.

import (
	"encoding/csv"
	"sync"
)

type SafeSPSSWriter struct {
	*csv.Writer
	m sync.Mutex
}

func NewSafeSPSSWriter(original *csv.Writer) *SafeSPSSWriter {
	return &SafeSPSSWriter{
		Writer: original,
	}
}

//Override write
func (w *SafeSPSSWriter) Write(row []string) error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.Writer.Write(row)
}

//Override flush
func (w *SafeSPSSWriter) Flush() {
	w.m.Lock()
	w.Writer.Flush()
	w.m.Unlock()
}
