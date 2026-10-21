package startflix

import (
	"context"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
)

func TestConstantVIDSRCModulePreservesI64BitPatterns(t *testing.T) {
	for _, value := range []uint64{
		0, 63, 64, 8191, 8192, 65530, 65536, 1 << 32, math.MaxUint32,
		math.MaxInt64, 1 << 63, math.MaxUint64,
	} {
		t.Run(strconv.FormatUint(value, 10), func(t *testing.T) {
			ctx := context.Background()
			r := wazero.NewRuntime(ctx)
			t.Cleanup(func() { require.NoError(t, r.Close(ctx)) })
			mod, err := r.Instantiate(ctx, constantVIDSRCModule(value, value))
			require.NoError(t, err)

			ptr, err := mod.ExportedFunction("alloc").Call(ctx, 0)
			require.NoError(t, err)
			require.Equal(t, []uint64{value}, ptr, "alloc must preserve the selected i64 bits")
			out, err := mod.ExportedFunction("decrypt").Call(ctx, 0, 0)
			require.NoError(t, err)
			require.Equal(t, []uint64{value}, out, "decrypt must preserve the selected i64 bits")
		})
	}
}
