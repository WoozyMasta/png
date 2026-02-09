// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package png

// intSize is either 32 or 64.
const intSize = 32 << (^uint(0) >> 63)

func abs(x int) int {
	// m := -1 if x < 0. m := 0 otherwise.
	m := x >> (intSize - 1)

	// In two's complement representation, the negative number
	// of any number (except the smallest one) can be computed
	// by flipping all the bits and add 1. This is faster than
	// code with a branch.
	// See Hacker's Delight, section 2-4.
	return (x ^ m) - m
}

// paeth implements the Paeth filter function, as per the PNG specification.
func paeth(a, b, c uint8) uint8 {
	// This is an optimized version of the sample code in the PNG spec.
	// For example, the sample code starts with:
	//	p := int(a) + int(b) - int(c)
	//	pa := abs(p - int(a))
	// but the optimized form uses fewer arithmetic operations:
	//	pa := int(b) - int(c)
	//	pa = abs(pa)
	pc := int(c)
	pa := int(b) - pc
	pb := int(a) - pc
	pc = abs(pa + pb)
	pa = abs(pa)
	pb = abs(pb)
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	}
	return c
}

// filterPaeth applies the Paeth filter to the cdat slice.
// cdat is the current row's data, pdat is the previous row's data.
func filterPaeth(cdat, pdat []byte, bytesPerPixel int) {
	switch {
	case bytesPerPixel == 1:
		filterPaeth1(cdat, pdat)
	case bytesPerPixel == 3 && len(cdat)%3 == 0:
		filterPaeth3(cdat, pdat)
	case bytesPerPixel == 4 && len(cdat)%4 == 0:
		filterPaeth4(cdat, pdat)
	default:
		filterPaethGeneric(cdat, pdat, bytesPerPixel)
	}
}

// filterPaeth1 for 1 byte per pixel
func filterPaeth1(cdat, pdat []byte) {
	var a, c int
	for j := 0; j < len(cdat); j++ {
		b := int(pdat[j])
		pa := b - c
		pb := a - c
		pc := abs(pa + pb)
		pa = abs(pa)
		pb = abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a = b
			} else {
				a = c
			}
		}
		a += int(cdat[j])
		a &= 0xff
		cdat[j] = uint8(a)
		c = b
	}
}

// filterPaeth3 for 3 bytes per pixel
func filterPaeth3(cdat, pdat []byte) {
	var a0, a1, a2, c0, c1, c2 int
	for j := 0; j < len(cdat); j += 3 {
		b0 := int(pdat[j])
		b1 := int(pdat[j+1])
		b2 := int(pdat[j+2])
		// byte 0
		pa := b0 - c0
		pb := a0 - c0
		pc := abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a0 = b0
			} else {
				a0 = c0
			}
		}
		a0 += int(cdat[j])
		a0 &= 0xff
		cdat[j] = uint8(a0)
		c0 = b0
		// byte 1
		pa = b1 - c1
		pb = a1 - c1
		pc = abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a1 = b1
			} else {
				a1 = c1
			}
		}
		a1 += int(cdat[j+1])
		a1 &= 0xff
		cdat[j+1] = uint8(a1)
		c1 = b1
		// byte 2
		pa = b2 - c2
		pb = a2 - c2
		pc = abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a2 = b2
			} else {
				a2 = c2
			}
		}
		a2 += int(cdat[j+2])
		a2 &= 0xff
		cdat[j+2] = uint8(a2)
		c2 = b2
	}
}

// filterPaeth4 for 4 bytes per pixel
func filterPaeth4(cdat, pdat []byte) {
	var a0, a1, a2, a3, c0, c1, c2, c3 int
	for j := 0; j < len(cdat); j += 4 {
		b0 := int(pdat[j])
		b1 := int(pdat[j+1])
		b2 := int(pdat[j+2])
		b3 := int(pdat[j+3])
		// byte 0
		pa := b0 - c0
		pb := a0 - c0
		pc := abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a0 = b0
			} else {
				a0 = c0
			}
		}
		a0 += int(cdat[j])
		a0 &= 0xff
		cdat[j] = uint8(a0)
		c0 = b0
		// byte 1
		pa = b1 - c1
		pb = a1 - c1
		pc = abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a1 = b1
			} else {
				a1 = c1
			}
		}
		a1 += int(cdat[j+1])
		a1 &= 0xff
		cdat[j+1] = uint8(a1)
		c1 = b1
		// byte 2
		pa = b2 - c2
		pb = a2 - c2
		pc = abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a2 = b2
			} else {
				a2 = c2
			}
		}
		a2 += int(cdat[j+2])
		a2 &= 0xff
		cdat[j+2] = uint8(a2)
		c2 = b2
		// byte 3
		pa = b3 - c3
		pb = a3 - c3
		pc = abs(pa + pb)
		pa, pb = abs(pa), abs(pb)
		if pa > pb || pa > pc {
			if pb <= pc {
				a3 = b3
			} else {
				a3 = c3
			}
		}
		a3 += int(cdat[j+3])
		a3 &= 0xff
		cdat[j+3] = uint8(a3)
		c3 = b3
	}
}

// filterPaethGeneric for generic bytes per pixel
func filterPaethGeneric(cdat, pdat []byte, bytesPerPixel int) {
	var a, b, c, pa, pb, pc int
	for i := 0; i < bytesPerPixel; i++ {
		a, c = 0, 0
		for j := i; j < len(cdat); j += bytesPerPixel {
			b = int(pdat[j])
			pa = b - c
			pb = a - c
			pc = abs(pa + pb)
			pa = abs(pa)
			pb = abs(pb)
			if pa > pb || pa > pc {
				if pb <= pc {
					a = b
				} else {
					a = c
				}
			}
			a += int(cdat[j])
			a &= 0xff
			cdat[j] = uint8(a)
			c = b
		}
	}
}
