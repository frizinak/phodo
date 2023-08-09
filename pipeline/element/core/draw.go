package core

import (
	"image"

	"github.com/frizinak/phodo/img48"
)

type Color interface {
	Color() [3]uint16
}

type SimpleColor struct {
	R, G, B uint16
}

func (s SimpleColor) Color() [3]uint16 {
	return [3]uint16{s.R, s.G, s.B}
}

func Draw(src, dst *img48.Img, p image.Point, trans func(r, g, b uint16) bool) {
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

func DrawRectangle(src Color, dst *img48.Img, rect image.Rectangle, width int) {
	_clr := src.Color()
	clr := _clr[:]
	ll := func(x, y int) {
		if x >= dst.Rect.Min.X && y >= dst.Rect.Min.Y && x < dst.Rect.Max.X && y < dst.Rect.Max.Y {
			o := y*dst.Stride + x*3
			pix := dst.Pix[o : o+3 : o+3]
			copy(pix, clr)
		}
	}

	for x := rect.Min.X; x < rect.Max.X; x++ {
		for i := 0; i < width; i++ {
			ll(x, rect.Min.Y+i)
			ll(x, rect.Max.Y-i-1)
		}
	}

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for i := 0; i < width; i++ {
			ll(rect.Min.X+i, y)
			ll(rect.Max.X-i-1, y)
		}
	}
}

func DrawHorizontalLine(src Color, dst *img48.Img, x1, x2, y int) {
	linehorizdrawer(src, dst)(x1, x2, y)
}

func DrawFilledCircle(src Color, dst *img48.Img, p image.Point, radius int) {
	o := linehorizdrawer(src, dst)

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

func DrawCircleBorder(src Color, dst *img48.Img, p image.Point, radius, border int) {
	if border >= radius {
		DrawFilledCircle(src, dst, p, radius)
		return
	}

	for n := radius - border; n <= radius; n++ {
		DrawCircle(src, dst, p, n)
	}
}

func DrawCircle(src Color, dst *img48.Img, p image.Point, radius int) {
	o := pointdrawer(src, dst)

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

func DrawCircleSrc(src, dst *img48.Img, sp, dp image.Point, outerRadius, innerRadius int) {
	// Note: Already tried optimizing using only integer arithmetic.
	//       (both int and uint32). Causes an almost 2x slowdown.

	s := 1.0
	if innerRadius < 0 {
		s = -1.0
	}
	or2 := float64(outerRadius * outerRadius)
	ir2 := float64(innerRadius * innerRadius)
	d := 1.0 / (or2 - s*ir2)
	o := s * ir2 / 4

	sl, dl := len(src.Pix), len(dst.Pix)

	for y := -outerRadius / 2; y < +outerRadius/2; y++ {
		do_ := (dp.Y + y) * dst.Stride
		so_ := (sp.Y + y) * src.Stride
		for x := -outerRadius / 2; x < +outerRadius/2; x++ {
			do := do_ + (dp.X+x)*3
			so := so_ + (sp.X+x)*3
			if do < 0 || so < 0 || do >= dl || so >= sl {
				continue
			}

			dpix := dst.Pix[do : do+3 : do+3]
			spix := src.Pix[so : so+3 : so+3]

			g := 4 * (float64(x*x+y*y) - o) * d
			if g > 1 {
				g = 1
			}
			if g < 0 {
				g = 0
			}

			// Note: And this is faster than converting the pixels to floats.
			//       *shrug*
			dist := uint32((1<<16 - 1) * g)
			idist := (1<<16 - 1) - dist

			dpix[0] = uint16((idist*uint32(spix[0]) + dist*uint32(dpix[0])) >> 16)
			dpix[1] = uint16((idist*uint32(spix[1]) + dist*uint32(dpix[1])) >> 16)
			dpix[2] = uint16((idist*uint32(spix[2]) + dist*uint32(dpix[2])) >> 16)
		}
	}
}

func DrawClipping(src Color, dst *img48.Img, threshold float64) {
	th := floatClampUint16(threshold * (1<<16 - 1))
	_clr := src.Color()
	clr := _clr[:]

	below := th <= 1<<15-1

	check := func(v uint16) bool {
		return (!below && v >= th) || (below && v <= th)
	}

	for o := 0; o < len(dst.Pix); o += 3 {
		if check(dst.Pix[o+0]) || check(dst.Pix[o+1]) || check(dst.Pix[o+2]) {
			copy(dst.Pix[o:o+3:o+3], clr)
		}
	}
}

func linehorizcb(dst *img48.Img, cb func(y int, pix []uint16)) func(x1, x2, y int) {
	return func(x1, x2, y int) {
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
		if o2 >= len(dst.Pix) {
			return
		}
		cb(y, dst.Pix[o1:o2+3:o2+3])
	}
}

func linehorizdrawer(src Color, dst *img48.Img) func(x1, x2, y int) {
	_clr := src.Color()
	clr := _clr[:]
	cb := linehorizcb(dst, func(y int, pix []uint16) {
		for n := 0; n < len(pix); n += 3 {
			copy(pix[n:n+3:n+3], clr)
		}
	})
	return cb
}

func pointdrawer(src Color, dst *img48.Img) func(x, y int) {
	_clr := src.Color()
	clr := _clr[:]
	return func(x, y int) {
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
		if o >= len(dst.Pix) {
			return
		}
		copy(dst.Pix[o:o+3:o+3], clr)
	}
}
