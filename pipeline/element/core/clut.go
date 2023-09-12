package core

import (
	"fmt"
	"math"

	"github.com/frizinak/phodo/img48"
)

//go:generate go run tools/generate_clut.go 4 8 12 16

func CLUT(img, clut *img48.Img, strength float64, interpolated bool) error {
	lvl := int(math.Round(math.Pow(float64(clut.Rect.Dx()*clut.Rect.Dy()), 1.0/6)))
	var f func(_, _ *img48.Img, _ float64)
	switch {
	case lvl == 4 && interpolated:
		f = CLUT4i
	case lvl == 4:
		f = CLUT4

	case lvl == 8 && interpolated:
		f = CLUT8i
	case lvl == 8:
		f = CLUT8

	case lvl == 12 && interpolated:
		f = CLUT12i
	case lvl == 12:
		f = CLUT12

	case lvl == 16 && interpolated:
		f = CLUT16i
	case lvl == 16:
		f = CLUT16
	default:
		return fmt.Errorf("unsupported clut (level: %d, size: %dx%d)", lvl, clut.Rect.Dx(), clut.Rect.Dy())
	}

	f(img, clut, strength)

	return nil
}

type rgbInterpolate struct {
	r, g, b int
}

func (c rgbInterpolate) Mul(f16 int) rgbInterpolate {
	c.r = (c.r * f16) >> 16
	c.g = (c.g * f16) >> 16
	c.b = (c.b * f16) >> 16
	return c
}

func (c rgbInterpolate) Add(k rgbInterpolate) rgbInterpolate {
	c.r = c.r + k.r
	c.g = c.g + k.g
	c.b = c.b + k.b
	return c
}

func (c rgbInterpolate) Interpolate(k rgbInterpolate, f16, n16 int) rgbInterpolate {
	c.r = (c.r*n16 + k.r*f16) >> 16
	c.g = (c.g*n16 + k.g*f16) >> 16
	c.b = (c.b*n16 + k.b*f16) >> 16
	return c
}

func (c rgbInterpolate) Apply(vals []uint16, strength float64) {
	if strength == 1 {
		vals[0], vals[1], vals[2] = intClampUint16(c.r), intClampUint16(c.g), intClampUint16(c.b)
	} else if strength != 0 {
		nstrength := 1 - strength
		vals[0] = floatClampUint16(float64(vals[0])*nstrength + float64(c.r)*strength)
		vals[1] = floatClampUint16(float64(vals[1])*nstrength + float64(c.g)*strength)
		vals[2] = floatClampUint16(float64(vals[2])*nstrength + float64(c.b)*strength)
	}
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
