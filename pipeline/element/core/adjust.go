package core

import (
	"math"

	"github.com/frizinak/phodo/img48"
)

func Contrast(img *img48.Img, n float64) {
	l := make([]uint16, 1<<16)
	v := n + 1
	if v < 0 {
		v = 0
	}
	if v > 2 {
		v = 2
	}
	const mul = 1 << 17
	div := int(mul / (1 - math.Abs(n)))

	const half = 1<<15 - 1
	switch {
	case 0 <= v && v <= 1:
		for i := 0; i < 1<<16; i++ {
			l[i] = uint16(half + (i-half)*mul/div)
		}
	case 1 < v && v < 2:
		for i := 0; i < 1<<16; i++ {
			x := (half + (i-half)*div/mul)
			if x > 1<<16-1 {
				x = 1<<16 - 1
			}
			if x < 0 {
				x = 0
			}
			l[i] = uint16(x)
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
	shift := int((1<<16 - 1) * n)
	for i := 0; i < 1<<16; i++ {
		f := i + shift
		if f > 1<<16-1 {
			f = 1<<16 - 1
		}
		if f < 0 {
			f = 0
		}
		l[i] = uint16(f)
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

func RGBMultiply(img *img48.Img, r, g, b float64) {
	sum := r + g + b
	f := 3.0 / sum
	r *= f
	g *= f
	b *= f

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		o_ := (y - img.Rect.Min.Y) * img.Stride
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			o := o_ + (x-img.Rect.Min.X)*3
			pix := img.Pix[o : o+3 : o+3]
			pix[0] = mul(pix[0], r)
			pix[1] = mul(pix[1], g)
			pix[2] = mul(pix[2], b)
		}
	}
}

func RGBAdd(img *img48.Img, r, g, b int) {
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		o_ := (y - img.Rect.Min.Y) * img.Stride
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			o := o_ + (x-img.Rect.Min.X)*3
			pix := img.Pix[o : o+3 : o+3]
			pix[0] = add(pix[0], r)
			pix[1] = add(pix[1], g)
			pix[2] = add(pix[2], b)
		}
	}
}

func Saturation(img *img48.Img, n float64) {
	factor := int(n * (1<<16 - 1))
	c := func(v uint16, avg int) uint16 {
		r := avg + (int(v)-avg)*factor/(1<<16-1)
		if r < 0 {
			return 0
		} else if r > 1<<16-1 {
			return 1<<16 - 1
		}
		return uint16(r)
	}

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		o_ := (y - img.Rect.Min.Y) * img.Stride
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			o := o_ + (x-img.Rect.Min.X)*3
			pix := img.Pix[o : o+3 : o+3]
			avg := (int(pix[0]) + int(pix[1]) + int(pix[2])) / 3
			pix[0] = c(pix[0], avg)
			pix[1] = c(pix[1], avg)
			pix[2] = c(pix[2], avg)
		}
	}
}

func Black(img *img48.Img, n float64) {
	l := make([]uint16, 1<<16)
	const m = 1<<16 - 1

	start := n * m
	if start < 0 {
		start = 0
	}

	rng := m - start
	for i := int(start); i <= m; i++ {
		l[i] = uint16((float64(i) - start) * m / rng)
	}

	LUT16(img, l)
}
