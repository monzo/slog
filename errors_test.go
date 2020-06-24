package slog

import (
	"encoding/json"
	"testing"

	"github.com/monzo/terrors"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestWireErrorSimple(t *testing.T) {
	wireErr := newSimpleWireError(assert.AnError)
	assert.Equal(t, WireErrorTypeSimple, wireErr.Typ)
	assert.Equal(t, "assert.AnError general error for testing", wireErr.Data)

	serialized, err := json.Marshal(wireErr)
	require.NoError(t, err)

	wireErrDe, err := NewWireErrorFromWrapper(serialized)
	require.NoError(t, err)

	err, problem := wireErrDe.Decode()
	require.NoError(t, problem)

	assert.Equal(t, "assert.AnError general error for testing", err.Error())
}

func TestWireErrorTerror(t *testing.T) {
	startTerr := terrors.BadRequest("code", "unique-message", map[string]string{
		"key": "value",
	})
	wireErr, err := newWireError(startTerr)
	require.NoError(t, err)

	assert.Equal(t, WireErrorTypeTerror, wireErr.Typ)
	assert.Contains(t, wireErr.Data, "unique-message")

	serialized, err := json.Marshal(wireErr)
	require.NoError(t, err)

	wireErrDe, err := NewWireErrorFromWrapper(serialized)
	require.NoError(t, err)

	err, problem := wireErrDe.Decode()
	require.NoError(t, problem)

	assert.IsType(t, &terrors.Error{}, err)
	terr := err.(*terrors.Error)

	assert.Equal(t, startTerr, terr)
}
