package simd

import "testing"

// rowLen is the data length of one 640px RGBA scanline
// (without the filter-type byte):
// the size filter() and the decode ftUp path operate on per row.
const rowLen = 640 * 4

func makeRow(seed byte) []byte {
	b := make([]byte, rowLen)
	for i := range b {
		b[i] = byte(i)*31 + seed
	}
	return b
}

func BenchmarkAddInto(b *testing.B) {
	dst := makeRow(1)
	src := makeRow(2)
	b.SetBytes(rowLen)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddInto(dst, src)
	}
}

func BenchmarkSubAndSumAbs(b *testing.B) {
	a := makeRow(1)
	c := makeRow(2)
	dst := make([]byte, rowLen)
	b.SetBytes(rowLen)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SubAndSumAbs(dst, a, c)
	}
}
