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

import (
	"math"
)

// Filter kernel coefficients and thresholds
const (
	// Mitchell-Netravali filter constants (B=1/3, C=1/3)
	// First region (|x| <= 1)
	mitchellCoeff1 = 7.0
	mitchellCoeff2 = 12.0
	mitchellCoeff3 = 5.33333333333 // (16/3)
	mitchellScale  = 1.0 / 6.0

	// Second region (1 < |x| <= 2)
	mitchellCoeff2b = 2.33333333333 // (7/3)
	mitchellCoeff7  = 20.0
	mitchellCoeff8  = 10.6666666667 // (32/3)

	// Sinc function threshold for avoiding division by near-zero
	sincThreshold = 1.220703e-4
)

func nearest(in float64) float64 {
	if in >= -0.5 && in < 0.5 {
		return 1
	}
	return 0
}

func linear(in float64) float64 {
	in = math.Abs(in)
	if in <= 1 {
		return 1 - in
	}
	return 0
}

func cubic(in float64) float64 {
	in = math.Abs(in)
	if in <= 1 {
		return in*in*(1.5*in-2.5) + 1.0
	}
	if in <= 2 {
		return in*(in*(2.5-0.5*in)-4.0) + 2.0
	}
	return 0
}

func mitchellnetravali(in float64) float64 {
	in = math.Abs(in)
	if in <= 1 {
		return (mitchellCoeff1*in*in*in - mitchellCoeff2*in*in + mitchellCoeff3) * mitchellScale
	}
	if in <= 2 {
		return (-mitchellCoeff2b*in*in*in + mitchellCoeff2*in*in - mitchellCoeff7*in + mitchellCoeff8) * mitchellScale
	}
	return 0
}

func sinc(x float64) float64 {
	x = math.Abs(x) * math.Pi
	if x >= sincThreshold {
		return math.Sin(x) / x
	}
	return 1
}

// Weights calculation scale factors
const (
	weightScale8  = 256   // scaling factor for 8-bit coefficients
	weightScale16 = 65536 // scaling factor for 16-bit coefficients

	// Lanczos2 and Lanczos3 constants
	lanczos2Ratio = 0.5       // a = 2
	lanczos3Ratio = 1.0 / 3.0 // a = 3
)

func lanczos2(in float64) float64 {
	if in > -2 && in < 2 {
		return sinc(in) * sinc(in*lanczos2Ratio)
	}
	return 0
}

func lanczos3(in float64) float64 {
	if in > -3 && in < 3 {
		return sinc(in) * sinc(in*lanczos3Ratio)
	}
	return 0
}

// range [-256,256]
func createWeights8(dy, filterLength int, blur, scale float64, kernel func(float64) float64) ([]int16, []int, int) {
	filterLength = filterLength * int(math.Max(math.Ceil(blur*scale), 1))
	filterFactor := math.Min(1./(blur*scale), 1)

	coeffs := make([]int16, dy*filterLength)
	start := make([]int, dy)
	for y := range dy {
		interpX := scale*(float64(y)+0.5) - 0.5
		start[y] = int(interpX) - filterLength/2 + 1
		interpX -= float64(start[y])
		for i := range filterLength {
			in := (interpX - float64(i)) * filterFactor
			coeffs[y*filterLength+i] = int16(kernel(in) * weightScale8)
		}
	}

	return coeffs, start, filterLength
}

// range [-65536,65536]
func createWeights16(dy, filterLength int, blur, scale float64, kernel func(float64) float64) ([]int32, []int, int) {
	filterLength = filterLength * int(math.Max(math.Ceil(blur*scale), 1))
	filterFactor := math.Min(1./(blur*scale), 1)

	coeffs := make([]int32, dy*filterLength)
	start := make([]int, dy)
	for y := range dy {
		interpX := scale*(float64(y)+0.5) - 0.5
		start[y] = int(interpX) - filterLength/2 + 1
		interpX -= float64(start[y])
		for i := range filterLength {
			in := (interpX - float64(i)) * filterFactor
			coeffs[y*filterLength+i] = int32(kernel(in) * weightScale16)
		}
	}

	return coeffs, start, filterLength
}

func createWeightsNearest(dy, filterLength int, blur, scale float64) ([]bool, []int, int) {
	filterLength = filterLength * int(math.Max(math.Ceil(blur*scale), 1))
	filterFactor := math.Min(1./(blur*scale), 1)

	coeffs := make([]bool, dy*filterLength)
	start := make([]int, dy)
	for y := range dy {
		interpX := scale*(float64(y)+0.5) - 0.5
		start[y] = int(interpX) - filterLength/2 + 1
		interpX -= float64(start[y])
		for i := range filterLength {
			in := (interpX - float64(i)) * filterFactor
			if in >= -0.5 && in < 0.5 {
				coeffs[y*filterLength+i] = true
			} else {
				coeffs[y*filterLength+i] = false
			}
		}
	}

	return coeffs, start, filterLength
}
