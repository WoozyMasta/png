// Command asmgen generates the amd64 SSE2 assembly kernels used by package
// internal/simd. Run via `make generate` (see the repo Makefile), which invokes:
//
//	go run . -out ../filters_amd64.s -stubs ../filters_stub_amd64.go -pkg simd
//
// The kernels are plain SSE2 (16-byte vectors),
// which is part of the amd64 baseline, so no runtime CPU feature detection is required.
// Each kernel processes only whole 16-byte blocks (len &^ 15);
// the Go wrappers in simd_amd64.go handle the scalar tail.
package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	// Only build these symbols when the purego tag is absent. Combined with the
	// _amd64 file-name suffix on the outputs this yields `amd64 && !purego`.
	ConstraintExpr("!purego")

	addIntoBlocks()
	subSumBlocks()

	Generate()
}

// addIntoBlocks: dst[i] += src[i] for the first len(dst)&^15 bytes.
func addIntoBlocks() {
	TEXT("addIntoBlocks", NOSPLIT, "func(dst, src []byte)")
	Doc("addIntoBlocks sets dst[i] += src[i] for the first len(dst)&^15 bytes (whole 16-byte blocks).")

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())
	n := Load(Param("dst").Len(), GP64())
	ANDQ(I32(-16), n) // n = len(dst) &^ 15

	i := GP64()
	XORQ(i, i)

	Label("loop")
	CMPQ(i, n)
	JGE(LabelRef("done"))

	a := XMM()
	b := XMM()
	MOVOU(Mem{Base: dst, Index: i, Scale: 1}, a)
	MOVOU(Mem{Base: src, Index: i, Scale: 1}, b)
	PADDB(b, a) // a += b (packed bytes)
	MOVOU(a, Mem{Base: dst, Index: i, Scale: 1})
	ADDQ(Imm(16), i)
	JMP(LabelRef("loop"))

	Label("done")
	RET()
}

// subSumBlocks: dst[i] = a[i] - b[i] and returns the sum of abs8(dst[i])
// over the first len(a)&^15 bytes. abs8(d) = d<128 ? d : 256-d,
// computed as min(out, -out) over unsigned bytes (the standard PMINUB/PSADBW trick).
func subSumBlocks() {
	TEXT("subSumBlocks", NOSPLIT, "func(dst, a, b []byte) uint64")
	Doc("subSumBlocks sets dst[i]=a[i]-b[i] and returns sum of abs8 over the first len(a)&^15 bytes.")

	dst := Load(Param("dst").Base(), GP64())
	aptr := Load(Param("a").Base(), GP64())
	bptr := Load(Param("b").Base(), GP64())
	n := Load(Param("a").Len(), GP64())
	ANDQ(I32(-16), n)

	acc := XMM()
	zero := XMM()
	PXOR(acc, acc)
	PXOR(zero, zero)

	i := GP64()
	XORQ(i, i)

	Label("loop")
	CMPQ(i, n)
	JGE(LabelRef("reduce"))

	va := XMM()
	vb := XMM()
	vneg := XMM()
	MOVOU(Mem{Base: aptr, Index: i, Scale: 1}, va)
	MOVOU(Mem{Base: bptr, Index: i, Scale: 1}, vb)
	PSUBB(vb, va)                                 // va = a - b = out
	MOVOU(va, Mem{Base: dst, Index: i, Scale: 1}) // store out
	PXOR(vneg, vneg)
	PSUBB(va, vneg)  // vneg = -out
	PMINUB(vneg, va) // va = min(out, -out) = abs8 bytes
	PSADBW(zero, va) // va = sum|abs| per 8-byte lane
	PADDQ(va, acc)   // acc += partial sums
	ADDQ(Imm(16), i)
	JMP(LabelRef("loop"))

	Label("reduce")
	// Horizontally add the two 64-bit lanes of acc.
	res := GP64()
	MOVQ(acc, res) // low lane
	hi := XMM()
	PSHUFD(Imm(0xEE), acc, hi) // hi low qword = acc high qword
	hireg := GP64()
	MOVQ(hi, hireg)
	ADDQ(hireg, res)
	Store(res, ReturnIndex(0))
	RET()
}
