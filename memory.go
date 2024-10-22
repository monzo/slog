package slog

import (
	"sync"
)

type InMemoryLogger struct {
	*sync.Mutex
	events EventSet
}

// NewInMemoryLogger creates a logger that will keep all log events in memory
// Call InMemoryLogger.Events to access all logged events
func NewInMemoryLogger() *InMemoryLogger {
	return &InMemoryLogger{
		Mutex: &sync.Mutex{},
	}
}

func (l *InMemoryLogger) Log(evs ...Event) {
	l.Lock()
	defer l.Unlock()
	l.events = append(l.events, evs...)
}

func (l *InMemoryLogger) Flush() error {
	return nil
}

func (l *InMemoryLogger) Events() EventSet {
	l.Lock()
	defer l.Unlock()
	output := make(EventSet, len(l.events))
	copy(output, l.events)
	return output
}
