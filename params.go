package slog

import (
	"context"
	"sync"
	"testing"
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
// It is not safe to modify the supplied map after passing it to WithParams.
func WithParams(parent context.Context, params map[string]string) context.Context {
	return context.WithValue(parent, contextKeyParamNode{}, &paramNode{
		Parent:      parent,
		ChildParams: params,
	})
}

// WithParam is shorthand for calling WithParams with a single key-value pair.
func WithParam(ctx context.Context, key, value string) context.Context {
	return WithParams(ctx, params{key: value})
}

// Params returns all parameters stored in the given context using WithParams. This
// function is intended to be used by libraries _other_ than slog that want access to the
// set of parameters (i.e. `monzo/terrors` functions).
//
// The return value is guaranteed to be non-nil and can be safely mutated by the caller.
func Params(ctx context.Context) map[string]string {
	paramNode := paramNodeFromContext(ctx)
	if paramNode == nil {
		return map[string]string{}
	}
	return paramNode.params()
}

type params map[string]string

type paramNode struct {
	// Parent and ChildParams are the original values passed to slog.WithParams. The
	// complete set of parameters for this node are determined by collecting any
	// parameters already contained in Parent and then merging that with ChildParams.
	//
	// NOTE: this collection happens lazily when the params are queried, at which point
	// we also cache the result in mergedParams to avoid repeating this work.
	Parent      context.Context
	ChildParams params

	mergedParams    params
	mergedParamsMtx sync.RWMutex
}

func (n *paramNode) params() params {
	n.mergedParamsMtx.Lock()
	defer n.mergedParamsMtx.Unlock()

	// If we've already flattened the params, we can return those directly
	if n.mergedParams != nil {
		// NOTE: we return a _copy_ of the cached map here to allow the caller to safely
		// mutate it without impacting other callers (and potentially causing panics if
		// the map is mutated concurrently). This trades off a small amount of performance
		// and memory usage for safety.
		return cloneStringMap(n.mergedParams)
	}

	// NOTE: we could propagate length hints down the parent chain in order to pass a
	// more accurate capacity hint here, but the minimum size of a map already takes up
	// to 8 K/V pairs without needing to allocate more buckets so in practice it doesn't
	// matter much anyway.
	result := make(params, len(n.ChildParams))
	n.collectAllParamsAssumingReadLock(result)

	// Cache the result for future requests
	n.mergedParams = result

	return cloneStringMap(result) // As above, we return a copy to allow safe mutation
}

func (n *paramNode) collectAllParams(dst params) {
	n.mergedParamsMtx.RLock()
	defer n.mergedParamsMtx.RUnlock()
	n.collectAllParamsAssumingReadLock(dst)
}

func (n *paramNode) collectAllParamsAssumingReadLock(res params) {
	// If we've already cached the flattened params for this node, we can accumulate from
	// those directly, avoiding potentially needing to traverse the parent chain to
	// collect all params
	if n.mergedParams != nil {
		for k, v := range n.mergedParams {
			res[k] = v
		}
		return
	}

	// Collect params from the parent node first (if it exists)
	if parentNode := paramNodeFromContext(n.Parent); parentNode != nil {
		parentNode.collectAllParams(res)
	}

	// Then merge the child params, overwriting any existing bindings for a given
	// parameter key so that more recent calls to WithParams take precedence.
	for k, v := range n.ChildParams {
		res[k] = v
	}

	// NOTE: here we intentionally _don't_ cache the result in this paramNode because
	// doing so would require us to clone the map, making our overall memory usage O(n^2)
	// in the length of the paramNode chain. The trade-off is that we may redundantly
	// re-traverse the parent chain collecting parameters in some use-cases (i.e. if
	// there is a very long parent chain with many leaf nodes at the bottom, and we query
	// parameters for each of the leaf nodes separately).
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
