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

func TestEventMetadata(t *testing.T) {
	testCases := []struct {
		desc     string
		message  string
		params   []interface{}
		expected map[string]interface{}
	}{
		{
			desc:     "Message with no params",
			message:  "test",
			params:   nil,
			expected: nil,
		},
		{
			desc:     "Message with no metadata",
			message:  "test %d",
			params:   []interface{}{43},
			expected: nil,
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
				"foo": "foo",
			},
		},
		{
			desc:    "Message with special error case",
			message: "test",
			params:  []interface{}{assert.AnError},
			expected: map[string]interface{}{
				"error": assert.AnError,
			},
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
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := Eventf(ErrorSeverity, nil, tC.message, tC.params...)
			assert.EqualValues(t, tC.expected, e.Metadata)
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
