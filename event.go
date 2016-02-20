package slog

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
)

type Severity int

const (
	TimeFormat             = "2006-01-02 15:04:05-0700 (MST)"
	TraceSeverity Severity = iota
	DebugSeverity
	InfoSeverity
	WarnSeverity
	ErrorSeverity
	CriticalSeverity
)

func (s Severity) String() string {
	switch s {
	case CriticalSeverity:
		return "CRITICAL"
	case ErrorSeverity:
		return "ERROR"
	case WarnSeverity:
		return "WARN"
	case InfoSeverity:
		return "INFO"
	case DebugSeverity:
		return "DEBUG"
	default:
		return "TRACE"
	}
}

// An Event is a discrete logging event
type Event struct {
	Context   context.Context
	Timestamp time.Time `json:"timestamp"`
	Severity  Severity  `json:"severity"`
	Message   string    `json:"message"`
	// Metadata are structured key-value pairs which describe the event.
	Metadata map[string]string `json:"meta,omitempty"`
	// Labels, like Metadata, are key-value pairs which describe the event. Unlike Metadata, these are intended to be
	// indexed.
	Labels map[string]string `json:"labels,omitempty"`
}

func (e Event) String() string {
	return fmt.Sprintf("[%s] %s %s (metadata=%v labels=%v)", e.Timestamp.Format(TimeFormat), e.Severity.String(),
		e.Message, e.Metadata, e.Labels)
}

// Eventf constructs an event from the given message string and formatting operands. Optionally, event metadata
// (map[string]string) can be provided as a final argument.
func Eventf(sev Severity, ctx context.Context, msg string, params ...interface{}) Event {
	if ctx == nil {
		ctx = context.Background()
	}
	metadata := map[string]string(nil)
	if len(params) > 0 {
		fmtOperands := countFmtOperands(msg)
		if len(params) > fmtOperands {
			param := params[len(params)-1]
			if param == nil {
				params = params[:len(params)-1]
			} else if metadata_, ok := param.(map[string]string); ok {
				metadata = metadata_
				params = params[:len(params)-1]
			}
		}
		if fmtOperands > 0 {
			msg = fmt.Sprintf(msg, params...)
		}
	}
	return Event{
		Context:   ctx,
		Timestamp: time.Now(),
		Severity:  sev,
		Message:   msg,
		Metadata:  metadata}
}
