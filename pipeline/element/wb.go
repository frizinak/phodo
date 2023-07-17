package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func RGBAdd(r, g, b int) pipeline.Element {
	return rgbAdd{
		pipeline.PlainNumber(r),
		pipeline.PlainNumber(g),
		pipeline.PlainNumber(b),
	}
}

func RGBMul(r, g, b float64) pipeline.Element {
	return rgbMul{
		pipeline.PlainNumber(r),
		pipeline.PlainNumber(g),
		pipeline.PlainNumber(b),
	}
}

type rgbAdd struct{ r, g, b pipeline.Number }

func (rgbAdd) Name() string { return "rgb-add" }
func (rgbAdd) Inline() bool { return true }

func (r rgbAdd) Encode(w pipeline.Writer) error {
	w.Number(r.r)
	w.Number(r.g)
	w.Number(r.b)
	return nil
}

func (rgb rgbAdd) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rgb.r = r.Number()
	rgb.g = r.Number()
	rgb.b = r.Number()
	return rgb, nil
}

func (r rgbAdd) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<r> <g> <b>)", r.Name()),
			"Adjusts rgb components of the current image by adding the given",
		},
		{
			"",
			"0-65535 <r> <g> and <b> values.",
		},
	}
}

func (rgb rgbAdd) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(rgb)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(rgb.Name())
	}

	_r, err := rgb.r.Execute(img)
	if err != nil {
		return img, err
	}
	_g, err := rgb.g.Execute(img)
	if err != nil {
		return img, err
	}
	_b, err := rgb.b.Execute(img)
	if err != nil {
		return img, err
	}

	r, g, b := int(_r), int(_g), int(_b)

	core.RGBAdd(img, r, g, b)

	return img, nil
}

type rgbMul struct{ r, g, b pipeline.Number }

func (rgbMul) Name() string { return "rgb-multiply" }
func (rgbMul) Inline() bool { return true }

func (r rgbMul) Encode(w pipeline.Writer) error {
	w.Number(r.r)
	w.Number(r.g)
	w.Number(r.b)
	return nil
}

func (rgb rgbMul) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rgb.r = r.Number()
	rgb.g = r.Number()
	rgb.b = r.Number()
	return rgb, nil
}

func (r rgbMul) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<r> <g> <b>)", r.Name()),
			"Adjusts rgb components of the current image by multiplying them",
		},
		{
			"",
			"with the given <r> <g> and <b> multipliers.",
		},
	}
}

func (rgb rgbMul) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(rgb)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(rgb.Name())
	}

	r, err := rgb.r.Execute(img)
	if err != nil {
		return img, err
	}
	g, err := rgb.g.Execute(img)
	if err != nil {
		return img, err
	}
	b, err := rgb.b.Execute(img)
	if err != nil {
		return img, err
	}

	core.RGBMultiply(img, r, g, b)

	return img, nil
}

type whiteBalanceSpot struct {
	x, y pipeline.Number
	r    pipeline.Number
}

func (whiteBalanceSpot) Name() string { return "white-balance-spot" }
func (whiteBalanceSpot) Inline() bool { return true }

func (wb whiteBalanceSpot) Encode(w pipeline.Writer) error {
	w.Number(wb.x)
	w.Number(wb.y)
	w.Number(wb.r)
	return nil
}

func (wb whiteBalanceSpot) Decode(r pipeline.Reader) (pipeline.Element, error) {
	wb.x = r.Number()
	wb.y = r.Number()
	wb.r = r.Number()
	return wb, nil
}

func (wb whiteBalanceSpot) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <r>)", wb.Name()),
			"Adjusts the whitebalance by assuming the average of the color of",
		},
		{
			"",
			"the pixels at <x> <y> represents a perfectly grey spot.",
		},
	}
}

func (wb whiteBalanceSpot) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(wb)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(wb.Name())
	}

	_x, err := wb.x.Execute(img)
	if err != nil {
		return img, err
	}
	_y, err := wb.y.Execute(img)
	if err != nil {
		return img, err
	}
	_r, err := wb.r.Execute(img)
	if err != nil {
		return img, err
	}

	cx, cy := int(_x), int(_y)

	var r, g, b uint64
	var n uint64
	o := func(x1, x2, y int) {
		if y >= img.Rect.Max.Y {
			y = img.Rect.Max.Y - 1
		}
		if x1 >= img.Rect.Max.X {
			x1 = img.Rect.Max.X - 1
		}
		if x2 >= img.Rect.Max.X {
			x2 = img.Rect.Max.X - 1
		}

		x1 = (x1 - img.Rect.Min.X)
		x2 = (x2 - img.Rect.Min.X)
		y = y - img.Rect.Min.Y
		if y < 0 {
			y = 0
		}
		if x1 < 0 {
			x1 = 0
		}
		if x2 < 0 {
			x2 = 0
		}

		o := y * img.Stride
		o1 := o + x1*3
		o2 := o + x2*3
		pix := img.Pix[o1 : o2+3 : o2+3]
		for i := 0; i < len(pix); i += 3 {
			n++
			r += uint64(pix[i+0])
			g += uint64(pix[i+1])
			b += uint64(pix[i+2])
		}
	}

	x := int(_r)
	y := 0
	e := 0

	for x >= y {
		o(cx-x, cx+x, cy+y)
		o(cx-x, cx+x, cy-y)
		o(cx-y, cx+y, cy+x)
		o(cx-y, cx+y, cy-x)
		if e <= 0 {
			y += 1
			e += 2*y + 1
		}

		if e > 0 {
			x -= 1
			e -= 2*x + 1
		}
	}

	if n == 0 {
		n = 1
	}

	r /= n
	g /= n
	b /= n
	avg := (r + g + b) / 3

	rm := float64(avg) / float64(r)
	gm := float64(avg) / float64(g)
	bm := float64(avg) / float64(b)

	return RGBMul(rm, gm, bm).Do(ctx, img)
}
