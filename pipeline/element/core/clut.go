package core

import (
	"fmt"
	"math"

	"github.com/frizinak/phodo/img48"
)

func CLUT(img, clut *img48.Img) error {
	lvl := uint16(math.Round(math.Pow(float64(clut.Rect.Dx()*clut.Rect.Dy()), 1.0/6)))
	switch lvl {
	case 4:
		CLUT4(img, clut)
	case 8:
		CLUT8(img, clut)
	case 12:
		CLUT12(img, clut)
	case 16:
		CLUT16(img, clut)
	default:
		return fmt.Errorf("unsupported clut (level: %d, size: %dx%d)", lvl, clut.Rect.Dx(), clut.Rect.Dy())
	}

	return nil
}

func CLUTSlow(img, clut *img48.Img, amount float64) error {
	lvl := uint16(math.Round(math.Pow(float64(clut.Rect.Dx()*clut.Rect.Dy()), 1.0/6)))
	switch lvl {
	// TODO
	case 12:
		CLUT12Slow(img, clut, amount)
	default:
		return fmt.Errorf("unsupported clut (level: %d, size: %dx%d)", lvl, clut.Rect.Dx(), clut.Rect.Dy())
	}

	return nil
}

func CLUT4(img, clut *img48.Img) {
	// 4 = level
	// 16 = level * level
	// 4096 = 256*256 / (level*level)
	// 512 = clut.Stride

	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 4096
		g := pix[1] / 4096
		b := pix[2] / 4096

		hx := int(r%16 + (g%4)*16)
		hy := int(b*4 + g/4)
		v := hy*192 + hx*3
		copy(pix, clut.Pix[v:v+3:v+3])
	}
}

func CLUT8(img, clut *img48.Img) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 1024
		g := pix[1] / 1024
		b := pix[2] / 1024

		hx := int((r % 64) + (g%8)*64)
		hy := int(b*8 + g/8)
		v := hy*1536 + hx*3
		copy(pix, clut.Pix[v:v+3:v+3])
	}
}

func CLUT12(img, clut *img48.Img) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 456
		g := pix[1] / 456
		b := pix[2] / 456

		hx := int(r%144 + (g%12)*144)
		hy := int(b*12 + g/12)
		v := hy*5184 + hx*3
		copy(pix, clut.Pix[v:v+3:v+3])
	}
}

func CLUT12Slow(img, clut *img48.Img, amount float64) {
	p0 := make([]uint16, 3)
	p1 := make([]uint16, 3)
	p2 := make([]uint16, 3)

	ix := func(ir, ig, ib int) int {
		if ir > 143 {
			ir = 143
		}
		if ig > 143 {
			ig = 143
		}
		if ib > 143 {
			ib = 143
		}
		return (ib*12+ig/12)*5184 + (ir+(ig%12)*144)*3
	}

	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r, g, b := float64(pix[0])/456, float64(pix[1])/456, float64(pix[2])/456
		ir, ig, ib := int(r), int(g), int(b)
		ri, gi, bi := r-float64(ir), g-float64(ig), b-float64(ib)
		_, _, _ = ri, gi, bi

		ipol2(p1, clut.Pix, ix(ir+0, ig+0, ib+0), 1)
		ipol2(p1, clut.Pix, ix(ir+1, ig+0, ib+0), ri)

		ipol2(p2, clut.Pix, ix(ir+0, ig+1, ib+0), 1)
		ipol2(p2, clut.Pix, ix(ir+1, ig+1, ib+0), ri)
		copy(p0, p1)
		ipol2(p0, p2, 0, gi)

		ipol2(p1, clut.Pix, ix(ir+0, ig+0, ib+1), 1)
		ipol2(p1, clut.Pix, ix(ir+1, ig+0, ib+1), ri)

		ipol2(p2, clut.Pix, ix(ir+0, ig+1, ib+1), 1)
		ipol2(p2, clut.Pix, ix(ir+1, ig+1, ib+1), ri)

		ipol2(p2, p1, 0, gi)
		ipol2(p0, p2, 0, bi)

		ipol2(pix, p0, 0, amount)
	}
}

func CLUT16(img, clut *img48.Img) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := pix[0] / 256
		g := pix[1] / 256
		b := pix[2] / 256

		hx := int(r%256 + (g%16)*256)
		hy := int(b*16 + g/16)
		v := hy*12288 + hx*3

		copy(pix, clut.Pix[v:v+3:v+3])
	}
}

func ipol2(dst, src []uint16, offset int, amount float64) {
	if offset < len(src) {
		if amount == 1 {
			copy(dst, src[offset:offset+3:offset+3])
			return
		} else if amount > 0 {
			ipoln(dst, src[offset:offset+3:offset+3], amount)
		}
	}
}

func ipoln(dpix, spix []uint16, amount float64) {
	a := uint32(amount * (1<<16 - 1))
	n := (1<<16 - 1) - a
	dpix[0] = uint16((uint32(dpix[0])*n + uint32(spix[0])*a) >> 16)
	dpix[1] = uint16((uint32(dpix[1])*n + uint32(spix[1])*a) >> 16)
	dpix[2] = uint16((uint32(dpix[2])*n + uint32(spix[2])*a) >> 16)
}
