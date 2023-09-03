package core

import (
	"sync"

	"github.com/frizinak/phodo/img48"
)

func p48(img *img48.Img, cb func(pix []uint16, y int)) {
	var (
		wg sync.WaitGroup
		h  = img.Rect.Dy()
		s  = img.Stride
	)

	wg.Add(h)
	for y := 0; y < h; y++ {
		go func(y int) {
			o := y * s
			cb(img.Pix[o:o+s:o+s], y)
			wg.Done()
		}(y)
	}

	wg.Wait()
}

func p48x(img *img48.Img, cb func(offset, x int)) {
	var (
		wg sync.WaitGroup
		w  = img.Rect.Dx()
	)

	wg.Add(w)
	for x := 0; x < w; x++ {
		go func(x int) {
			cb(x*3, x)
			wg.Done()
		}(x)
	}

	wg.Wait()
}
