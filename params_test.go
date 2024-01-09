package slog

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Params(t *testing.T) {
	t.Run("Params is empty by default", func(t *testing.T) {
		emptyParams := Params(context.Background())
		require.NotNil(t, emptyParams)
		require.Empty(t, emptyParams)
	})

	t.Run("Params does not panic on nil context", func(t *testing.T) {
		assert.Empty(t, Params(nil))
	})

	t.Run("Params returns a map that can be mutated", func(t *testing.T) {
		ctx := WithParams(context.Background(), map[string]string{"key": "value"})

		heldParams := Params(ctx)
		heldParams["key"] = "new_value"

		// The original context should not be affected by our mutation
		assert.Equal(t, map[string]string{"key": "value"}, Params(ctx))
	})
}

func Test_WithParams(t *testing.T) {
	t.Run("single use of WithParams", func(t *testing.T) {
		// WithParams returns a new context containing the params
		ctx1 := context.Background()
		ctx2 := WithParams(ctx1, map[string]string{"key": "value"})
		assert.Empty(t, ctx1)
		assert.Equal(t, map[string]string{"key": "value"}, Params(ctx2))
	})

	t.Run("multiple calls to WithParams", func(t *testing.T) {
		ctx1 := context.Background()
		ctx2 := WithParams(ctx1, map[string]string{"key1": "value1"})
		ctx3 := WithParams(ctx2, map[string]string{"key2": "value2"})
		assert.Empty(t, ctx1)
		assert.Equal(t, map[string]string{"key1": "value1"}, Params(ctx2))
		assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, Params(ctx3))
	})

	t.Run("later values take precedence", func(t *testing.T) {
		ctx1 := context.Background()
		ctx2 := WithParams(ctx1, map[string]string{"key": "value1"})
		ctx3 := WithParams(ctx2, map[string]string{"key": "value2"})
		assert.Empty(t, ctx1)
		assert.Equal(t, map[string]string{"key": "value1"}, Params(ctx2))
		assert.Equal(t, map[string]string{"key": "value2"}, Params(ctx3))
	})
}

func Benchmark_WithParams(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_ = WithParams(ctx, map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		})
	}
}

func Benchmark_WithParam(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		ctx := WithParam(ctx, "k1", "v1")
		ctx = WithParam(ctx, "k2", "v2")
		_ = WithParam(ctx, "k3", "v3")
	}
}

func Benchmark_Params_Uncached(b *testing.B) {
	contexts := make([]context.Context, b.N)
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		ctx = WithParam(ctx, "k1", "v1")
		ctx = WithParam(ctx, "k2", "v2")
		ctx = WithParam(ctx, "k3", "v3")
		contexts[i] = ctx
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Params(contexts[i])
	}
}

func Benchmark_Params_Cached(b *testing.B) {
	ctx := context.Background()
	ctx = WithParam(ctx, "k1", "v1")
	ctx = WithParam(ctx, "k2", "v2")
	ctx = WithParam(ctx, "k3", "v3")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Params(ctx)
	}
}
