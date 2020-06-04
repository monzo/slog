package slog

import (
	"errors"
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
	assert.EqualValues(t, metadata, e.Metadata)
	assert.EqualValues(t, map[string]interface{}{
		"foo": "foo",
	}, e.RawMetadata)
}

func TestEventRawMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"error": assert.AnError,
	}
	e := Eventf(CriticalSeverity, nil, "msg", metadata)

	assert.Equal(t, metadata, e.RawMetadata)
	assert.Equal(t, map[string]string{
		"error": assert.AnError.Error(),
	}, e.Metadata)
}

func TestEventSpecialErrorCase(t *testing.T) {
	e := Eventf(ErrorSeverity, nil, "Eaten by a grue", assert.AnError)

	assert.Equal(t, "Eaten by a grue", e.Message)
	assert.Equal(t, map[string]string{
		"error": assert.AnError.Error(),
	}, e.Metadata)

	assert.Equal(t, map[string]interface{}{
		"error": assert.AnError,
	}, e.RawMetadata)
}

func TestEventSpecialErrorCaseWithFormat(t *testing.T) {
	e := Eventf(ErrorSeverity, nil, "game over: %v", errors.New("eaten by a grue"))

	assert.Equal(t, "game over: eaten by a grue", e.Message)
	assert.Equal(t, map[string]string(nil), e.Metadata)
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
	assert.EqualValues(t, param, e.Metadata)
}
