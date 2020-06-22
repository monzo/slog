package slog

import (
	"testing"

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
	}{
		{
			desc:            "Message with no params",
			message:         "test",
			params:          nil,
			expected:        nil,
			expectedMessage: "test",
		},
		{
			desc:            "Message with no metadata",
			message:         "test %d",
			params:          []interface{}{43},
			expected:        nil,
			expectedMessage: "test 43",
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
		},
		{
			desc:    "Message with special error case",
			message: "test",
			params:  []interface{}{assert.AnError},
			expected: map[string]interface{}{
				"error": assert.AnError,
			},
			expectedMessage: "test",
		},
		{
			desc:    "Message with special error case and metadata",
			message: "test",
			params: []interface{}{assert.AnError, map[string]interface{}{
				"foo": "bar",
			}},
			expected: map[string]interface{}{
				"error": assert.AnError,
				"foo":   "bar",
			},
			expectedMessage: "test",
		},
		{
			desc:    "Message with interpolated error",
			message: "eaten by a grue: %v",
			params:  []interface{}{assert.AnError},
			expected: map[string]interface{}{
				"error": assert.AnError,
			},
			expectedMessage: "eaten by a grue: assert.AnError general error for testing",
		},
		{
			desc:    "Message with error param and metadata",
			message: "eaten by a grue: %v",
			params: []interface{}{assert.AnError, map[string]interface{}{
				"foo": "bar",
			}},
			expected: map[string]interface{}{
				"foo":   "bar",
				"error": assert.AnError,
			},
			expectedMessage: "eaten by a grue: assert.AnError general error for testing",
		},
		{
			desc:            "Message with metadata nil explicitly",
			message:         "Foo %s",
			params:          []interface{}{"bar", nil, nil},
			expected:        nil,
			expectedMessage: "Foo bar",
		},
		{
			desc:            "Invalid: too many format params",
			message:         "Foo %s %s",
			params:          []interface{}{"bar"},
			expected:        nil,
			expectedMessage: "Foo bar %!s(MISSING)",
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
		},
		{
			desc:    "Invalid: too many format params with error",
			message: "Foo %s %s %s",
			params:  []interface{}{"bar", assert.AnError},
			expected: map[string]interface{}{
				"error": assert.AnError,
			},
			expectedMessage: "Foo bar assert.AnError general error for testing %!s(MISSING)",
		},
		{
			desc:    "Invalid: too many format params with error and metadata",
			message: "Foo %s %s %s %s",
			params: []interface{}{"bar", assert.AnError, map[string]interface{}{
				"meta": "data",
			}},
			expected: map[string]interface{}{
				"error": assert.AnError,
				"meta":  "data",
			},
			expectedMessage: "Foo bar assert.AnError general error for testing map[meta:data] %!s(MISSING)",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := Eventf(ErrorSeverity, nil, tC.message, tC.params...)
			assert.EqualValues(t, tC.expected, e.Metadata)
			assert.Equal(t, tC.expectedMessage, e.Message)
		})
	}
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
