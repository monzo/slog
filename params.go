package slog

import (
	"context"
	"sync"
)

type contextKeyParams struct{}

type paramsStack struct {
	stack    []map[string]string
	stackMtx sync.Mutex
}

func WithParams(ctx context.Context, params map[string]string) context.Context {
	oldStack := paramsStackFromContext(ctx)
	newStack := paramsStack{stack: append(oldStack.stack, params)}
	return context.WithValue(ctx, contextKeyParams{}, &newStack)
}

func WithParam(ctx context.Context, key, value string) context.Context {
	return WithParams(ctx, map[string]string{key: value})
}

func Params(ctx context.Context) map[string]string {
	params := paramsStackFromContext(ctx)
	params.stackMtx.Lock()
	defer params.stackMtx.Unlock()

	switch len(params.stack) {
	case 0:
		// No parameters stored in this context
		return map[string]string{}

	case 1:
		// WithParams called exactly once, so we can return the map directly
		return params.stack[0]

	default:
		// WithParams called multiple times, so we need to merge the maps. If a key is
		// present in multiple maps, the value from the map that was most recently pushed
		// to the stack takes precedence.
		mergedParams := map[string]string{}
		for _, paramMap := range params.stack {
			for k, v := range paramMap {
				mergedParams[k] = v
			}
		}

		// Write the merged parameter map back to the stack so that future calls to
		// Params can return it directly
		params.stack = []map[string]string{mergedParams}
		return mergedParams
	}
}

func paramsStackFromContext(ctx context.Context) *paramsStack {
	if ctx == nil {
		return nil
	}
	stack, ok := ctx.Value(contextKeyParams{}).(*paramsStack)
	if !ok {
		Critical(ctx, "internal error: slog.paramsStackFromContext: context value is not a *paramsStack")
		return &paramsStack{}
	}
	return stack
}
