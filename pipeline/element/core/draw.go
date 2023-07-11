package core

import (
	"image"

	"github.com/frizinak/phodo/img48"
)

func Draw(p image.Point, src, dst *img48.Img, trans func(r, g, b uint16) bool) {
	sr := src.Rect
	if d := sr.Dx() - dst.Rect.Dx() + p.X; d > 0 {
		sr.Max.X -= d
	}
	if d := sr.Dy() - dst.Rect.Dy() + p.Y; d > 0 {
		sr.Max.Y -= d
	}

	for y := sr.Min.Y; y < sr.Max.Y; y++ {
		so_ := (y - sr.Min.Y) * src.Stride
		do_ := (y + p.Y - sr.Min.Y) * dst.Stride
		for x := sr.Min.X; x < sr.Max.X; x++ {
			so := so_ + (x-sr.Min.X)*3
			do := do_ + (x+p.X-sr.Min.X)*3
			p := src.Pix[so : so+3 : so+3]
			if do < len(dst.Pix) && (trans == nil ||
				!trans(p[0], p[1], p[2])) {
				copy(dst.Pix[do:do+3:do+3], p)
			}
		}
	}
}
