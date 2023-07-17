package core

import (
	"image"

	"github.com/frizinak/phodo/img48"
)

type Color interface {
	Color() [3]uint16
}

func Draw(p image.Point, src, dst *img48.Img, trans func(r, g, b uint16) bool) {
	sr := src.Rect
	if d := sr.Dx() - dst.Rect.Dx() + p.X; d > 0 {
		sr.Max.X -= d
	}
	if d := sr.Dy() - dst.Rect.Dy() + p.Y; d > 0 {
		sr.Max.Y -= d
	}

	for y := sr.Min.Y; y < sr.Max.Y; y++ {
		so_ := (y - sr.Min.Y) * src.Stride
		do_ := (y + p.Y - sr.Min.Y) * dst.Stride
		for x := sr.Min.X; x < sr.Max.X; x++ {
			so := so_ + (x-sr.Min.X)*3
			do := do_ + (x+p.X-sr.Min.X)*3
			if do < 0 {
				continue
			}
			p := src.Pix[so : so+3 : so+3]
			if do < len(dst.Pix) && (trans == nil ||
				!trans(p[0], p[1], p[2])) {
				copy(dst.Pix[do:do+3:do+3], p)
			}
		}
	}
}

func DrawFilledCircle(p image.Point, radius int, src Color, dst *img48.Img) {
	_clr := src.Color()
	clr := _clr[:]
	o := func(x1, x2, y int) {
		if y >= dst.Rect.Max.Y {
			y = dst.Rect.Max.Y - 1
		}
		if x1 >= dst.Rect.Max.X {
			x1 = dst.Rect.Max.X - 1
		}
		if x2 >= dst.Rect.Max.X {
			x2 = dst.Rect.Max.X - 1
		}

		x1 = (x1 - dst.Rect.Min.X)
		x2 = (x2 - dst.Rect.Min.X)
		y = y - dst.Rect.Min.Y
		if y < 0 {
			y = 0
		}
		if x1 < 0 {
			x1 = 0
		}
		if x2 < 0 {
			x2 = 0
		}

		o := y * dst.Stride
		o1 := o + x1*3
		o2 := o + x2*3
		pix := dst.Pix[o1 : o2+3 : o2+3]
		for n := 0; n < len(pix); n += 3 {
			copy(pix[n:n+3:n+3], clr)
		}
	}

	cx, cy := p.X, p.Y
	x := radius
	y := 0
	e := 0

	for x >= y {
		o(cx-x, cx+x, cy+y)
		o(cx-x, cx+x, cy-y)
		o(cx-y, cx+y, cy+x)
		o(cx-y, cx+y, cy-x)
		if e <= 0 {
			y += 1
			e += 2*y + 1
		}

		if e > 0 {
			x -= 1
			e -= 2*x + 1
		}
	}
}

func DrawCircleBorder(p image.Point, radius, border int, src Color, dst *img48.Img) {
	if border >= radius {
		DrawFilledCircle(p, radius, src, dst)
		return
	}

	for n := radius - border; n <= radius; n++ {
		DrawCircle(p, n, src, dst)
	}
}

func DrawCircle(p image.Point, radius int, src Color, dst *img48.Img) {
	_clr := src.Color()
	clr := _clr[:]
	o := func(x, y int) {
		if y >= dst.Rect.Max.Y {
			y = dst.Rect.Max.Y - 1
		}
		if x >= dst.Rect.Max.X {
			x = dst.Rect.Max.X - 1
		}

		x = (x - dst.Rect.Min.X)
		y = y - dst.Rect.Min.Y
		if y < 0 {
			y = 0
		}
		if x < 0 {
			x = 0
		}

		o_ := y * dst.Stride
		o := o_ + x*3
		copy(dst.Pix[o:o+3:o+3], clr)
	}

	cx, cy := p.X, p.Y
	x := radius
	y := 0
	e := 0

	for x >= y {
		o(cx+x, cy+y)
		o(cx+y, cy+x)
		o(cx-y, cy+x)
		o(cx-x, cy+y)
		o(cx-x, cy-y)
		o(cx-y, cy-x)
		o(cx+y, cy-x)
		o(cx+x, cy-y)

		if e <= 0 {
			y += 1
			e += 2*y + 1
		}

		if e > 0 {
			x -= 1
			e -= 2*x + 1
		}
	}
}
