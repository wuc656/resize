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

import (
	"image"
	"image/color"
	"testing"
)

func TestNearestGeneric(t *testing.T) {
	in := image.NewRGBA(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.RGBA{255, 0, 0, 255})
	out := image.NewRGBA64(image.Rect(0, 0, 2, 2))
	starts := []int{0, 0}
	counts := []int{1, 1}

	nearestGeneric(in, out, starts, counts)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestGeneric failed to process color properly")
	}
}

func TestNearestRGBA(t *testing.T) {
	in := image.NewRGBA(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.RGBA{255, 0, 0, 255})
	out := image.NewRGBA(image.Rect(0, 0, 2, 2))
	starts := []int{0, 0}
	counts := []int{1, 1}

	nearestRGBA(in, out, starts, counts)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestRGBA failed to process color properly")
	}

	// Test downscale averaging (count > 1)
	in.Set(1, 0, color.RGBA{0, 0, 255, 255})
	nearestRGBA(in, out, []int{0, 0}, []int{2, 1})
	c := out.At(0, 0).(color.RGBA)
	if c.R != 127 || c.B != 127 {
		t.Errorf("nearestRGBA count=2: got %+v, want R=127 B=127", c)
	}
}

func TestNearestNRGBA(t *testing.T) {
	in := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.NRGBA{255, 0, 0, 255})
	out := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	starts := []int{0, 0}
	counts := []int{1, 1}

	nearestNRGBA(in, out, starts, counts)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestNRGBA failed to process color properly")
	}
}

func TestNearestRGBA64(t *testing.T) {
	in := image.NewRGBA64(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.RGBA64{65535, 0, 0, 65535})
	out := image.NewRGBA64(image.Rect(0, 0, 2, 2))
	starts := []int{0, 0}
	counts := []int{1, 1}

	nearestRGBA64(in, out, starts, counts)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestRGBA64 failed to process color properly")
	}
}

func TestNearestNRGBA64(t *testing.T) {
	in := image.NewNRGBA64(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.NRGBA64{65535, 0, 0, 65535})
	out := image.NewNRGBA64(image.Rect(0, 0, 2, 2))
	starts := []int{0, 0}
	counts := []int{1, 1}

	nearestNRGBA64(in, out, starts, counts)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestNRGBA64 failed to process color properly")
	}
}
