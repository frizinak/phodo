package core

import (
	"github.com/frizinak/phodo/img48"
)

func LUT8(img *img48.Img, lut []uint8) {
	l := img.Rect.Dx() * 3
	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			r := uint(pix[o+0])
			g := uint(pix[o+1])
			b := uint(pix[o+2])
			pix[o+0] = uint16(lut[r+0])<<8 | uint16(lut[1<<16+r])
			pix[o+1] = uint16(lut[g+0])<<8 | uint16(lut[1<<16+g])
			pix[o+2] = uint16(lut[b+0])<<8 | uint16(lut[1<<16+b])
		}
	})
}

func LUT16(img *img48.Img, lut []uint16) {
	l := img.Rect.Dx() * 3
	P48(img, func(pix []uint16, _ int) {
		for o := 0; o < l; o += 3 {
			pix[o+0] = lut[pix[o+0]]
			pix[o+1] = lut[pix[o+1]]
			pix[o+2] = lut[pix[o+2]]
		}
	})
}

func LUT16Y(img *img48.Img, lut []uint16) {
	pix := ycbcr(img)

	l := img.Rect.Dx() * 3

	P48y(img, func(offset, _ int) {
		for o_ := 0; o_ < l; o_ += 3 {
			o := offset + o_
			yy := int(lut[pix[o+0]>>16]) << 16
			cb := pix[o+1]
			cr := pix[o+2]

			r := 91881 * cr
			g := -22554*cb - 46802*cr
			b := 116130 * cb

			img.Pix[o+0] = intClampUint16((yy + r) >> 16)
			img.Pix[o+1] = intClampUint16((yy + g) >> 16)
			img.Pix[o+2] = intClampUint16((yy + b) >> 16)
		}
	})
}
