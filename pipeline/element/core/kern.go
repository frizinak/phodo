package core

import (
	"sync"

	"github.com/frizinak/phodo/img48"
)

func KernelApplyCbCr(img *img48.Img, kernel [][]int) {
	pix := ycbcr(img)
	width, height := img.Rect.Dx(), img.Rect.Dy()
	radius := len(kernel) / 2

	var wg sync.WaitGroup
	for y := radius; y < height-radius; y++ {
		o_ := y * img.Stride
		wg.Add(1)
		go func(y, o_ int) {
			for x := radius; x < width-radius; x++ {
				o := o_ + x*3
				var sumCb, sumCr int
				for j := -radius; j <= radius; j++ {
					o_ := (y + j) * img.Stride
					kern := kernel[j+radius]
					for i := -radius; i <= radius; i++ {
						o := o_ + (x+i)*3
						k := kern[i+radius]
						sumCb += pix[o+1] * k
						sumCr += pix[o+2] * k
					}
				}

				yy := pix[o]
				cb := sumCb >> 16
				cr := sumCr >> 16

				r := 91881 * cr
				g := -22554*cb - 46802*cr
				b := 116130 * cb

				img.Pix[o+0] = intClampUint16((yy + r) >> 16)
				img.Pix[o+1] = intClampUint16((yy + g) >> 16)
				img.Pix[o+2] = intClampUint16((yy + b) >> 16)
			}
			wg.Done()
		}(y, o_)
	}

	wg.Wait()
}

func KernelApplyY(img *img48.Img, kernel [][]int) {
	pix := ycbcr(img)
	width, height := img.Rect.Dx(), img.Rect.Dy()
	radius := len(kernel) / 2

	var wg sync.WaitGroup
	for y := radius; y < height-radius; y++ {
		o_ := y * img.Stride
		wg.Add(1)
		go func(y, o_ int) {
			for x := radius; x < width-radius; x++ {
				o := o_ + x*3
				var sumYY int
				for j := -radius; j <= radius; j++ {
					o_ := (y + j) * img.Stride
					kern := kernel[j+radius]
					for i := -radius; i <= radius; i++ {
						o := o_ + (x+i)*3
						k := kern[i+radius]
						sumYY += pix[o] * k
					}
				}

				yy := sumYY
				cb := pix[o+1]
				cr := pix[o+2]

				r := 91881 * cr
				g := -22554*cb - 46802*cr
				b := 116130 * cb

				img.Pix[o+0] = intClampUint16((yy + r) >> 16)
				img.Pix[o+1] = intClampUint16((yy + g) >> 16)
				img.Pix[o+2] = intClampUint16((yy + b) >> 16)
			}
			wg.Done()
		}(y, o_)
	}

	wg.Wait()
}

func KernelApply(img *img48.Img, kernel [][]int) {
	pix := make([]int, len(img.Pix))
	for i := range img.Pix {
		pix[i] = int(img.Pix[i])
	}

	width, height := img.Rect.Dx(), img.Rect.Dy()
	radius := len(kernel) / 2

	var wg sync.WaitGroup
	for y := radius; y < height-radius; y++ {
		o_ := y * img.Stride
		wg.Add(1)
		go func(y, o_ int) {
			for x := radius; x < width-radius; x++ {
				o := o_ + x*3
				var sumR, sumG, sumB int
				for j := -radius; j <= radius; j++ {
					o_ := (y + j) * img.Stride
					kern := kernel[j+radius]
					for i := -radius; i <= radius; i++ {
						o := o_ + (x+i)*3
						k := kern[i+radius]
						sumR += int(pix[o+0]) * k
						sumG += int(pix[o+1]) * k
						sumB += int(pix[o+2]) * k
					}
				}

				img.Pix[o+0] = intClampUint16(sumR)
				img.Pix[o+1] = intClampUint16(sumG)
				img.Pix[o+2] = intClampUint16(sumB)
			}
			wg.Done()
		}(y, o_)
	}

	wg.Wait()
}
