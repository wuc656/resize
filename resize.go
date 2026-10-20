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

// Package resize implements various image resizing methods.
//
// The package works with the Image interface described in the image package.
// Various interpolation methods are provided and multiple processors may be
// utilized in the computations.
//
// Example:
//
//	imgResized := resize.Resize(1000, 0, imgOld, resize.MitchellNetravali)
package resize

import (
	"image"
	"runtime"
	"sync"
)

// An InterpolationFunction provides the parameters that describe an
// interpolation kernel. It returns the number of samples to take
// and the kernel function to use for sampling.
type InterpolationFunction int

// InterpolationFunction constants
const (
	// Nearest-neighbor interpolation
	NearestNeighbor InterpolationFunction = iota
	// Bilinear interpolation
	Bilinear
	// Bicubic interpolation (with cubic hermite spline)
	Bicubic
	// Mitchell-Netravali interpolation
	MitchellNetravali
	// Lanczos interpolation (a=2)
	Lanczos2
	// Lanczos interpolation (a=3)
	Lanczos3
)

// kernal, returns an InterpolationFunctions taps and kernel.
func (i InterpolationFunction) kernel() (int, func(float64) float64) {
	switch i {
	case Bilinear:
		return 2, linear
	case Bicubic:
		return 4, cubic
	case MitchellNetravali:
		return 4, mitchellnetravali
	case Lanczos2:
		return 4, lanczos2
	case Lanczos3:
		return 6, lanczos3
	default:
		// Default to NearestNeighbor.
		return 2, nearest
	}
}

// values <1 will sharpen the image
const blur = 1.0

// aspectRatioRoundingBias is used when preserving aspect ratio with 0 dimensions.
// The value 0.7 biases rounding upward to avoid premature truncation when converting
// the calculated float dimension to uint, ensuring proper aspect ratio preservation.
const aspectRatioRoundingBias = 0.7

// Resize scales an image to new width and height using the interpolation function interp.
// A new image with the given dimensions will be returned.
// If one of the parameters width or height is set to 0, its size will be calculated so that
// the aspect ratio is that of the originating image.
// The resizing algorithm uses channels for parallel computation.
// If the input image has width or height of 0, it is returned unchanged.
func Resize(width, height uint, img image.Image, interp InterpolationFunction) image.Image {
	scaleX, scaleY := calcFactors(width, height, float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))
	if width == 0 {
		width = uint(aspectRatioRoundingBias + float64(img.Bounds().Dx())/scaleX)
	}
	if height == 0 {
		height = uint(aspectRatioRoundingBias + float64(img.Bounds().Dy())/scaleY)
	}

	// Trivial case: return input image
	if int(width) == img.Bounds().Dx() && int(height) == img.Bounds().Dy() {
		return img
	}

	// Input image has no pixels
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		return img
	}

	if interp == NearestNeighbor {
		return resizeNearest(width, height, scaleX, scaleY, img, interp)
	}

	taps, kernel := interp.kernel()

	// Generic access to image.Image is slow in tight loops.
	// The optimal access has to be determined from the concrete image type.
	switch input := img.(type) {
	case *image.RGBA:
		// 8-bit precision
		temp := image.NewRGBA(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights8(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA)
			resizeRGBA(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights8(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA)
			resizeRGBA(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.NRGBA:
		// 8-bit precision
		temp := image.NewRGBA(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights8(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA)
			resizeNRGBA(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights8(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA)
			resizeRGBA(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result

	case *image.YCbCr:
		// 8-bit precision
		// accessing the YCbCr arrays in a tight loop is slow.
		// converting the image to ycc increases performance by 2x.
		temp := newYCC(image.Rect(0, 0, input.Bounds().Dy(), int(width)), input.SubsampleRatio)
		result := newYCC(image.Rect(0, 0, int(width), int(height)), image.YCbCrSubsampleRatio444)

		coeffs, offset, filterLength := createWeights8(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		in := imageYCbCrToYCC(input)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*ycc)
			resizeYCbCr(in, slice, scaleX, coeffs, offset, filterLength)
		})

		coeffs, offset, filterLength = createWeights8(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*ycc)
			resizeYCbCr(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result.YCbCr()
	case *image.RGBA64:
		// 16-bit precision
		temp := image.NewRGBA64(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewRGBA64(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights16(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA64)
			resizeRGBA64(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights16(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA64)
			resizeRGBA64(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.NRGBA64:
		// 16-bit precision
		temp := image.NewRGBA64(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewRGBA64(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights16(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA64)
			resizeNRGBA64(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights16(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA64)
			resizeRGBA64(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.Gray:
		// 8-bit precision
		temp := image.NewGray(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewGray(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights8(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.Gray)
			resizeGray(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights8(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.Gray)
			resizeGray(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.Gray16:
		// 16-bit precision
		temp := image.NewGray16(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewGray16(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights16(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.Gray16)
			resizeGray16(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights16(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.Gray16)
			resizeGray16(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	default:
		// 16-bit precision
		temp := image.NewRGBA64(image.Rect(0, 0, img.Bounds().Dy(), int(width)))
		result := image.NewRGBA64(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeights16(temp.Bounds().Dy(), taps, blur, scaleX, kernel)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA64)
			resizeGeneric(img, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeights16(result.Bounds().Dy(), taps, blur, scaleY, kernel)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA64)
			resizeRGBA64(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	}
}

func resizeNearest(width, height uint, scaleX, scaleY float64, img image.Image, interp InterpolationFunction) image.Image {
	taps, _ := interp.kernel()

	switch input := img.(type) {
	case *image.RGBA:
		// 8-bit precision
		temp := image.NewRGBA(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA)
			nearestRGBA(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA)
			nearestRGBA(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.NRGBA:
		// 8-bit precision
		temp := image.NewNRGBA(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewNRGBA(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.NRGBA)
			nearestNRGBA(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.NRGBA)
			nearestNRGBA(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.YCbCr:
		// 8-bit precision
		// accessing the YCbCr arrays in a tight loop is slow.
		// converting the image to ycc increases performance by 2x.
		temp := newYCC(image.Rect(0, 0, input.Bounds().Dy(), int(width)), input.SubsampleRatio)
		result := newYCC(image.Rect(0, 0, int(width), int(height)), image.YCbCrSubsampleRatio444)

		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		in := imageYCbCrToYCC(input)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*ycc)
			nearestYCbCr(in, slice, scaleX, coeffs, offset, filterLength)
		})

		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*ycc)
			nearestYCbCr(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result.YCbCr()
	case *image.RGBA64:
		// 16-bit precision
		temp := image.NewRGBA64(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewRGBA64(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA64)
			nearestRGBA64(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA64)
			nearestRGBA64(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.NRGBA64:
		// 16-bit precision
		temp := image.NewNRGBA64(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewNRGBA64(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.NRGBA64)
			nearestNRGBA64(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.NRGBA64)
			nearestNRGBA64(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.Gray:
		// 8-bit precision
		temp := image.NewGray(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewGray(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.Gray)
			nearestGray(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.Gray)
			nearestGray(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	case *image.Gray16:
		// 16-bit precision
		temp := image.NewGray16(image.Rect(0, 0, input.Bounds().Dy(), int(width)))
		result := image.NewGray16(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.Gray16)
			nearestGray16(input, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.Gray16)
			nearestGray16(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	default:
		// 16-bit precision
		temp := image.NewRGBA64(image.Rect(0, 0, img.Bounds().Dy(), int(width)))
		result := image.NewRGBA64(image.Rect(0, 0, int(width), int(height)))

		// horizontal filter, results in transposed temporary image
		coeffs, offset, filterLength := createWeightsNearest(temp.Bounds().Dy(), taps, blur, scaleX)
		runParallel(numCPUs(temp.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(temp, i, n).(*image.RGBA64)
			nearestGeneric(img, slice, scaleX, coeffs, offset, filterLength)
		})

		// horizontal filter on transposed image, result is not transposed
		coeffs, offset, filterLength = createWeightsNearest(result.Bounds().Dy(), taps, blur, scaleY)
		runParallel(numCPUs(result.Bounds().Dy()), func(i, n int) {
			slice := makeSlice(result, i, n).(*image.RGBA64)
			nearestRGBA64(temp, slice, scaleY, coeffs, offset, filterLength)
		})
		return result
	}

}

func runParallel(n int, fn func(i, total int)) {
	if n <= 1 {
		fn(0, 1)
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := range n {
		go func(idx int) {
			defer wg.Done()
			fn(idx, n)
		}(i)
	}
	wg.Wait()
}

func numCPUs(h int) int {
	cpus := runtime.GOMAXPROCS(0)
	if h < cpus {
		return max(h, 1)
	}
	return cpus
}

// Calculates scaling factors using old and new image dimensions.
func calcFactors(width, height uint, oldWidth, oldHeight float64) (scaleX, scaleY float64) {
	if width == 0 {
		if height == 0 {
			scaleX = 1.0
			scaleY = 1.0
		} else {
			scaleY = oldHeight / float64(height)
			scaleX = scaleY
		}
	} else {
		scaleX = oldWidth / float64(width)
		if height == 0 {
			scaleY = scaleX
		} else {
			scaleY = oldHeight / float64(height)
		}
	}
	return
}

type imageWithSubImage interface {
	image.Image
	SubImage(image.Rectangle) image.Image
}

func makeSlice(img imageWithSubImage, i, n int) image.Image {
	return img.SubImage(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y+i*img.Bounds().Dy()/n, img.Bounds().Max.X, img.Bounds().Min.Y+(i+1)*img.Bounds().Dy()/n))
}
