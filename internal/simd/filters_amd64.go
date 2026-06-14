//go:build amd64 && !purego

// Package simd provide simd
package simd

// AddInto sets dst[i] += src[i] for i in range(len(dst)).
// len(src) must be >= len(dst).
//
// Whole 16-byte blocks are handled by the SSE2 kernel;
// the remainder is done in Go.
func AddInto(dst, src []byte) {
	n := len(dst)
	blocks := n &^ 15
	if blocks > 0 {
		addIntoBlocks(dst, src)
	}
	for i := blocks; i < n; i++ {
		dst[i] += src[i]
	}
}

// SubAndSumAbs sets dst[i] = a[i] - b[i] for i in range(len(a))
// and returns the sum of abs8(dst[i]),
// where abs8(d) interprets d as a signed int8 magnitude
// (d < 128 ? d : 256-d). len(dst) and len(b) must be >= len(a).
func SubAndSumAbs(dst, a, b []byte) int {
	n := len(a)
	blocks := n &^ 15
	var sum uint64
	if blocks > 0 {
		sum = subSumBlocks(dst, a, b)
	}
	for i := blocks; i < n; i++ {
		d := a[i] - b[i]
		dst[i] = d
		if d < 128 {
			sum += uint64(d)
		} else {
			sum += uint64(256 - int(d))
		}
	}
	return int(sum)
}
