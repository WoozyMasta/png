//go:build amd64 && !purego

package simd

import "golang.org/x/sys/cpu"

// hasSSSE3 reports whether the CPU supports SSSE3 (required by PSHUFB).
// SSE2 kernels (AddInto, SubAndSumAbs) do not need it;
// shuffle kernels fall back to scalar Go when it is absent.
var hasSSSE3 = cpu.X86.HasSSSE3
