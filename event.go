package slog

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

type Severity int

const (
	TimeFormat                = "2006-01-02 15:04:05-0700 (MST)"
	TraceSeverity    Severity = 1
	DebugSeverity    Severity = 2
	InfoSeverity     Severity = 3
	WarnSeverity     Severity = 4
	ErrorSeverity    Severity = 5
	CriticalSeverity Severity = 6
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

type logMetadataProvider interface {
	LogMetadata() map[string]string
}

// An Event is a discrete logging event
type Event struct {
	Context   context.Context `json:"-"`
	Id        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Severity  Severity        `json:"severity"`
	Message   string          `json:"message"`
	// Metadata are structured key-value pairs which describe the event.
	Metadata map[string]interface{} `json:"meta,omitempty"`
	// Labels, like Metadata, are key-value pairs which describe the event. Unlike Metadata, these are intended to be
	// indexed.
	Labels map[string]string `json:"labels,omitempty"`
}

func (e Event) String() string {
	return fmt.Sprintf("[%s] %s %s (metadata=%v labels=%v id=%s)", e.Timestamp.Format(TimeFormat), e.Severity.String(),
		e.Message, e.Metadata, e.Labels, e.Id)
}

// Eventf constructs an event from the given message string and formatting operands. Optionally, event metadata
// (map[string]interface{}, or map[string]string) can be provided as a final argument.
func Eventf(sev Severity, ctx context.Context, msg string, params ...interface{}) Event {
	if ctx == nil {
		ctx = context.Background()
	}

	id, err := uuid.NewV4()
	if err != nil {
		return Event{}
	}

	metadata := map[string]interface{}(nil)
	if len(params) > 0 {

		fmtOperands := countFmtOperands(msg)

		// If we have been provided with more params than we have formatting arguments
		// then the last param should be a metadata map
		if len(params) > fmtOperands {
			metadataParam := params[len(params)-1]
			params = params[:len(params)-1]

			// This is deprecated, but continue to support a map of strings.
			if metadataParam, ok := metadataParam.(map[string]string); ok {
				// Note: we merge the metadata here to avoid mutating the map
				metadata = mergeMetadata(metadata, stringMapToInterfaceMap(metadataParam))
			}

			// Check for 'raw' metadata rather than strings.
			if metadataParam, ok := metadataParam.(map[string]interface{}); ok {
				metadata = mergeMetadata(metadata, metadataParam)
			}
		}

		// If any of the provided params can be "upgraded" to a logMetadataProvider i.e.
		// they themselves have a LogMetadata method that returns a map[string]string
		// then we merge these params with the metadata.
		for _, param := range params {
			param, ok := param.(logMetadataProvider)
			if !ok {
				continue
			}
			metadata = mergeMetadata(metadata, stringMapToInterfaceMap(param.LogMetadata()))
		}

		if fmtOperands > 0 {
			msg = fmt.Sprintf(msg, params...)
		}
	}

	return Event{
		Context:   ctx,
		Id:        id.String(),
		Timestamp: time.Now().UTC(),
		Severity:  sev,
		Message:   msg,
		Metadata:  metadata,
	}
}

func stringMapToInterfaceMap(m map[string]string) map[string]interface{} {
	shim := make(map[string]interface{}, len(m))
	for k, v := range m {
		shim[k] = v
	}
	return shim
}

// mergeMetadata merges the metadata but preserves existing entries
func mergeMetadata(current, new map[string]interface{}) map[string]interface{} {
	if len(new) == 0 {
		return current
	}

	if current == nil {
		current = map[string]interface{}{}
	}

	for k, v := range new {
		if _, ok := current[k]; !ok {
			current[k] = v
		}
	}

	return current
}
