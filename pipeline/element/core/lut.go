package core

import "github.com/frizinak/phodo/img48"

func LUT8(img *img48.Img, lut []uint8) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		r := uint(pix[0])
		g := uint(pix[1])
		b := uint(pix[2])
		pix[0] = uint16(lut[r+0])<<8 | uint16(lut[1<<16+r])
		pix[1] = uint16(lut[g+0])<<8 | uint16(lut[1<<16+g])
		pix[2] = uint16(lut[b+0])<<8 | uint16(lut[1<<16+b])
	}
}

func LUT16(img *img48.Img, lut []uint16) {
	for o := 0; o < len(img.Pix); o += 3 {
		pix := img.Pix[o : o+3 : o+3]
		pix[0] = lut[pix[0]]
		pix[1] = lut[pix[1]]
		pix[2] = lut[pix[2]]
	}
}
