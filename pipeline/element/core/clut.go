package core

import (
	"fmt"
	"math"

	"github.com/frizinak/phodo/img48"
)

func CLUT(img, clut *img48.Img, amount float64) error {
	lvl := uint16(math.Round(math.Pow(float64(clut.Rect.Dx()*clut.Rect.Dy()), 1.0/6)))
	switch lvl {
	case 4:
		CLUT4(img, clut, amount)
	case 8:
		CLUT8(img, clut, amount)
	case 12:
		CLUT12(img, clut, amount)
	case 16:
		CLUT16(img, clut, amount)
	default:
		return fmt.Errorf("unsupported clut (level: %d, size: %dx%d)", lvl, clut.Rect.Dx(), clut.Rect.Dy())
	}

	return nil
}

func CLUT4(img, clut *img48.Img, amount float64) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 4096
		g := pix[1] / 4096
		b := pix[2] / 4096

		hx := int(r%16 + (g%4)*16)
		hy := int(b*4 + g/4)
		v := hy*192 + hx*3
		ipol(pix, clut.Pix[v:v+3:v+3], amount)
	}
}

func CLUT8(img, clut *img48.Img, amount float64) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 1024
		g := pix[1] / 1024
		b := pix[2] / 1024

		hx := int((r % 64) + (g%8)*64)
		hy := int(b*8 + g/8)
		v := hy*1536 + hx*3
		ipol(pix, clut.Pix[v:v+3:v+3], amount)
	}
}

func CLUT12(img, clut *img48.Img, amount float64) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 456
		g := pix[1] / 456
		b := pix[2] / 456

		hx := int(r%144 + (g%12)*144)
		hy := int(b*12 + g/12)
		v := hy*5184 + hx*3
		ipol(pix, clut.Pix[v:v+3:v+3], amount)
	}
}

func CLUT16(img, clut *img48.Img, amount float64) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 256
		g := pix[1] / 256
		b := pix[2] / 256

		hx := int(r%256 + (g%16)*256)
		hy := int(b*16 + g/16)
		v := hy*12288 + hx*3
		ipol(pix, clut.Pix[v:v+3:v+3], amount)
	}
}

func ipol(dst, src []uint16, am float64) {
	if am == 1 {
		copy(dst, src)
	} else if am != 0 {
		nam := 1 - am - 1e-12
		dst[0] = uint16(float64(dst[0])*nam + float64(src[0])*am)
		dst[1] = uint16(float64(dst[1])*nam + float64(src[1])*am)
		dst[2] = uint16(float64(dst[2])*nam + float64(src[2])*am)
	}
}
