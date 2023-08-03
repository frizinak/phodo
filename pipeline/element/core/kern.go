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
