/*
Copyright (c) 2012, Jan Schlicht <jan.schlicht@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
*/

package resize

import "image"

// Constants for bit depth clamping operations
const (
	// uint8 maximum value
	maxUint8 = 0xFF
	// uint8 range for overflow detection
	uint8Range = 256
	// uint16 maximum value
	maxUint16 = 0xFFFF
	// uint16 range for overflow detection
	uint16Range = 65536
)

// Keep value in [0,255] range.
func clampUint8(in int32) uint8 {
	// casting a negative int to an uint will result in an overflown
	// large uint. this behavior will be exploited here and in other functions
	// to achieve a higher performance.
	if uint32(in) < uint8Range {
		return uint8(in)
	}
	if in > maxUint8 {
		return maxUint8
	}
	return 0
}

// Keep value in [0,65535] range.
func clampUint16(in int64) uint16 {
	if uint64(in) < uint16Range {
		return uint16(in)
	}
	if in > maxUint16 {
		return maxUint16
	}
	return 0
}

func resizeGeneric(in image.Image, out *image.RGBA64, coeffs []int64, starts, lens []int, sums []int64, stride int) {
	newBounds := out.Bounds()
	inMinX, inMinY := in.Bounds().Min.X, in.Bounds().Min.Y

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var rA, gA, bA, aA int64
			start := starts[y]
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for i, coeff := range cRow {
				r, g, b, a := in.At(start+i+inMinX, x+inMinY).RGBA()
				rA += coeff * int64(r)
				gA += coeff * int64(g)
				bA += coeff * int64(b)
				aA += coeff * int64(a)
			}

			sum := sums[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*8
			d := out.Pix[xo : xo+8 : xo+8]
			value := clampUint16(rA / sum)
			d[0] = uint8(value >> 8)
			d[1] = uint8(value)
			value = clampUint16(gA / sum)
			d[2] = uint8(value >> 8)
			d[3] = uint8(value)
			value = clampUint16(bA / sum)
			d[4] = uint8(value >> 8)
			d[5] = uint8(value)
			value = clampUint16(aA / sum)
			d[6] = uint8(value >> 8)
			d[7] = uint8(value)
		}
	}
}

func resizeRGBA(in *image.RGBA, out *image.RGBA, coeffs []int32, starts, lens []int, sums []int32, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var rA, gA, bA, aA int32
			xi := starts[y] * 4
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				p := row[xi : xi+4 : xi+4]
				rA += coeff * int32(p[0])
				gA += coeff * int32(p[1])
				bA += coeff * int32(p[2])
				aA += coeff * int32(p[3])
				xi += 4
			}

			sum := sums[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*4
			d := out.Pix[xo : xo+4 : xo+4]
			d[0] = clampUint8(rA / sum)
			d[1] = clampUint8(gA / sum)
			d[2] = clampUint8(bA / sum)
			d[3] = clampUint8(aA / sum)
		}
	}
}

func resizeNRGBA(in *image.NRGBA, out *image.RGBA, coeffs []int32, starts, lens []int, sums []int32, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var rA, gA, bA, aA int32
			xi := starts[y] * 4
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				p := row[xi : xi+4 : xi+4]
				// Forward alpha-premultiplication
				a := int32(p[3])
				r := (int32(p[0]) * a) / 0xff
				g := (int32(p[1]) * a) / 0xff
				b := (int32(p[2]) * a) / 0xff

				rA += coeff * r
				gA += coeff * g
				bA += coeff * b
				aA += coeff * a
				xi += 4
			}

			sum := sums[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*4
			d := out.Pix[xo : xo+4 : xo+4]
			d[0] = clampUint8(rA / sum)
			d[1] = clampUint8(gA / sum)
			d[2] = clampUint8(bA / sum)
			d[3] = clampUint8(aA / sum)
		}
	}
}

func resizeRGBA64(in *image.RGBA64, out *image.RGBA64, coeffs []int64, starts, lens []int, sums []int64, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var rA, gA, bA, aA int64
			xi := starts[y] * 8
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				p := row[xi : xi+8 : xi+8]
				rA += coeff * (int64(p[0])<<8 | int64(p[1]))
				gA += coeff * (int64(p[2])<<8 | int64(p[3]))
				bA += coeff * (int64(p[4])<<8 | int64(p[5]))
				aA += coeff * (int64(p[6])<<8 | int64(p[7]))
				xi += 8
			}

			sum := sums[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*8
			d := out.Pix[xo : xo+8 : xo+8]
			value := clampUint16(rA / sum)
			d[0] = uint8(value >> 8)
			d[1] = uint8(value)
			value = clampUint16(gA / sum)
			d[2] = uint8(value >> 8)
			d[3] = uint8(value)
			value = clampUint16(bA / sum)
			d[4] = uint8(value >> 8)
			d[5] = uint8(value)
			value = clampUint16(aA / sum)
			d[6] = uint8(value >> 8)
			d[7] = uint8(value)
		}
	}
}

func resizeNRGBA64(in *image.NRGBA64, out *image.RGBA64, coeffs []int64, starts, lens []int, sums []int64, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var rA, gA, bA, aA int64
			xi := starts[y] * 8
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				p := row[xi : xi+8 : xi+8]
				// Forward alpha-premultiplication
				a := int64(uint16(p[6])<<8 | uint16(p[7]))
				r := (int64(uint16(p[0])<<8|uint16(p[1])) * a) / 0xffff
				g := (int64(uint16(p[2])<<8|uint16(p[3])) * a) / 0xffff
				b := (int64(uint16(p[4])<<8|uint16(p[5])) * a) / 0xffff

				rA += coeff * r
				gA += coeff * g
				bA += coeff * b
				aA += coeff * a
				xi += 8
			}

			sum := sums[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*8
			d := out.Pix[xo : xo+8 : xo+8]
			value := clampUint16(rA / sum)
			d[0] = uint8(value >> 8)
			d[1] = uint8(value)
			value = clampUint16(gA / sum)
			d[2] = uint8(value >> 8)
			d[3] = uint8(value)
			value = clampUint16(bA / sum)
			d[4] = uint8(value >> 8)
			d[5] = uint8(value)
			value = clampUint16(aA / sum)
			d[6] = uint8(value >> 8)
			d[7] = uint8(value)
		}
	}
}

func resizeGray(in *image.Gray, out *image.Gray, coeffs []int32, starts, lens []int, sums []int32, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var gray int32
			xi := starts[y]
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				gray += coeff * int32(row[xi])
				xi++
			}

			offset := (y-newBounds.Min.Y)*out.Stride + (x - newBounds.Min.X)
			out.Pix[offset] = clampUint8(gray / sums[y])
		}
	}
}

func resizeGray16(in *image.Gray16, out *image.Gray16, coeffs []int64, starts, lens []int, sums []int64, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var gray int64
			xi := starts[y] * 2
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				p := row[xi : xi+2 : xi+2]
				gray += coeff * int64(uint16(p[0])<<8|uint16(p[1]))
				xi += 2
			}

			offset := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*2
			d := out.Pix[offset : offset+2 : offset+2]
			value := clampUint16(gray / sums[y])
			d[0] = uint8(value >> 8)
			d[1] = uint8(value)
		}
	}
}

func resizeYCbCr(in *ycc, out *ycc, coeffs []int32, starts, lens []int, sums []int32, stride int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var p [3]int32
			xi := starts[y] * 3
			l := lens[y]
			cRow := coeffs[y*stride : y*stride+l : y*stride+l]
			for _, coeff := range cRow {
				pix := row[xi : xi+3 : xi+3]
				p[0] += coeff * int32(pix[0])
				p[1] += coeff * int32(pix[1])
				p[2] += coeff * int32(pix[2])
				xi += 3
			}

			sum := sums[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*3
			d := out.Pix[xo : xo+3 : xo+3]
			d[0] = clampUint8(p[0] / sum)
			d[1] = clampUint8(p[1] / sum)
			d[2] = clampUint8(p[2] / sum)
		}
	}
}

func nearestYCbCr(in *ycc, out *ycc, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y] * 3
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*3
			d := out.Pix[xo : xo+3 : xo+3]

			if count == 1 {
				pix := row[xi : xi+3 : xi+3]
				d[0] = pix[0]
				d[1] = pix[1]
				d[2] = pix[2]
				continue
			}

			var p [3]uint32
			for range count {
				pix := row[xi : xi+3 : xi+3]
				p[0] += uint32(pix[0])
				p[1] += uint32(pix[1])
				p[2] += uint32(pix[2])
				xi += 3
			}
			c := uint32(count)
			d[0] = uint8(p[0] / c)
			d[1] = uint8(p[1] / c)
			d[2] = uint8(p[2] / c)
		}
	}
}
