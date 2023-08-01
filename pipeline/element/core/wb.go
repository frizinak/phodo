package core

import "github.com/frizinak/phodo/img48"

func WhiteBalanceCalc(img *img48.Img, x, y, radius int) (r, g, b float64) {
	var rs, gs, bs uint64
	var n uint64
	o := linehorizcb(img, func(y int, pix []uint16) {
		for i := 0; i < len(pix); i += 3 {
			n++
			rs += uint64(pix[i+0])
			gs += uint64(pix[i+1])
			bs += uint64(pix[i+2])
		}
	})

	x_ := radius
	y_ := 0
	e := 0

	for x_ >= y_ {
		o(x-x_, x+x_, y+y_)
		o(x-x_, x+x_, y-y_)
		o(x-y_, x+y_, y+x_)
		o(x-y_, x+y_, y-x_)
		if e <= 0 {
			y_ += 1
			e += 2*y_ + 1
		}

		if e > 0 {
			x_ -= 1
			e -= 2*x_ + 1
		}
	}

	if n == 0 {
		n = 1
	}

	rs /= n
	gs /= n
	bs /= n
	avg := (rs + gs + bs) / 3

	r = float64(avg) / float64(rs)
	g = float64(avg) / float64(gs)
	b = float64(avg) / float64(bs)

	return
}
