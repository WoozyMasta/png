//go:build amd64 && !purego

package simd

import (
	"os"

	"golang.org/x/sys/cpu"
)

// pureGo disables all assembly kernels at runtime when PNG_PUREGO=1 is set.
// The purego build tag does the same at compile time.
var pureGo = os.Getenv("PNG_PUREGO") != ""

// hasSSE2 gates the SSE2 kernels (AddInto, SubAndSumAbs).
// SSE2 is baseline on amd64; the only reason to skip it is PNG_PUREGO.
var hasSSE2 = !pureGo

// hasSSSE3 gates the PSHUFB kernels (ExpandRGBToRGBA, CompactRGBAToRGB).
var hasSSSE3 = !pureGo && cpu.X86.HasSSSE3
