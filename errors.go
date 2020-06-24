package slog

import (
	"encoding/json"
	"fmt"

	"github.com/monzo/terrors"
	"github.com/pkg/errors"
)

type WireErrorType string

var WireErrorTypeTerror WireErrorType = "terror"
var WireErrorTypeSimple WireErrorType = "simple"

// WireError is an error which can be (json) [de]serialized.
// It lets us capture both standard Go errors and Monzo's terrors (which
// can already be serialized) in a single type.
// Whilst the internal error data could be transported as bytes with any encoding,
// we intentionally encode as JSON and serialize as a string. This means that even
// if the error is not deserialized on the other side of the wire, we still maintain
// human readable output if the string is printed.
type WireError struct {
	Typ  WireErrorType `json:"type"`
	Data string        `json:"data"`
}

func newWireError(inputErr error) (*WireError, error) {
	switch t := inputErr.(type) {
	case *terrors.Error:
		data, err := json.Marshal(t)
		if err != nil {
			return nil, err
		}
		return &WireError{
			Typ:  WireErrorTypeTerror,
			Data: string(data),
		}, nil
	}
	return newSimpleWireError(inputErr), nil
}

func newSimpleWireError(err error) *WireError {
	return &WireError{
		Typ:  WireErrorTypeSimple,
		Data: fmt.Sprintf("%v", err),
	}
}

func NewWireErrorFromWrapper(wrapper []byte) (*WireError, error) {
	wire := WireError{}
	err := json.Unmarshal(wrapper, &wire)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal to WireError")
	}
	return &wire, nil
}

func (w *WireError) Decode() (error, error) {
	switch w.Typ {
	case WireErrorTypeSimple:
		return errors.New(w.Data), nil
	case WireErrorTypeTerror:
		terr := &terrors.Error{}
		err := json.Unmarshal([]byte(w.Data), terr)
		if err != nil {
			return nil, err
		}
		return terr, nil
	default:
		return nil, errors.New("unknown wire error type")
	}
}
