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

func Test_FloatToUint8(t *testing.T) {
	var testData = []struct {
		in       float32
		expected uint8
	}{
		{0, 0},
		{255, 255},
		{128, 128},
		{1, 1},
		{256, 255},
	}
	for _, test := range testData {
		actual := floatToUint8(test.in)
		if actual != test.expected {
			t.Fail()
		}
	}
}

func Test_FloatToUint16(t *testing.T) {
	var testData = []struct {
		in       float32
		expected uint16
	}{
		{0, 0},
		{65535, 65535},
		{128, 128},
		{1, 1},
		{65536, 65535},
	}
	for _, test := range testData {
		actual := floatToUint16(test.in)
		if actual != test.expected {
			t.Fail()
		}
	}
}

func TestNearestGeneric(t *testing.T) {
	in := image.NewRGBA(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.RGBA{255, 0, 0, 255})
	out := image.NewRGBA64(image.Rect(0, 0, 2, 2))
	coeffs := []bool{true, false, false, false}
	offset := []int{0, 0}

	nearestGeneric(in, out, 1.0, coeffs, offset, 2)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestGeneric failed to process color properly")
	}
}

func TestNearestRGBA(t *testing.T) {
	in := image.NewRGBA(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.RGBA{255, 0, 0, 255})
	out := image.NewRGBA(image.Rect(0, 0, 2, 2))
	coeffs := []bool{true, false, false, false}
	offset := []int{0, 0}

	nearestRGBA(in, out, 1.0, coeffs, offset, 2)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestRGBA failed to process color properly")
	}
}

func TestNearestNRGBA(t *testing.T) {
	in := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.NRGBA{255, 0, 0, 255})
	out := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	coeffs := []bool{true, false, false, false}
	offset := []int{0, 0}

	nearestNRGBA(in, out, 1.0, coeffs, offset, 2)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestNRGBA failed to process color properly")
	}
}

func TestNearestRGBA64(t *testing.T) {
	in := image.NewRGBA64(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.RGBA64{65535, 0, 0, 65535})
	out := image.NewRGBA64(image.Rect(0, 0, 2, 2))
	coeffs := []bool{true, false, false, false}
	offset := []int{0, 0}

	nearestRGBA64(in, out, 1.0, coeffs, offset, 2)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestRGBA64 failed to process color properly")
	}
}

func TestNearestNRGBA64(t *testing.T) {
	in := image.NewNRGBA64(image.Rect(0, 0, 2, 2))
	in.Set(0, 0, color.NRGBA64{65535, 0, 0, 65535})
	out := image.NewNRGBA64(image.Rect(0, 0, 2, 2))
	coeffs := []bool{true, false, false, false}
	offset := []int{0, 0}

	nearestNRGBA64(in, out, 1.0, coeffs, offset, 2)

	r, _, _, _ := out.At(0, 0).RGBA()
	if r == 0 {
		t.Errorf("nearestNRGBA64 failed to process color properly")
	}
}
