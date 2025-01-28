//go:build go1.21
// +build go1.21

package slog

import "maps"

// cloneStringMap wraps maps.Clone on Go 1.21 and newer.
func cloneStringMap(m map[string]string) map[string]string {
	return maps.Clone(m)
}
