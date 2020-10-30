package slog

import "context"

type contextKey string

// we use a key here of an unexported type so that no code outside of this package
// can access these params other than via the exported accessor funcs
var slogParamsContextKey contextKey = "slog-params"

// WithParams packs a map of slog params into a context.Context, overwriting
// any previously packed params of the same key. These fields would also be
// overwritten by any params of the same name in the actual slog.Info/Error/etc call.
func WithParams(ctx context.Context, params map[string]string) context.Context {
	p := ParamsFromContext(ctx)
	p = mergeParams(p, params)

	return context.WithValue(ctx, slogParamsContextKey, p)
}

// ParamsFromContext returns any slog params stored within the context, if no params
// are found it will return a non-nil empty map
func ParamsFromContext(ctx context.Context) map[string]string {
	m, ok := ctx.Value(slogParamsContextKey).(map[string]string)
	if !ok {
		// return a non-nil map so we can still assign to it
		// without checking for nil
		return make(map[string]string)
	}
	return m
}

// mergeParams will merge map b into map a, overwriting any keys in a that overlap
// with b
func mergeParams(a, b map[string]string) map[string]string {
	// take a copy of map A so we don't mutate the original map
	copyA := make(map[string]string, len(a))
	for k, v := range a {
		copyA[k] = v
	}
	// write all keys from map b into our copy of a
	for k, v := range b {
		copyA[k] = v
	}

	return copyA
}
