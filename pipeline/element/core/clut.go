package core

import (
	"fmt"
	"math"

	"github.com/frizinak/phodo/img48"
)

func CLUT(img, clut *img48.Img, strength float64, iterations int) error {
	lvl := uint16(math.Round(math.Pow(float64(clut.Rect.Dx()*clut.Rect.Dy()), 1.0/6)))
	switch lvl {
	case 4:
		for i := 0; i < iterations; i++ {
			CLUT4(img, clut, strength)
		}
	case 8:
		for i := 0; i < iterations; i++ {
			CLUT8(img, clut, strength)
		}
	case 12:
		for i := 0; i < iterations; i++ {
			CLUT12(img, clut, strength)
		}
	case 16:
		for i := 0; i < iterations; i++ {
			CLUT16(img, clut, strength)
		}
	default:
		return fmt.Errorf("unsupported clut (level: %d, size: %dx%d)", lvl, clut.Rect.Dx(), clut.Rect.Dy())
	}

	return nil
}

func CLUT4(img, clut *img48.Img, strength float64) {
	l := img.Stride // major performance impact
	p48(img, func(rpix []uint16, _ int) {
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

func CLUT8(img, clut *img48.Img, strength float64) {
	l := img.Stride // major performance impact
	p48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]
			r := pix[0] / 1024
			g := pix[1] / 1024
			b := pix[2] / 1024

			hx := int((r % 64) + (g%8)*64)
			hy := int(b*8 + g/8)
			v := hy*1536 + hx*3
			ipol(pix, clut.Pix[v:v+3:v+3], strength)
		}
	})
}

func CLUT12(img, clut *img48.Img, strength float64) {
	l := img.Stride // major performance impact
	p48(img, func(rpix []uint16, _ int) {
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

func CLUT16(img, clut *img48.Img, strength float64) {
	l := img.Stride // major performance impact
	p48(img, func(rpix []uint16, _ int) {
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

func ipol(dst, src []uint16, strength float64) {
	if strength == 1 {
		copy(dst, src)
	} else if strength != 0 {
		nstrength := 1 - strength - 1e-12
		dst[0] = uint16(float64(dst[0])*nstrength + float64(src[0])*strength)
		dst[1] = uint16(float64(dst[1])*nstrength + float64(src[1])*strength)
		dst[2] = uint16(float64(dst[2])*nstrength + float64(src[2])*strength)
	}
}
