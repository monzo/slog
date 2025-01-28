//go:build !go1.21
// +build !go1.21

package slog

// cloneStringMap provides a polyfill for maps.Clone on Go 1.20 and older.
//
// This implementation is not as efficient as the built-in implementation (which uses
// runtime.clone internally).
func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return m
	}

	res := make(map[string]string, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
