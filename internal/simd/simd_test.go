package simd

import (
	"bytes"
	"math/rand"
	"testing"
)

// testLengths includes sub-block (<16), exact multiples, and just-over
// multiples so the vector body and the scalar tail are both exercised.
var testLengths = []int{0, 1, 7, 15, 16, 17, 31, 32, 33, 48, 63, 64, 255, 256, 257, 1000}

func randBytes(n int, r *rand.Rand) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.Intn(256))
	}
	return b
}

// refAbs8 is the PNG spec definition, kept independent of the kernels
// so the test does not validate the implementation against itself.
func refAbs8(d byte) int {
	if d < 128 {
		return int(d)
	}
	return 256 - int(d)
}

func TestAddInto(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for _, n := range testLengths {
		dst := randBytes(n, r)
		src := randBytes(n, r)
		got := append([]byte(nil), dst...)
		want := append([]byte(nil), dst...)
		AddInto(got, src)
		for i := 0; i < n; i++ {
			want[i] += src[i]
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("AddInto n=%d:\n got %x\nwant %x", n, got, want)
		}
	}
}

func TestSubAndSumAbs(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for _, n := range testLengths {
		a := randBytes(n, r)
		b := randBytes(n, r)
		gotDst := make([]byte, n)
		wantDst := make([]byte, n)
		wantSum := 0
		for i := 0; i < n; i++ {
			d := a[i] - b[i]
			wantDst[i] = d
			wantSum += refAbs8(d)
		}
		gotSum := SubAndSumAbs(gotDst, a, b)
		if gotSum != wantSum || !bytes.Equal(gotDst, wantDst) {
			t.Fatalf("SubAndSumAbs n=%d: sum got %d want %d\n dst got %x\n    want %x",
				n, gotSum, wantSum, gotDst, wantDst)
		}
	}
}
