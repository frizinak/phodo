package element

import (
	"encoding/hex"
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func RGB16(r, g, b uint16) clrRGB16 { return clrRGB16{[3]uint16{r, g, b}} }

func RGB8(r, g, b uint8) clrRGB16 {
	return clrRGB16{[3]uint16{uint16(r) * 257, uint16(g) * 257, uint16(b) * 257}}
}

func Hex(str string) (clrRGB16, error) {
	if len(str) == 0 {
		return clrRGB16{}, nil
	}
	if str[0] == '#' {
		return Hex(str[1:])
	}

	switch len(str) {
	case 3:
		h := make([]byte, 6)
		h[0], h[1] = str[0], str[0]
		h[2], h[3] = str[1], str[1]
		h[4], h[5] = str[2], str[2]
		str = string(h)
		fallthrough
	case 6:
		d, err := hex.DecodeString(str)
		if err != nil {
			return clrRGB16{}, err
		}
		return RGB8(d[0], d[1], d[2]), nil
	case 12:
		d, err := hex.DecodeString(str)
		if err != nil {
			return clrRGB16{}, err
		}
		return RGB16(
			uint16(d[0])<<8|uint16(d[1]),
			uint16(d[2])<<8|uint16(d[3]),
			uint16(d[4])<<8|uint16(d[5]),
		), nil
	default:
		return clrRGB16{}, fmt.Errorf("'%s' is not a valid color", str)
	}
}

type Color interface {
	Color() [3]uint16
	pipeline.Element
}

type clrHex struct {
	str string
	clrRGB16
}

func (clrHex) Name() string { return "hex" }
func (clrHex) Inline() bool { return true }

func (hex clrHex) Encode(w pipeline.Writer) error {
	w.String(hex.str)
	return nil
}

func (hex clrHex) Decode(r pipeline.Reader) (pipeline.Element, error) {
	hex.str = r.String()
	var err error
	hex.clrRGB16, err = Hex(hex.str)
	return hex, err
}

func (hex clrHex) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<hex>)", hex.Name()),
			"Creates a color. <hex> can be of length 3, 6 or 12.",
		},
	}
}

type clrRGB struct {
	r, g, b int
	clrRGB16
}

func (clrRGB) Name() string { return "rgb" }
func (clrRGB) Inline() bool { return true }

func (clr clrRGB) Encode(w pipeline.Writer) error {
	w.Int(clr.r)
	w.Int(clr.r)
	w.Int(clr.r)
	return nil
}

func (clr clrRGB) Decode(r pipeline.Reader) (pipeline.Element, error) {
	clr.r = r.Int()
	clr.g = r.Int()
	clr.b = r.Int()
	clr.clrRGB16 = RGB8(uint8(clr.r), uint8(clr.g), uint8(clr.b))
	return clr, nil
}

func (clr clrRGB) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<r> <g> <b>)", clr.Name()),
			"Creates a color from 0-255 rgb components.",
		},
	}
}

type clrRGB16 struct {
	c [3]uint16
}

func (clrRGB16) Name() string { return "rgb16" }
func (clrRGB16) Inline() bool { return true }

func (clr clrRGB16) Encode(w pipeline.Writer) error {
	w.Int(int(clr.c[0]))
	w.Int(int(clr.c[1]))
	w.Int(int(clr.c[2]))
	return nil
}

func (clr clrRGB16) Decode(r pipeline.Reader) (pipeline.Element, error) {
	clr.c[0] = uint16(r.Int())
	clr.c[1] = uint16(r.Int())
	clr.c[2] = uint16(r.Int())
	return clr, nil
}

func (clr clrRGB16) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<r> <g> <b>)", clr.Name()),
			"Creates a color from 0-65535 rgb components.",
		},
	}
}

func (clr clrRGB16) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(clr)

	return img, nil
}

func (clr clrRGB16) Color() [3]uint16 { return clr.c }
