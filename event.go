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
	Metadata map[string]string `json:"meta,omitempty"`
	// RawMetadata is metadata before it has been transformed to string form.
	// Metadata and RawMetadata will always contain the same keys.
	RawMetadata map[string]interface{} `json:"rawMeta,omitempty"`
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

	rawMetadata := map[string]interface{}(nil)
	metadata := map[string]string(nil)
	if len(params) > 0 {

		fmtOperands := countFmtOperands(msg)

		// Catch the special error case for Event(..., "Something happened", err).
		// We explicitly don't process Event(..., "Something happened: %v", err) this way, as
		// then intentions are not the same, and we don't want to take an opinion on whether to
		// remove the format argument in this case.
		if len(params) == 1 && fmtOperands == 0 {
			errParam, ok := params[0].(error)
			if ok {
				return createErrorEvent(ctx, sev, id, msg, errParam)
			}
		}

		// If we have been provided with more params than we have formatting arguments
		// then the last param should be a metadata map
		if len(params) > fmtOperands {
			metadataParam := params[len(params)-1]
			params = params[:len(params)-1]

			if metadataParam, ok := metadataParam.(map[string]string); ok {
				// Note: we merge the metadata here to avoid mutating the map
				metadata = mergeMetadata(metadata, metadataParam)
				// To maintain the invariant that Metadata contains the stringified version of RawMetadata
				// we copy all values in RawMetadata too. This isn't very efficient, but the invariant is
				// important for consumers of RawMetadata.
				rawMetadata = make(map[string]interface{}, len(metadata))
				for k, v := range metadata {
					rawMetadata[k] = v
				}
			}

			// Check for 'raw' metadata rather than strings, and convert to string form if found.
			if metadataParam, ok := metadataParam.(map[string]interface{}); ok {
				rawMetadata = mergeRawMetadata(rawMetadata, metadataParam)

				metadata = make(map[string]string, len(rawMetadata))
				for k, v := range rawMetadata {
					metadata[k] = fmt.Sprint(v)
				}
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

			metadata = mergeMetadata(metadata, param.LogMetadata())
		}

		if fmtOperands > 0 {
			msg = fmt.Sprintf(msg, params...)
		}
	}

	return Event{
		Context:     ctx,
		Id:          id.String(),
		Timestamp:   time.Now().UTC(),
		Severity:    sev,
		Message:     msg,
		Metadata:    metadata,
		RawMetadata: rawMetadata,
	}
}

func createErrorEvent(ctx context.Context, sev Severity, id *uuid.UUID, msg string, err error) Event {
	return Event{
		Context:   ctx,
		Id:        id.String(),
		Timestamp: time.Now().UTC(),
		Severity:  sev,
		Message:   msg,
		Metadata: map[string]string{
			"error": err.Error(),
		},
		RawMetadata: map[string]interface{}{
			"error": err,
		},
	}
}

// mergeMetadata merges the metadata but preserves existing entries
func mergeMetadata(current, new map[string]string) map[string]string {
	if len(new) == 0 {
		return current
	}

	if current == nil {
		current = map[string]string{}
	}

	for k, v := range new {
		if _, ok := current[k]; !ok {
			current[k] = v
		}
	}

	return current
}

// mergeRawMetadata merges the metadata but preserves existing entries
func mergeRawMetadata(current, new map[string]interface{}) map[string]interface{} {
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
