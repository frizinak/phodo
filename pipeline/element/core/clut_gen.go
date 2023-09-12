// Generated, do not edit!
package core

import "github.com/frizinak/phodo/img48"

func CLUT4(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]
			r := pix[0] / 4096
			g := pix[1] / 4096
			b := pix[2] / 4096

			hx := int(r%16 + (g%4)*16)
			hy := int(b*4 + g/4)
			v := hy*192 + hx*3
			ipol(pix, clut.Pix[v:v+3:v+3], strength)
		}
	})
}

func CLUT4i(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]

			rr, gg, bb := float64(pix[0])*0.000244140625, float64(pix[1])*0.000244140625, float64(pix[2])*0.000244140625
			r0, g0, b0 := int(rr), int(gg), int(bb)

			x := int((1<<16 - 1) * (rr - float64(r0)))
			y := int((1<<16 - 1) * (gg - float64(g0)))
			z := int((1<<16 - 1) * (bb - float64(b0)))
			xi, yi, zi := (1<<16-1)-x, (1<<16-1)-y, (1<<16-1)-z

			r0v := r0 % 16
			r1v := r0v
			if r0 < 15 {
				r1v = (r0 + 1) % 16
			}

			b0v := b0 * 4
			b1v := b0v
			if b0 < 15 {
				b1v = (b0 + 1) * 4
			}

			g0v0 := g0 / 4
			g1v0 := g0v0
			g0v1 := (g0 % 4) * 16
			g1v1 := g0v1
			if g0 < 15 {
				g1v0 = (g0 + 1) / 4
				g1v1 = ((g0 + 1) % 4) * 16
			}

			r0vg0v1 := (r0v + g0v1) * 3
			r0vg1v1 := (r0v + g1v1) * 3
			r1vg0v1 := (r1v + g0v1) * 3
			r1vg1v1 := (r1v + g1v1) * 3
			b0vg0v0 := (b0v + g0v0) * 192
			b0vg1v0 := (b0v + g1v0) * 192
			b1vg0v0 := (b1v + g0v0) * 192
			b1vg1v0 := (b1v + g1v0) * 192

			v := b0vg0v0 + r0vg0v1
			c000 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r0vg1v1
			c010 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r0vg0v1
			c001 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r0vg1v1
			c011 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r1vg1v1
			c110 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r1vg1v1
			c111 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg0v0 + r1vg0v1
			c100 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r1vg0v1
			c101 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			c00 := c000.Interpolate(c100, x, xi)
			c01 := c001.Interpolate(c101, x, xi)
			c10 := c010.Interpolate(c110, x, xi)
			c11 := c011.Interpolate(c111, x, xi)

			c0 := c00.Interpolate(c10, y, yi)
			c1 := c01.Interpolate(c11, y, yi)

			c0.Interpolate(c1, z, zi).Apply(pix, strength)
		}
	})
}

func CLUT8(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]
			r := pix[0] / 1024
			g := pix[1] / 1024
			b := pix[2] / 1024

			hx := int(r%64 + (g%8)*64)
			hy := int(b*8 + g/8)
			v := hy*1536 + hx*3
			ipol(pix, clut.Pix[v:v+3:v+3], strength)
		}
	})
}

func CLUT8i(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]

			rr, gg, bb := float64(pix[0])*0.0009765625, float64(pix[1])*0.0009765625, float64(pix[2])*0.0009765625
			r0, g0, b0 := int(rr), int(gg), int(bb)

			x := int((1<<16 - 1) * (rr - float64(r0)))
			y := int((1<<16 - 1) * (gg - float64(g0)))
			z := int((1<<16 - 1) * (bb - float64(b0)))
			xi, yi, zi := (1<<16-1)-x, (1<<16-1)-y, (1<<16-1)-z

			r0v := r0 % 64
			r1v := r0v
			if r0 < 63 {
				r1v = (r0 + 1) % 64
			}

			b0v := b0 * 8
			b1v := b0v
			if b0 < 63 {
				b1v = (b0 + 1) * 8
			}

			g0v0 := g0 / 8
			g1v0 := g0v0
			g0v1 := (g0 % 8) * 64
			g1v1 := g0v1
			if g0 < 63 {
				g1v0 = (g0 + 1) / 8
				g1v1 = ((g0 + 1) % 8) * 64
			}

			r0vg0v1 := (r0v + g0v1) * 3
			r0vg1v1 := (r0v + g1v1) * 3
			r1vg0v1 := (r1v + g0v1) * 3
			r1vg1v1 := (r1v + g1v1) * 3
			b0vg0v0 := (b0v + g0v0) * 1536
			b0vg1v0 := (b0v + g1v0) * 1536
			b1vg0v0 := (b1v + g0v0) * 1536
			b1vg1v0 := (b1v + g1v0) * 1536

			v := b0vg0v0 + r0vg0v1
			c000 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r0vg1v1
			c010 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r0vg0v1
			c001 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r0vg1v1
			c011 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r1vg1v1
			c110 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r1vg1v1
			c111 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg0v0 + r1vg0v1
			c100 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r1vg0v1
			c101 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			c00 := c000.Interpolate(c100, x, xi)
			c01 := c001.Interpolate(c101, x, xi)
			c10 := c010.Interpolate(c110, x, xi)
			c11 := c011.Interpolate(c111, x, xi)

			c0 := c00.Interpolate(c10, y, yi)
			c1 := c01.Interpolate(c11, y, yi)

			c0.Interpolate(c1, z, zi).Apply(pix, strength)
		}
	})
}

func CLUT12(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]
			r := pix[0] / 456
			g := pix[1] / 456
			b := pix[2] / 456

			hx := int(r%144 + (g%12)*144)
			hy := int(b*12 + g/12)
			v := hy*5184 + hx*3
			ipol(pix, clut.Pix[v:v+3:v+3], strength)
		}
	})
}

func CLUT12i(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]

			rr, gg, bb := float64(pix[0])*0.0021929824561403508, float64(pix[1])*0.0021929824561403508, float64(pix[2])*0.0021929824561403508
			r0, g0, b0 := int(rr), int(gg), int(bb)

			x := int((1<<16 - 1) * (rr - float64(r0)))
			y := int((1<<16 - 1) * (gg - float64(g0)))
			z := int((1<<16 - 1) * (bb - float64(b0)))
			xi, yi, zi := (1<<16-1)-x, (1<<16-1)-y, (1<<16-1)-z

			r0v := r0 % 144
			r1v := r0v
			if r0 < 143 {
				r1v = (r0 + 1) % 144
			}

			b0v := b0 * 12
			b1v := b0v
			if b0 < 143 {
				b1v = (b0 + 1) * 12
			}

			g0v0 := g0 / 12
			g1v0 := g0v0
			g0v1 := (g0 % 12) * 144
			g1v1 := g0v1
			if g0 < 143 {
				g1v0 = (g0 + 1) / 12
				g1v1 = ((g0 + 1) % 12) * 144
			}

			r0vg0v1 := (r0v + g0v1) * 3
			r0vg1v1 := (r0v + g1v1) * 3
			r1vg0v1 := (r1v + g0v1) * 3
			r1vg1v1 := (r1v + g1v1) * 3
			b0vg0v0 := (b0v + g0v0) * 5184
			b0vg1v0 := (b0v + g1v0) * 5184
			b1vg0v0 := (b1v + g0v0) * 5184
			b1vg1v0 := (b1v + g1v0) * 5184

			v := b0vg0v0 + r0vg0v1
			c000 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r0vg1v1
			c010 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r0vg0v1
			c001 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r0vg1v1
			c011 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r1vg1v1
			c110 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r1vg1v1
			c111 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg0v0 + r1vg0v1
			c100 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r1vg0v1
			c101 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			c00 := c000.Interpolate(c100, x, xi)
			c01 := c001.Interpolate(c101, x, xi)
			c10 := c010.Interpolate(c110, x, xi)
			c11 := c011.Interpolate(c111, x, xi)

			c0 := c00.Interpolate(c10, y, yi)
			c1 := c01.Interpolate(c11, y, yi)

			c0.Interpolate(c1, z, zi).Apply(pix, strength)
		}
	})
}

func CLUT16(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]
			r := pix[0] / 256
			g := pix[1] / 256
			b := pix[2] / 256

			hx := int(r%256 + (g%16)*256)
			hy := int(b*16 + g/16)
			v := hy*12288 + hx*3
			ipol(pix, clut.Pix[v:v+3:v+3], strength)
		}
	})
}

func CLUT16i(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]

			rr, gg, bb := float64(pix[0])*0.00390625, float64(pix[1])*0.00390625, float64(pix[2])*0.00390625
			r0, g0, b0 := int(rr), int(gg), int(bb)

			x := int((1<<16 - 1) * (rr - float64(r0)))
			y := int((1<<16 - 1) * (gg - float64(g0)))
			z := int((1<<16 - 1) * (bb - float64(b0)))
			xi, yi, zi := (1<<16-1)-x, (1<<16-1)-y, (1<<16-1)-z

			r0v := r0 % 256
			r1v := r0v
			if r0 < 255 {
				r1v = (r0 + 1) % 256
			}

			b0v := b0 * 16
			b1v := b0v
			if b0 < 255 {
				b1v = (b0 + 1) * 16
			}

			g0v0 := g0 / 16
			g1v0 := g0v0
			g0v1 := (g0 % 16) * 256
			g1v1 := g0v1
			if g0 < 255 {
				g1v0 = (g0 + 1) / 16
				g1v1 = ((g0 + 1) % 16) * 256
			}

			r0vg0v1 := (r0v + g0v1) * 3
			r0vg1v1 := (r0v + g1v1) * 3
			r1vg0v1 := (r1v + g0v1) * 3
			r1vg1v1 := (r1v + g1v1) * 3
			b0vg0v0 := (b0v + g0v0) * 12288
			b0vg1v0 := (b0v + g1v0) * 12288
			b1vg0v0 := (b1v + g0v0) * 12288
			b1vg1v0 := (b1v + g1v0) * 12288

			v := b0vg0v0 + r0vg0v1
			c000 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r0vg1v1
			c010 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r0vg0v1
			c001 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r0vg1v1
			c011 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r1vg1v1
			c110 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r1vg1v1
			c111 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg0v0 + r1vg0v1
			c100 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r1vg0v1
			c101 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			c00 := c000.Interpolate(c100, x, xi)
			c01 := c001.Interpolate(c101, x, xi)
			c10 := c010.Interpolate(c110, x, xi)
			c11 := c011.Interpolate(c111, x, xi)

			c0 := c00.Interpolate(c10, y, yi)
			c1 := c01.Interpolate(c11, y, yi)

			c0.Interpolate(c1, z, zi).Apply(pix, strength)
		}
	})
}
