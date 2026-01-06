package slog

import (
	"context"
	"testing"

	"github.com/benbjohnson/immutable"
)

// WithParams returns a copy of the parent context containing the given log parameters.
// Any log events generated using the returned context will include these parameters
// as metadata.
//
// For example:
//
//	 ctx := slog.WithParams(ctx, map[string]string{
//	   "foo_id": fooID,
//	   "bar_id": barID,
//	 })
//
//	slog.Info(ctx, "Linking foo to bar")  // includes foo_id and bar_id parameters
//
// If the parent context already contains parameters set by a previous call to WithParams,
// the new parameters will be merged with the existing set, with newer values taking
// precedence over older ones.
//
// We copy the contents of the map into an internal structure here, so while it
// is safe to modify the map after being passed in, any changes won't be visible
// to successive slog calls.
func WithParams(parent context.Context, input map[string]string) context.Context {
	var p params
	if node := paramNodeFromContext(parent); node != nil {
		p = node.mergedParams
	} else {
		p = immutable.NewMap[string, string](nil)
	}

	for k, v := range input {
		p = p.Set(k, v)
	}
	return context.WithValue(parent, contextKeyParamNode{}, &paramNode{
		mergedParams: p,
	})
}

// WithParam is shorthand for calling WithParams with a single key-value pair.
func WithParam(ctx context.Context, key, value string) context.Context {
	return WithParams(ctx, map[string]string{key: value})
}

// Params returns all parameters stored in the given context using WithParams. This
// function is intended to be used by libraries _other_ than slog that want access to the
// set of parameters (e.g. `monzo/terrors` functions).
//
// The return value is guaranteed to be non-nil and can be safely mutated by the caller.
func Params(ctx context.Context) map[string]string {
	paramNode := paramNodeFromContext(ctx)
	if paramNode == nil {
		return map[string]string{}
	}
	// As above, we return a copy to allow safe mutation
	return paramsToBuiltinMap(paramNode.mergedParams)
}

func paramsToBuiltinMap(p params) map[string]string {
	m := map[string]string{}

	it := p.Iterator()
	for !it.Done() {
		k, v, _ := it.Next()
		m[k] = v
	}
	return m
}

type params = *immutable.Map[string, string]

type paramNode struct {
	mergedParams params
}

type contextKeyParamNode struct{}

func paramNodeFromContext(ctx context.Context) *paramNode {
	if ctx == nil {
		return nil
	}

	stackAny := ctx.Value(contextKeyParamNode{})
	if stackAny == nil {
		return nil
	}

	stack, ok := stackAny.(*paramNode)
	if !ok {
		// This should never happen, and would typically indicate a bug in this library.
		// If it happens in a test case we panic to ensure the failure isn't silently
		// occurring in unit tests, otherwise we just log loudly.
		errMsg := "internal error: slog.paramNodeFromContext: context value is not a *paramNode"
		if testing.Testing() {
			panic(errMsg)
		} else {
			Critical(context.Background(), errMsg)
			return nil
		}
	}

	return stack
}
