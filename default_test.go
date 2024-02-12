package slog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLogger(t *testing.T) {
	logger := &testLogLogger{}
	oldLogger := DefaultLogger()
	SetDefaultLogger(logger)
	defer SetDefaultLogger(oldLogger)

	Trace(context.Background(), "Important trace message", "foo")
	Debug(context.Background(), "Important debug message", "foo")
	Info(context.Background(), "Important info message", "foo")
	Warn(context.Background(), "Important warn message", "foo")
	Error(context.Background(), "Important error message", "foo")
	Critical(context.Background(), "Important critical message", "foo")

	require.Equal(t, 6, len(logger.events))
	assert.Equal(t, TraceSeverity, logger.events[0].Severity)
	assert.Equal(t, DebugSeverity, logger.events[1].Severity)
	assert.Equal(t, InfoSeverity, logger.events[2].Severity)
	assert.Equal(t, WarnSeverity, logger.events[3].Severity)
	assert.Equal(t, ErrorSeverity, logger.events[4].Severity)
	assert.Equal(t, CriticalSeverity, logger.events[5].Severity)
}

func TestDefaultLoggerWithLeveledLogger(t *testing.T) {
	logger := &testLogLeveledLogger{t: t}
	oldLogger := DefaultLogger()
	SetDefaultLogger(logger)
	defer SetDefaultLogger(oldLogger)

	Trace(context.Background(), "Important trace message", "foo")
	Debug(context.Background(), "Important debug message", "foo")
	Info(context.Background(), "Important info message", "foo")
	Warn(context.Background(), "Important warn message", "foo")
	Error(context.Background(), "Important error message", "foo")
	Critical(context.Background(), "Important critical message", "foo")

	require.Equal(t, 6, len(logger.items))

	assert.Equal(t, TraceSeverity, logger.items[0].Severity)
	assert.Equal(t, "Important trace message", logger.items[0].OriginalMessage)

	assert.Equal(t, DebugSeverity, logger.items[1].Severity)
	assert.Equal(t, "Important debug message", logger.items[1].OriginalMessage)

	assert.Equal(t, InfoSeverity, logger.items[2].Severity)
	assert.Equal(t, "Important info message", logger.items[2].OriginalMessage)

	assert.Equal(t, WarnSeverity, logger.items[3].Severity)
	assert.Equal(t, "Important warn message", logger.items[3].OriginalMessage)

	assert.Equal(t, ErrorSeverity, logger.items[4].Severity)
	assert.Equal(t, "Important error message", logger.items[4].OriginalMessage)

	assert.Equal(t, CriticalSeverity, logger.items[5].Severity)
	assert.Equal(t, "Important critical message", logger.items[5].OriginalMessage)
}

func TestDefaultLoggerWithFromErrorLogger(t *testing.T) {
	logger := &testLogFromErrorLogger{t: t}
	oldLogger := DefaultLogger()
	SetDefaultLogger(logger)
	defer SetDefaultLogger(oldLogger)

	FromError(context.Background(), "This error ends up as debug", context.Canceled, "foo")
	FromError(context.Background(), "This error ends up as error", context.DeadlineExceeded, "foo")

	require.Equal(t, 2, len(logger.items))

	assert.Equal(t, DebugSeverity, logger.items[0].Severity)
	assert.Equal(t, "This error ends up as debug", logger.items[0].OriginalMessage)
	assert.Equal(t, []interface{}{context.Canceled, "foo"}, logger.items[0].Params)

	assert.Equal(t, ErrorSeverity, logger.items[1].Severity)
	assert.Equal(t, "This error ends up as error", logger.items[1].OriginalMessage)
	assert.Equal(t, []interface{}{context.DeadlineExceeded, "foo"}, logger.items[1].Params)
}

func TestNilDefaultLogger(t *testing.T) {
	oldLogger := DefaultLogger()
	SetDefaultLogger(nil)
	defer SetDefaultLogger(oldLogger)

	require.Nil(t, DefaultLogger())

	Trace(context.Background(), "Important trace message", "foo")
	Debug(context.Background(), "Important debug message", "foo")
	Info(context.Background(), "Important info message", "foo")
	Warn(context.Background(), "Important warn message", "foo")
	Error(context.Background(), "Important error message", "foo")
	Critical(context.Background(), "Important critical message", "foo")
	FromError(context.Background(), "Important from error message", context.Canceled, "foo")
}

// testLogLeveledLogger implements the Logger interface
type testLogLogger struct {
	events []Event
}

func (l *testLogLogger) Log(evs ...Event) {
	l.events = append(l.events, evs...)
}

func (l *testLogLogger) Flush() error {
	return nil
}

type logItem struct {
	Severity        Severity
	OriginalMessage string
	Params          []interface{}
}

// testLogLeveledLogger implements the Logger and LeveledLogger interfaces
type testLogLeveledLogger struct {
	t     *testing.T
	items []logItem
}

func (l *testLogLeveledLogger) Critical(ctx context.Context, msg string, params ...interface{}) {
	l.items = append(l.items, logItem{Severity: CriticalSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogLeveledLogger) Error(ctx context.Context, msg string, params ...interface{}) {
	l.items = append(l.items, logItem{Severity: ErrorSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogLeveledLogger) Warn(ctx context.Context, msg string, params ...interface{}) {
	l.items = append(l.items, logItem{Severity: WarnSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogLeveledLogger) Info(ctx context.Context, msg string, params ...interface{}) {
	l.items = append(l.items, logItem{Severity: InfoSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogLeveledLogger) Debug(ctx context.Context, msg string, params ...interface{}) {
	l.items = append(l.items, logItem{Severity: DebugSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogLeveledLogger) Trace(ctx context.Context, msg string, params ...interface{}) {
	l.items = append(l.items, logItem{Severity: TraceSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogLeveledLogger) Log(evs ...Event) {
	l.t.Fail() // We expect this method to not be called
}

func (l *testLogLeveledLogger) Flush() error {
	return nil
}

// testLogFromErrorLogger implements the Logger and FromErrorLogger interfaces
type testLogFromErrorLogger struct {
	t     *testing.T
	items []logItem
}

func (l *testLogFromErrorLogger) FromError(ctx context.Context, msg string, err error, params ...interface{}) {
	if errors.Is(err, context.Canceled) {
		l.items = append(l.items, logItem{Severity: DebugSeverity, OriginalMessage: msg, Params: params})
		return
	}
	l.items = append(l.items, logItem{Severity: ErrorSeverity, OriginalMessage: msg, Params: params})
}

func (l *testLogFromErrorLogger) Log(evs ...Event) {
	l.t.Fail() // We expect this method to not be called
}

func (l *testLogFromErrorLogger) Flush() error {
	return nil
}
