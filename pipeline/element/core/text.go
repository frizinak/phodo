package core

import (
	"image"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func FontLoad(d []byte) (*sfnt.Font, error) {
	col, err := opentype.ParseCollection(d)
	if err != nil {
		return nil, err
	}

	return col.Font(0)
}

func Text(dst draw.Image, src image.Image, x, y int, size float64, text string, fnt *sfnt.Font) error {
	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return err
	}

	d := font.Drawer{
		Dst:  dst,
		Src:  src,
		Face: face,
		Dot:  fixed.P(x, y),
	}

	d.DrawString(text)
	return nil
}
