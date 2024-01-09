package slog

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Params(t *testing.T) {
	{
		// Params is empty by default
		ctx := context.Background()
		require.Empty(t, Params(ctx))
	}
	{
		// WithParams returns a new context containing the params
		ctx := WithParams(context.Background(), map[string]string{"foo": "bar"})
		require.Equal(t, map[string]string{"foo": "bar"}, Params(ctx))
		require.Equal(t, map[string]string{"foo": "bar"}, Params(ctx))
	}
	{
		// Multiple calls to WithParams stack
		ctx := WithParams(context.Background(), map[string]string{"foo": "bar"})
		ctx = WithParams(ctx, map[string]string{"baz": "qux"})
		require.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, Params(ctx))
		require.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, Params(ctx))
	}
	{
		// Later values take precedence
		ctx := WithParams(context.Background(), map[string]string{"foo": "bar"})
		ctx = WithParams(ctx, map[string]string{"foo": "baz"})
		require.Equal(t, map[string]string{"foo": "baz"}, Params(ctx))
	}
}
