/*
Copyright (c) 2014, Charlie Vieth <charlie.vieth@gmail.com>

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

func nearestGeneric(in image.Image, out *image.RGBA64, starts, counts []int) {
	newBounds := out.Bounds()
	inMinX, inMinY := in.Bounds().Min.X, in.Bounds().Min.Y

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			start := starts[y]
			count := counts[y]
			var rA, gA, bA, aA uint64
			for i := range count {
				r, g, b, a := in.At(start+i+inMinX, x+inMinY).RGBA()
				rA += uint64(r)
				gA += uint64(g)
				bA += uint64(b)
				aA += uint64(a)
			}

			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*8
			d := out.Pix[xo : xo+8 : xo+8]
			c := uint64(count)
			v := uint16(rA / c)
			d[0] = uint8(v >> 8)
			d[1] = uint8(v)
			v = uint16(gA / c)
			d[2] = uint8(v >> 8)
			d[3] = uint8(v)
			v = uint16(bA / c)
			d[4] = uint8(v >> 8)
			d[5] = uint8(v)
			v = uint16(aA / c)
			d[6] = uint8(v >> 8)
			d[7] = uint8(v)
		}
	}
}

func nearestRGBA(in *image.RGBA, out *image.RGBA, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y] * 4
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*4
			d := out.Pix[xo : xo+4 : xo+4]

			if count == 1 {
				p := row[xi : xi+4 : xi+4]
				d[0] = p[0]
				d[1] = p[1]
				d[2] = p[2]
				d[3] = p[3]
				continue
			}

			var rA, gA, bA, aA uint32
			for range count {
				p := row[xi : xi+4 : xi+4]
				rA += uint32(p[0])
				gA += uint32(p[1])
				bA += uint32(p[2])
				aA += uint32(p[3])
				xi += 4
			}
			c := uint32(count)
			d[0] = uint8(rA / c)
			d[1] = uint8(gA / c)
			d[2] = uint8(bA / c)
			d[3] = uint8(aA / c)
		}
	}
}

func nearestNRGBA(in *image.NRGBA, out *image.NRGBA, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y] * 4
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*4
			d := out.Pix[xo : xo+4 : xo+4]

			if count == 1 {
				p := row[xi : xi+4 : xi+4]
				d[0] = p[0]
				d[1] = p[1]
				d[2] = p[2]
				d[3] = p[3]
				continue
			}

			var rA, gA, bA, aA uint32
			for range count {
				p := row[xi : xi+4 : xi+4]
				rA += uint32(p[0])
				gA += uint32(p[1])
				bA += uint32(p[2])
				aA += uint32(p[3])
				xi += 4
			}
			c := uint32(count)
			d[0] = uint8(rA / c)
			d[1] = uint8(gA / c)
			d[2] = uint8(bA / c)
			d[3] = uint8(aA / c)
		}
	}
}

func nearestRGBA64(in *image.RGBA64, out *image.RGBA64, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y] * 8
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*8
			d := out.Pix[xo : xo+8 : xo+8]

			if count == 1 {
				p := row[xi : xi+8 : xi+8]
				copy(d, p)
				continue
			}

			var rA, gA, bA, aA uint64
			for range count {
				p := row[xi : xi+8 : xi+8]
				rA += uint64(uint16(p[0])<<8 | uint16(p[1]))
				gA += uint64(uint16(p[2])<<8 | uint16(p[3]))
				bA += uint64(uint16(p[4])<<8 | uint16(p[5]))
				aA += uint64(uint16(p[6])<<8 | uint16(p[7]))
				xi += 8
			}
			c := uint64(count)
			v := uint16(rA / c)
			d[0] = uint8(v >> 8)
			d[1] = uint8(v)
			v = uint16(gA / c)
			d[2] = uint8(v >> 8)
			d[3] = uint8(v)
			v = uint16(bA / c)
			d[4] = uint8(v >> 8)
			d[5] = uint8(v)
			v = uint16(aA / c)
			d[6] = uint8(v >> 8)
			d[7] = uint8(v)
		}
	}
}

func nearestNRGBA64(in *image.NRGBA64, out *image.NRGBA64, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y] * 8
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*8
			d := out.Pix[xo : xo+8 : xo+8]

			if count == 1 {
				p := row[xi : xi+8 : xi+8]
				copy(d, p)
				continue
			}

			var rA, gA, bA, aA uint64
			for range count {
				p := row[xi : xi+8 : xi+8]
				rA += uint64(uint16(p[0])<<8 | uint16(p[1]))
				gA += uint64(uint16(p[2])<<8 | uint16(p[3]))
				bA += uint64(uint16(p[4])<<8 | uint16(p[5]))
				aA += uint64(uint16(p[6])<<8 | uint16(p[7]))
				xi += 8
			}
			c := uint64(count)
			v := uint16(rA / c)
			d[0] = uint8(v >> 8)
			d[1] = uint8(v)
			v = uint16(gA / c)
			d[2] = uint8(v >> 8)
			d[3] = uint8(v)
			v = uint16(bA / c)
			d[4] = uint8(v >> 8)
			d[5] = uint8(v)
			v = uint16(aA / c)
			d[6] = uint8(v >> 8)
			d[7] = uint8(v)
		}
	}
}

func nearestGray(in *image.Gray, out *image.Gray, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y]
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x - newBounds.Min.X)

			if count == 1 {
				out.Pix[xo] = row[xi]
				continue
			}

			var gray uint32
			for range count {
				gray += uint32(row[xi])
				xi++
			}
			out.Pix[xo] = uint8(gray / uint32(count))
		}
	}
}

func nearestGray16(in *image.Gray16, out *image.Gray16, starts, counts []int) {
	newBounds := out.Bounds()

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			xi := starts[y] * 2
			count := counts[y]
			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*2
			d := out.Pix[xo : xo+2 : xo+2]

			if count == 1 {
				p := row[xi : xi+2 : xi+2]
				d[0] = p[0]
				d[1] = p[1]
				continue
			}

			var gray uint64
			for range count {
				p := row[xi : xi+2 : xi+2]
				gray += uint64(uint16(p[0])<<8 | uint16(p[1]))
				xi += 2
			}
			v := uint16(gray / uint64(count))
			d[0] = uint8(v >> 8)
			d[1] = uint8(v)
		}
	}
}
