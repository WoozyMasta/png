// Command asmgen generates the amd64 SSE2 assembly kernels used by package
// internal/simd. Run via `make generate` (see the repo Makefile), which invokes:
//
//	go run . -out ../filters_amd64.s -stubs ../filters_stub_amd64.go -pkg simd
//
// The add/sum kernels are plain SSE2 (part of the amd64 baseline).
// Shuffle kernels use PSHUFB (SSSE3);
// the Go wrappers gate them on a runtime CPU feature check (golang.org/x/sys/cpu)
// and fall back to scalar Go otherwise.
// The Go wrappers also handle the scalar tail left by each kernel.
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
	expandRGBToRGBABlocks()
	compactRGBAToRGBBlocks()

	Generate()
}

// expandRGBToRGBABlocks converts RGB pixels to RGBA (alpha = 0xFF),
// 4 pixels at a time via PSHUFB (SSSE3).
// It loads 16 src bytes per iteration but consumes only 12,
// so it stops while a 16-byte load is in bounds (i+16 <= len(src))
// and returns the number of src bytes consumed;
// Go wrapper finishes the tail.
func expandRGBToRGBABlocks() {
	TEXT("expandRGBToRGBABlocks", NOSPLIT, "func(dst, src []byte) int")
	Doc("expandRGBToRGBABlocks converts whole groups of 4 RGB pixels to RGBA (alpha 0xFF) and returns src bytes consumed.")

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())
	n := Load(Param("src").Len(), GP64())

	shuf := XMM()
	MOVOU(expandShuffle(), shuf)
	alpha := XMM()
	MOVOU(expandAlpha(), alpha)

	i := GP64() // src position
	XORQ(i, i)
	j := GP64() // dst position
	XORQ(j, j)

	Label("loop")
	end := GP64()
	LEAQ(Mem{Base: i, Disp: 16}, end) // end = i + 16
	CMPQ(end, n)
	JG(LabelRef("done")) // stop if a 16-byte load would over-read

	x := XMM()
	MOVOU(Mem{Base: src, Index: i, Scale: 1}, x)
	PSHUFB(shuf, x) // gather R,G,B into lanes, zero the alpha lane
	POR(alpha, x)   // set alpha lane to 0xFF
	MOVOU(x, Mem{Base: dst, Index: j, Scale: 1})
	ADDQ(Imm(12), i)
	ADDQ(Imm(16), j)
	JMP(LabelRef("loop"))

	Label("done")
	Store(i, ReturnIndex(0))
	RET()
}

// compactRGBAToRGBBlocks converts RGBA pixels to RGB (drops alpha),
// 4 pixels at a time via PSHUFB.
// It stores 16 dst bytes per iteration but only 12 are valid,
// so it stops while a 16-byte store is in bounds (j+16 <= len(dst))
// and returns the number of dst bytes written;
// Go wrapper finishes the tail.
func compactRGBAToRGBBlocks() {
	TEXT("compactRGBAToRGBBlocks", NOSPLIT, "func(dst, src []byte) int")
	Doc("compactRGBAToRGBBlocks converts whole groups of 4 RGBA pixels to RGB and returns dst bytes written.")

	dst := Load(Param("dst").Base(), GP64())
	src := Load(Param("src").Base(), GP64())
	n := Load(Param("dst").Len(), GP64())

	shuf := XMM()
	MOVOU(compactShuffle(), shuf)

	i := GP64() // src position
	XORQ(i, i)
	j := GP64() // dst position
	XORQ(j, j)

	Label("loop")
	end := GP64()
	LEAQ(Mem{Base: j, Disp: 16}, end) // end = j + 16
	CMPQ(end, n)
	JG(LabelRef("done")) // stop if a 16-byte store would over-write

	x := XMM()
	MOVOU(Mem{Base: src, Index: i, Scale: 1}, x)
	PSHUFB(shuf, x) // pack R,G,B of 4 pixels into the low 12 bytes
	MOVOU(x, Mem{Base: dst, Index: j, Scale: 1})
	ADDQ(Imm(16), i)
	ADDQ(Imm(12), j)
	JMP(LabelRef("loop"))

	Label("done")
	Store(j, ReturnIndex(0))
	RET()
}

// expandShuffle gathers, for 4 RGB pixels (bytes 0..11 of the 16-byte load),
// R,G,B into each output pixel's first 3 bytes; 0x80 zeroes the alpha byte.
func expandShuffle() Mem {
	g := GLOBL("expandShuffleMask", RODATA|NOPTR)
	DATA(0, String("\x00\x01\x02\x80\x03\x04\x05\x80\x06\x07\x08\x80\x09\x0a\x0b\x80"))
	return g
}

// expandAlpha is 0xFF in every alpha byte position, 0 elsewhere.
func expandAlpha() Mem {
	g := GLOBL("expandAlphaMask", RODATA|NOPTR)
	DATA(0, String("\x00\x00\x00\xff\x00\x00\x00\xff\x00\x00\x00\xff\x00\x00\x00\xff"))
	return g
}

// compactShuffle packs R,G,B of 4 RGBA pixels into the low 12 bytes;
// top 4 bytes are don't-care (0x80 -> zero).
func compactShuffle() Mem {
	g := GLOBL("compactShuffleMask", RODATA|NOPTR)
	DATA(0, String("\x00\x01\x02\x04\x05\x06\x08\x09\x0a\x0c\x0d\x0e\x80\x80\x80\x80"))
	return g
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
