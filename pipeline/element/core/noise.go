package core

import (
	"image"
	"sync"

	"github.com/frizinak/phodo/img48"
)

func generateGaussianKernel(size int, sigma float64) [][]int {
	kernel := make([][]int, size)
	center := size / 2
	sum := 0

	for y := 0; y < size; y++ {
		kernel[y] = make([]int, size)
		for x := 0; x < size; x++ {
			weight := gaussian(float64(x-center), float64(y-center), sigma)
			kernel[y][x] = weight
			sum += weight
		}
	}

	if sum == 0 {
		sum = 1
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			kernel[y][x] = (kernel[y][x] * (1<<16 - 1)) / sum
		}
	}

	return kernel
}

func DenoiseChromaGaussian(img *img48.Img, radius int) {
	KernelApplyCbCr(img, generateGaussianKernel(2*radius+1, 1.5))
}

func DenoiseLuminanceMedian(img *img48.Img, readRadius, writeRadius int, blend bool) {
	dim := 2*readRadius + 1
	region := dim * dim
	if writeRadius < 1 {
		writeRadius = readRadius
	}
	writeDim := 2*writeRadius + 1

	width, height := img.Rect.Dx(), img.Rect.Dy()
	orect := image.Rect(0, 0, writeDim, writeDim)
	pix := ycbcr(img)

	_yyl := make([]int, height*region)

	var wg sync.WaitGroup
	for y := 0; y < height; y += writeRadius {
		wg.Add(1)
		go func(y int) {
			for x := 0; x < width; x += writeRadius {
				s, e := y*region, (y+1)*region
				yyl := _yyl[s:s:e]
				for j := y - readRadius; j <= y+readRadius && j < height; j++ {
					if j > 0 {
						o_ := j * img.Stride
						for i := x - readRadius; i <= x+readRadius && i < width; i++ {
							if i >= 0 {
								o := o_ + i*3
								yyl = append(yyl, pix[o])
							}
						}
					}
				}

				yy := median(yyl)

				if blend {
					cv := img48.New(orect, img.Exif)
					for ny := 0; ny < writeDim; ny++ {
						o_ := ny * cv.Stride
						ro_ := (y + ny) * img.Stride
						for nx := 0; nx < writeDim; nx++ {
							o := o_ + nx*3
							ro := ro_ + (x+nx)*3
							if ro >= len(pix) {
								break
							}
							cb := pix[ro+1]
							cr := pix[ro+2]
							cv.Pix[o+0] = intClampUint16((yy + 91881*cr) >> 16)
							cv.Pix[o+1] = intClampUint16((yy - 22554*cb - 46802*cr) >> 16)
							cv.Pix[o+2] = intClampUint16((yy + 116130*cb) >> 16)
						}
					}

					DrawCircleSrc(
						cv,
						img,
						image.Point{writeDim / 2, writeDim / 2},
						image.Point{x + writeDim/2, y + writeDim/2},
						writeDim,
						0,
					)

					continue
				}

				for ny := y; ny < y+writeRadius && ny < height; ny++ {
					o_ := ny * img.Stride
					for nx := x; nx < x+writeRadius && nx < width; nx++ {
						o := o_ + nx*3
						cb := pix[o+1]
						cr := pix[o+2]
						img.Pix[o+0] = intClampUint16((yy + 91881*cr) >> 16)
						img.Pix[o+1] = intClampUint16((yy - 22554*cb - 46802*cr) >> 16)
						img.Pix[o+2] = intClampUint16((yy + 116130*cb) >> 16)
					}
				}
			}
			wg.Done()
		}(y)
	}

	wg.Wait()
}

func DenoiseChromaMedian(img *img48.Img, readRadius, writeRadius int, blend bool) {
	dim := 2*readRadius + 1
	region := dim * dim
	if writeRadius < 1 {
		writeRadius = readRadius
	}
	writeDim := 2*writeRadius + 1

	width, height := img.Rect.Dx(), img.Rect.Dy()
	orect := image.Rect(0, 0, writeDim, writeDim)
	pix := ycbcr(img)

	_cbl := make([]int, height*region)
	_crl := make([]int, height*region)
	var wg sync.WaitGroup
	for y := 0; y < height; y += writeRadius {
		wg.Add(1)
		go func(y int) {
			for x := 0; x < width; x += writeRadius {
				s, e := y*region, (y+1)*region
				cbl := _cbl[s:s:e]
				crl := _crl[s:s:e]
				for j := y - readRadius; j <= y+readRadius && j < height; j++ {
					if j > 0 {
						o_ := j * img.Stride
						for i := x - readRadius; i <= x+readRadius && i < width; i++ {
							if i >= 0 {
								o := o_ + i*3
								cbl = append(cbl, pix[o+1])
								crl = append(crl, pix[o+2])
							}
						}
					}
				}

				cb := median(cbl)
				cr := median(crl)
				r := 91881 * cr
				g := -22554*cb - 46802*cr
				b := 116130 * cb

				if blend {
					cv := img48.New(orect, img.Exif)
					for ny := 0; ny < writeDim; ny++ {
						o_ := ny * cv.Stride
						ro_ := (y + ny) * img.Stride
						for nx := 0; nx < writeDim; nx++ {
							o := o_ + nx*3
							ro := ro_ + (x+nx)*3
							if ro >= len(pix) {
								break
							}
							yy := pix[ro]
							cv.Pix[o+0] = intClampUint16((yy + r) >> 16)
							cv.Pix[o+1] = intClampUint16((yy + g) >> 16)
							cv.Pix[o+2] = intClampUint16((yy + b) >> 16)
						}
					}

					DrawCircleSrc(
						cv,
						img,
						image.Point{writeDim / 2, writeDim / 2},
						image.Point{x + writeDim/2, y + writeDim/2},
						writeDim,
						0,
					)

					continue
				}

				for ny := y; ny < y+writeRadius && ny < height; ny++ {
					o_ := ny * img.Stride
					for nx := x; nx < x+writeRadius && nx < width; nx++ {
						o := o_ + nx*3
						yy := pix[o]
						img.Pix[o+0] = intClampUint16((yy + r) >> 16)
						img.Pix[o+1] = intClampUint16((yy + g) >> 16)
						img.Pix[o+2] = intClampUint16((yy + b) >> 16)
					}
				}
			}
			wg.Done()
		}(y)
	}

	wg.Wait()
}

func ycbcr(img *img48.Img) []int {
	width, height := img.Rect.Dx(), img.Rect.Dy()
	pix := make([]int, len(img.Pix))
	for y := 0; y < height; y++ {
		o_ := y * img.Stride
		for x := 0; x < width; x++ {
			o := o_ + x*3
			r, g, b := int(img.Pix[o+0]), int(img.Pix[o+1]), int(img.Pix[o+2])
			pix[o+0] = 19595*r + 38470*g + 7471*b
			pix[o+1] = (-11056*r - 21712*g + 32768*b) >> 16
			pix[o+2] = (32768*r - 27440*g - 5328*b) >> 16
		}
	}

	return pix
}
