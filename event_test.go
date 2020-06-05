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

func TestEventfMetadataParam(t *testing.T) {
	metadata := map[string]string{
		"foo": "foo",
	}

	param := map[string]string{
		"bar": "bar",
	}

	e := Eventf(CriticalSeverity, nil, "foo: %v", param, metadata)
	expected := map[string]interface{}{
		"foo": "foo",
	}
	assert.EqualValues(t, expected, e.Metadata)
}

func TestEventfMetadataParamInterface(t *testing.T) {
	metadata := map[string]interface{}{
		"foo": 3,
	}

	e := Eventf(CriticalSeverity, nil, "foo", metadata)
	expected := map[string]interface{}{
		"foo": 3,
	}
	assert.EqualValues(t, expected, e.Metadata)
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
