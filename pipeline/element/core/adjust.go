package core

import (
	"math"

	"github.com/frizinak/phodo/img48"
)

func ContrastY(img *img48.Img, n float64) {
	n--
	if n < -1 {
		n = -1
	}
	if n > 1 {
		n = 1
	}

	l := make([]uint16, 1<<16)
	const mul = 1 << 17
	div := int(mul / (1 - math.Abs(n)))

	const half = 1<<15 - 1
	switch {
	case -1 <= n && n <= 0:
		for i := 0; i < 1<<16; i++ {
			l[i] = uint16(half + (i-half)*mul/div)
		}
	case 0 < n && n < 1:
		for i := 0; i < 1<<16; i++ {
			l[i] = intClampUint16(half + (i-half)*div/mul)
		}
	default:
		for i := half; i < 1<<16; i++ {
			l[i] = 1<<16 - 1
		}
	}

	LUT16Y(img, l)
}

func Contrast(img *img48.Img, n float64) {
	n--
	if n < -1 {
		n = -1
	}
	if n > 1 {
		n = 1
	}

	l := make([]uint16, 1<<16)
	const mul = 1 << 17
	div := int(mul / (1 - math.Abs(n)))

	const half = 1<<15 - 1
	switch {
	case -1 <= n && n <= 0:
		for i := 0; i < 1<<16; i++ {
			l[i] = uint16(half + (i-half)*mul/div)
		}
	case 0 < n && n < 1:
		for i := 0; i < 1<<16; i++ {
			l[i] = intClampUint16(half + (i-half)*div/mul)
		}
	default:
		for i := half; i < 1<<16; i++ {
			l[i] = 1<<16 - 1
		}
	}

	LUT16(img, l)
}

func Brightness(img *img48.Img, n float64) {
	l := make([]uint16, 1<<16)
	shift := int((1<<16 - 1) * (n - 1))
	for i := 0; i < 1<<16; i++ {
		l[i] = intClampUint16(i + shift)
	}

	LUT16(img, l)
}

func Gamma(img *img48.Img, n float64) {
	e := 1.0 / n
	l := make([]uint16, 1<<16)

	for i := 0; i < 1<<16; i++ {
		l[i] = uint16(math.Pow(float64(i)/(1<<16-1), e) * (1<<16 - 1))
	}

	LUT16(img, l)
}

func Eq(img *img48.Img, ns ...float64) {
	if len(ns) == 0 {
		return
	}
	if len(ns) == 1 {
		ns = []float64{ns[0], ns[0]}
	}

	l := make([]uint16, 1<<16)
	c := uint32(0xffff/(len(ns)-1) + 1)
	cf := float64(c)

	var i uint32
	for ; i < 1<<16; i++ {
		ix := i / c
		maxd := float64(i%c) / cf
		mind := 1.0 - maxd
		factor := ns[ix]*mind + ns[ix+1]*maxd
		l[i] = floatClampUint16(factor * float64(i))
	}

	LUT16(img, l)
}

func RGBMultiply(img *img48.Img, r, g, b float64, norm bool) {
	if norm {
		sum := r + g + b
		f := 3.0 / sum
		r *= f
		g *= f
		b *= f
	}

	l := img.Rect.Dx() * 3
	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			pix[o+0] = mul(pix[o+0], r)
			pix[o+1] = mul(pix[o+1], g)
			pix[o+2] = mul(pix[o+2], b)
		}
	})
}

func RGBAdd(img *img48.Img, r, g, b int) {
	l := img.Rect.Dx() * 3
	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			pix[o+0] = add(pix[o+0], r)
			pix[o+1] = add(pix[o+1], g)
			pix[o+2] = add(pix[o+2], b)
		}
	})
}

func Saturation(img *img48.Img, n float64) {
	factor := int(n * (1<<16 - 1))
	l := img.Rect.Dx() * 3
	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			r, g, b := int(pix[o+0]), int(pix[o+1]), int(pix[o+2])
			avg := (r + g + b) / 3
			pix[o+0] = intClampUint16(avg + ((r-avg)*factor)>>16)
			pix[o+1] = intClampUint16(avg + ((g-avg)*factor)>>16)
			pix[o+2] = intClampUint16(avg + ((b-avg)*factor)>>16)
		}
	})
}

func Black(img *img48.Img, n float64) {
	l := make([]uint16, 1<<16)
	const m = 1<<16 - 1
	start := (n - 1) * m
	if start < 0 {
		start = 0
	}
	rng := m - start
	for i := int(start); i <= m; i++ {
		l[i] = uint16((float64(i) - start) * m / rng)
	}

	LUT16(img, l)
}

func Invert(img *img48.Img, r, g, b float64) {
	l := img.Rect.Dx() * 3
	or := uint32((1<<16 - 1) * r)
	og := uint32((1<<16 - 1) * g)
	ob := uint32((1<<16 - 1) * b)
	nr := 1<<16 - 1 - or
	ng := 1<<16 - 1 - og
	nb := 1<<16 - 1 - ob

	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			r, g, b := uint32(pix[o+0]), uint32(pix[o+1]), uint32(pix[o+2])
			pix[o+0] = uint16((or*((1<<16-1)-r) + nr*r) >> 16)
			pix[o+1] = uint16((og*((1<<16-1)-g) + ng*g) >> 16)
			pix[o+2] = uint16((ob*((1<<16-1)-b) + nb*b) >> 16)
		}
	})
}

func InvertFilm(img *img48.Img, r, g, b float64) {
	exp := func(v uint16, a float64) uint16 {
		return floatClampUint16(0xffff * math.Pow(1+float64(v)/0xffff, a))
	}

	l := img.Rect.Dx() * 3
	re, ge, be := -(g * r), -g, -(g * b)
	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			pix[o+0] = exp(pix[o+0], re)
			pix[o+1] = exp(pix[o+1], ge)
			pix[o+2] = exp(pix[o+2], be)
		}
	})
}
