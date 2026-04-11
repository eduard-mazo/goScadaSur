// pkg/api/joblog.go
package api

import (
	"fmt"
	"sync"
	"time"
)

// LogEntry es una entrada individual del log de un job.
type LogEntry struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"` // info | warn | error
	Msg   string    `json:"msg"`
}

// BufferLogger implementa xmlcreator.Logger acumulando entradas en memoria
// para devolverlas en la respuesta HTTP del job.
type BufferLogger struct {
	mu      sync.Mutex
	Entries []LogEntry
}

func (b *BufferLogger) add(level, format string, args ...any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Entries = append(b.Entries, LogEntry{
		Time:  time.Now(),
		Level: level,
		Msg:   fmt.Sprintf(format, args...),
	})
}

func (b *BufferLogger) Infof(f string, a ...any)  { b.add("info", f, a...) }
func (b *BufferLogger) Warnf(f string, a ...any)  { b.add("warn", f, a...) }
func (b *BufferLogger) Errorf(f string, a ...any) { b.add("error", f, a...) }

// Counts devuelve totales por nivel para resúmenes rápidos en la UI.
func (b *BufferLogger) Counts() (info, warn, errs int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, e := range b.Entries {
		switch e.Level {
		case "info":
			info++
		case "warn":
			warn++
		case "error":
			errs++
		}
	}
	return
}
