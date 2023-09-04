package core

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
)

type Color interface {
	Color() (r, g, b uint16)
}

type SimpleColor struct{ R, G, B uint16 }

func (s SimpleColor) Color() (uint16, uint16, uint16) { return s.R, s.G, s.B }

type Blender func(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16)

func BlendScreen(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
	r = 0xffff - uint16((uint32(0xffff-sr)*uint32(0xffff-dr))>>16)
	g = 0xffff - uint16((uint32(0xffff-sg)*uint32(0xffff-dg))>>16)
	b = 0xffff - uint16((uint32(0xffff-sb)*uint32(0xffff-db))>>16)
	return
}

func BlendMultiply(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
	r = uint16((uint32(sr) * uint32(dr)) >> 16)
	g = uint16((uint32(sg) * uint32(dg)) >> 16)
	b = uint16((uint32(sb) * uint32(db)) >> 16)
	return
}

func BlendOverlay(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
	c := func(s, d uint16) uint16 {
		if s > 0x7fff {
			return 0xffff - uint16(uint32(0xffff-s)*uint32(0xffff-d)>>15)
		}
		return uint16((uint32(s) * uint32(d)) >> 15)
	}

	r = c(sr, dr)
	g = c(sg, dg)
	b = c(sb, db)
	return
}

func BlendDarken(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
	c := func(s, d uint16) uint16 {
		if s < d {
			return s
		}
		return d
	}

	r = c(sr, dr)
	g = c(sg, dg)
	b = c(sb, db)
	return
}

func BlendLighten(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
	c := func(s, d uint16) uint16 {
		if s > d {
			return s
		}
		return d
	}

	r = c(sr, dr)
	g = c(sg, dg)
	b = c(sb, db)
	return
}

func BlendOpacity(opacity float64) Blender {
	d := uint32((1<<16 - 1) * opacity)
	id := (1<<16 - 1) - d
	return func(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
		r = uint16((d*uint32(sr) + id*uint32(dr)) >> 16)
		g = uint16((d*uint32(sg) + id*uint32(dg)) >> 16)
		b = uint16((d*uint32(sb) + id*uint32(db)) >> 16)
		return
	}
}

func blendCopy(_, _ int, _, _ int, sr, sg, sb uint16, _, _, _ uint16) (r, g, b uint16) {
	return sr, sg, sb
}

func BlendMask(mask *img48.Img) Blender {
	w, h := mask.Rect.Dx(), mask.Rect.Dy()

	c := func(mask uint16, s, d uint16) uint16 {
		k := uint32(mask)
		l := 1<<16 - 1 - k
		return uint16((uint32(s)*k + uint32(d)*l) >> 16)
	}

	return func(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
		if sx < 0 || sy < 0 || sx >= w || sy >= h {
			return dr, dg, db
		}
		o := sy*mask.Stride + sx*3
		msk := mask.Pix[o : o+3 : o+3]

		r = c(msk[0], sr, dr)
		g = c(msk[1], sg, dg)
		b = c(msk[2], sb, db)

		return
	}
}

func BlendKey(key Color, fuzz float64) Blender {
	_kr, _kg, _kb := key.Color()
	kr, kg, kb := int(_kr), int(_kg), int(_kb)
	f := int(fuzz * (1<<16 - 1))
	krMin, krMax := intClampUint16(kr-f), intClampUint16(kr+f)
	kgMin, kgMax := intClampUint16(kg-f), intClampUint16(kg+f)
	kbMin, kbMax := intClampUint16(kb-f), intClampUint16(kb+f)
	return func(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
		r, g, b = sr, sg, sb
		if r >= krMin && r <= krMax && g >= kgMin && g <= kgMax && b >= kbMin && b <= kbMax {
			r, g, b = dr, dg, db
		}

		return
	}
}

func Blend(blend ...Blender) Blender {
	return func(sx, sy int, dx, dy int, sr, sg, sb uint16, dr, dg, db uint16) (r, g, b uint16) {
		r, g, b = sr, sg, sb
		for _, bl := range blend {
			r, g, b = bl(sx, sy, dx, dy, r, g, b, dr, dg, db)
		}
		return
	}
}

func Draw(src, dst *img48.Img, p image.Point, blender Blender) {
	sr := src.Rect
	if d := sr.Dx() - dst.Rect.Dx() + p.X; d > 0 {
		sr.Max.X -= d
	}
	if d := sr.Dy() - dst.Rect.Dy() + p.Y; d > 0 {
		sr.Max.Y -= d
	}

	if blender == nil {
		blender = blendCopy
	}

	w := sr.Dx()
	dw, dh := dst.Rect.Dx(), dst.Rect.Dy()
	P48(src, func(pix []uint16, y int) {
		ny := y + p.Y
		if ny < 0 {
			return
		}
		if ny >= dh {
			return
		}

		do_ := ny * dst.Stride
		for x := 0; x < w; x++ {
			nx := x + p.X
			if nx < 0 {
				continue
			}
			if nx >= dw {
				break
			}

			so := x * 3
			do := do_ + nx*3
			spix := pix[so : so+3 : so+3]
			dpix := dst.Pix[do : do+3 : do+3]
			dpix[0], dpix[1], dpix[2] = blender(
				x, y,
				nx, ny,
				spix[0], spix[1], spix[2],
				dpix[0], dpix[1], dpix[2],
			)
		}
	})
}

func DrawRectangle(src Color, dst *img48.Img, rect image.Rectangle, width int) {
	r, g, b := src.Color()
	clr := []uint16{r, g, b}
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	ll := func(x, y int) {
		if x >= 0 && y >= 0 && x < w && y < h {
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

	sl := len(src.Pix)

	for y := -outerRadius / 2; y < outerRadius/2; y++ {
		ny := dp.Y + y - dst.Rect.Min.Y
		if ny < dst.Rect.Min.Y {
			continue
		}
		if ny >= dst.Rect.Max.Y {
			break
		}
		do_ := ny * dst.Stride
		so_ := (sp.Y + y) * src.Stride
		for x := -outerRadius / 2; x < outerRadius/2; x++ {
			nx := dp.X + x - dst.Rect.Min.X
			if nx < dst.Rect.Min.X {
				continue
			}
			if nx >= dst.Rect.Max.X {
				break
			}

			do := do_ + nx*3
			so := so_ + (sp.X+x)*3
			if so < 0 || so >= sl {
				continue
			}

			if do > len(dst.Pix) {
				fmt.Println(nx, ny, dst.Rect)
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

func DrawClipping(src Color, dst *img48.Img, threshold float64, singleChannel bool) {
	th := floatClampUint16(threshold * (1<<16 - 1))
	r, g, b := src.Color()
	clr := []uint16{r, g, b}

	below := th <= 1<<15-1

	var check func(r, g, b uint16) bool
	switch {
	case singleChannel && below:
		check = func(r, g, b uint16) bool {
			return r <= th || g <= th || b <= th
		}
	case singleChannel && !below:
		check = func(r, g, b uint16) bool {
			return r >= th || g >= th || b >= th
		}
	case !singleChannel && below:
		check = func(r, g, b uint16) bool {
			return r/3+g/3+b/3 <= th
		}
	case !singleChannel && !below:
		check = func(r, g, b uint16) bool {
			return r/3+g/3+b/3 >= th
		}
	}

	l := dst.Rect.Dx() * 3
	P48(dst, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			if check(dst.Pix[o+0], dst.Pix[o+1], dst.Pix[o+2]) {
				copy(dst.Pix[o:o+3:o+3], clr)
			}
		}
	})
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
	r, g, b := src.Color()
	clr := []uint16{r, g, b}

	cb := linehorizcb(dst, func(y int, pix []uint16) {
		for n := 0; n < len(pix); n += 3 {
			copy(pix[n:n+3:n+3], clr)
		}
	})

	return cb
}

func pointdrawer(src Color, dst *img48.Img) func(x, y int) {
	r, g, b := src.Color()
	clr := []uint16{r, g, b}

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
