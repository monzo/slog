package slog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventfFormatsParams(t *testing.T) {
	e := Eventf(CriticalSeverity, nil, "foo: %s", "bar")
	assert.Equal(t, "foo: bar", e.Message)
}

func TestEventfNilContext(t *testing.T) {
	e := Eventf(CriticalSeverity, nil, "foo: %s", "bar")
	if e.Context == nil {
		t.Error("background context should have been used automatically")
	}
}

func TestOriginalMessagePreserved(t *testing.T) {
	testCases := []struct {
		desc             string
		message          string
		params           []interface{}
		expectedMessage  string
		expectedOriginal string
	}{
		{
			desc:             "no formatting",
			message:          "foo",
			params:           []interface{}{},
			expectedMessage:  "foo",
			expectedOriginal: "foo",
		},
		{
			desc:             "no formatting with error",
			message:          "foo",
			params:           []interface{}{assert.AnError},
			expectedMessage:  "foo",
			expectedOriginal: "foo",
		},
		{
			desc:             "simple format string",
			message:          "foo: %s",
			params:           []interface{}{"bar"},
			expectedMessage:  "foo: bar",
			expectedOriginal: "foo: %s",
		},
		{
			desc:             "formatting with error",
			message:          "foo: %v",
			params:           []interface{}{assert.AnError},
			expectedMessage:  "foo: assert.AnError general error for testing",
			expectedOriginal: "foo: %v",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := Eventf(ErrorSeverity, nil, tC.message, tC.params...)
			assert.Equal(t, tC.expectedMessage, e.Message)
			assert.Equal(t, tC.expectedOriginal, e.OriginalMessage)
		})
	}
}

func TestEventMetadata(t *testing.T) {
	testCases := []struct {
		desc            string
		message         string
		params          []interface{}
		expected        map[string]interface{}
		expectedMessage string
		expectedError   error
	}{
		{
			desc:            "Message with no params",
			message:         "test",
			params:          nil,
			expected:        nil,
			expectedMessage: "test",
			expectedError:   nil,
		},
		{
			desc:            "Message with no metadata",
			message:         "test %d",
			params:          []interface{}{43},
			expected:        nil,
			expectedMessage: "test 43",
			expectedError:   nil,
		},
		{
			desc:    "Message with string metadata",
			message: "test",
			params: []interface{}{
				map[string]string{
					"foo": "bar",
				},
			},
			expected: map[string]interface{}{
				"foo": "bar",
			},
			expectedMessage: "test",
			expectedError:   nil,
		},
		{
			desc:    "Message with interface metadata",
			message: "test",
			params: []interface{}{
				map[string]interface{}{
					"foo": 42,
				},
			},
			expected: map[string]interface{}{
				"foo": 42,
			},
			expectedMessage: "test",
			expectedError:   nil,
		},
		{
			desc:    "map as format arg with metadata",
			message: "foo: %v",
			params: []interface{}{
				map[string]string{
					"bar": "bar",
				},
				map[string]string{
					"foo": "foo",
				},
			},
			expected: map[string]interface{}{
				"bar": "bar",
				"foo": "foo",
			},
			expectedMessage: "foo: map[bar:bar]",
			expectedError:   nil,
		},
		{
			desc:            "Message with special error case",
			message:         "test",
			params:          []interface{}{assert.AnError},
			expected:        map[string]interface{}(nil),
			expectedMessage: "test",
			expectedError:   assert.AnError,
		},
		{
			desc:    "Message with special error case and metadata",
			message: "test",
			params: []interface{}{assert.AnError, map[string]interface{}{
				"foo": "bar",
			}},
			expected: map[string]interface{}{
				"foo": "bar",
			},
			expectedMessage: "test",
			expectedError:   assert.AnError,
		},
		{
			desc:            "Message with interpolated error",
			message:         "eaten by a grue: %v",
			params:          []interface{}{assert.AnError},
			expected:        map[string]interface{}(nil),
			expectedMessage: "eaten by a grue: assert.AnError general error for testing",
			expectedError:   assert.AnError,
		},
		{
			desc:    "Message with error param and metadata",
			message: "eaten by a grue: %v",
			params: []interface{}{assert.AnError, map[string]interface{}{
				"foo": "bar",
			}},
			expected: map[string]interface{}{
				"foo": "bar",
			},
			expectedMessage: "eaten by a grue: assert.AnError general error for testing",
			expectedError:   assert.AnError,
		},
		{
			desc:            "Message with metadata nil explicitly",
			message:         "Foo %s",
			params:          []interface{}{"bar", nil, nil},
			expected:        nil,
			expectedMessage: "Foo bar",
			expectedError:   nil,
		},
		{
			desc:            "Invalid: too many format params",
			message:         "Foo %s %s",
			params:          []interface{}{"bar"},
			expected:        nil,
			expectedMessage: "Foo bar %!s(MISSING)",
			expectedError:   nil,
		},
		{
			desc:    "Invalid: too many format params with metadata",
			message: "Foo %s %s %s",
			params: []interface{}{"bar", map[string]interface{}{
				"meta": "data",
			}},
			expected: map[string]interface{}{
				"meta": "data",
			},
			expectedMessage: "Foo bar map[meta:data] %!s(MISSING)",
			expectedError:   nil,
		},
		{
			desc:            "Invalid: too many format params with error",
			message:         "Foo %s %s %s",
			params:          []interface{}{"bar", assert.AnError},
			expected:        map[string]interface{}(nil),
			expectedMessage: "Foo bar assert.AnError general error for testing %!s(MISSING)",
			expectedError:   assert.AnError,
		},
		{
			desc:    "Invalid: too many format params with error and metadata",
			message: "Foo %s %s %s %s",
			params: []interface{}{"bar", assert.AnError, map[string]interface{}{
				"meta": "data",
			}},
			expected: map[string]interface{}{
				"meta": "data",
			},
			expectedMessage: "Foo bar assert.AnError general error for testing map[meta:data] %!s(MISSING)",
			expectedError:   assert.AnError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := Eventf(ErrorSeverity, nil, tC.message, tC.params...)
			assert.EqualValues(t, tC.expected, e.Metadata)
			assert.Equal(t, tC.expectedMessage, e.Message)
			assert.Equal(t, tC.expectedError, e.Error)
		})
	}
}

func Test_InlineParamsTakePrecedenceOverContextParams(t *testing.T) {
	ctx := WithParams(context.Background(), map[string]string{
		"key1": "value_to_be_shadowed",
		"key2": "other_value",
	})

	e := Eventf(ErrorSeverity, ctx, "test message", map[string]string{
		"key1": "value",
	})

	assert.Equal(t, map[string]any{
		"key1": "value",
		"key2": "other_value",
	}, e.Metadata)
}

type testLogMetadataProvider map[string]string

func (p testLogMetadataProvider) LogMetadata() map[string]string {
	return p
}

func TestEventfLogMetadataProvider(t *testing.T) {
	param := testLogMetadataProvider{
		"foo": "bar",
	}

	e := Eventf(CriticalSeverity, nil, "foo: %v", param)
	expected := map[string]interface{}{
		"foo": "bar",
	}
	assert.EqualValues(t, expected, e.Metadata)
}

func TestSerializeDeserialize(t *testing.T) {
	event := Event{
		Context:         context.Background(),
		Id:              "test",
		Timestamp:       time.Now(),
		Severity:        ErrorSeverity,
		Message:         "foo",
		OriginalMessage: "foo",
		Metadata: map[string]interface{}{
			"string": "value",
			"number": float64(42),
		},
		Labels: map[string]string{
			"label": "foo",
		},
		Error: errors.New("an error"),
	}
	out, err := json.Marshal(&event)
	assert.NoError(t, err)

	var undo Event
	err = json.Unmarshal(out, &undo)
	assert.NoError(t, err)

	assert.Equal(t, event.Id, undo.Id)
	assert.Equal(t, event.Severity, undo.Severity)
	assert.Equal(t, event.Message, undo.Message)
	assert.Equal(t, event.Metadata, undo.Metadata)
	assert.Equal(t, event.Labels, undo.Labels)

	// Note: go error types will not serialize by default, so we do not expect
	// any data here.
	assert.Equal(t, map[string]interface{}{}, undo.Error)
}

func TestSerializeDeserializeError(t *testing.T) {
	type serializableError struct {
		Message string `json:"message"`
	}

	event := Event{
		Context:         context.Background(),
		Id:              "test",
		Timestamp:       time.Now(),
		Severity:        ErrorSeverity,
		Message:         "foo",
		OriginalMessage: "foo",
		Metadata: map[string]interface{}{
			"string": "value",
			"number": float64(42),
		},
		Labels: map[string]string{
			"label": "foo",
		},
		Error: serializableError{
			Message: "test",
		},
	}
	out, err := json.Marshal(&event)
	assert.NoError(t, err)

	var undo Event
	err = json.Unmarshal(out, &undo)
	assert.NoError(t, err)

	assert.Equal(t, event.Id, undo.Id)
	assert.Equal(t, event.Severity, undo.Severity)
	assert.Equal(t, event.Message, undo.Message)
	assert.Equal(t, event.Metadata, undo.Metadata)
	assert.Equal(t, event.Labels, undo.Labels)

	assert.Equal(t, map[string]interface{}{
		"message": "test",
	}, undo.Error)
}

func BenchmarkLogMetadataInterface(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Eventf(ErrorSeverity, nil, "foo", map[string]interface{}{
			"string": "foo",
			"number": 42,
		})
	}
}

func BenchmarkLogMetadataStrings(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Eventf(ErrorSeverity, nil, "foo", map[string]string{
			"string": "foo",
			"number": "42",
		})
	}
}

func BenchmarkLogMetadataInterpolated(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Eventf(ErrorSeverity, nil, "foo %s %d", "foo", 42)
	}
}
