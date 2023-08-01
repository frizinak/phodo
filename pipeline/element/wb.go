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

func WhiteBalanceSpot(x, y, r int) pipeline.Element {
	return whiteBalanceSpot{
		x: pipeline.PlainNumber(x),
		y: pipeline.PlainNumber(y),
		r: pipeline.PlainNumber(r),
	}
}

type rgbAdd struct{ r, g, b pipeline.Value }

func (rgbAdd) Name() string { return "rgb-add" }
func (rgbAdd) Inline() bool { return true }

func (r rgbAdd) Encode(w pipeline.Writer) error {
	w.Value(r.r)
	w.Value(r.g)
	w.Value(r.b)
	return nil
}

func (rgb rgbAdd) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rgb.r = r.Value()
	rgb.g = r.Value()
	rgb.b = r.Value()
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

	r, err := rgb.r.Int(img)
	if err != nil {
		return img, err
	}
	g, err := rgb.g.Int(img)
	if err != nil {
		return img, err
	}
	b, err := rgb.b.Int(img)
	if err != nil {
		return img, err
	}

	core.RGBAdd(img, r, g, b)

	return img, nil
}

type rgbMul struct{ r, g, b pipeline.Value }

func (rgbMul) Name() string { return "rgb-multiply" }
func (rgbMul) Inline() bool { return true }

func (r rgbMul) Encode(w pipeline.Writer) error {
	w.Value(r.r)
	w.Value(r.g)
	w.Value(r.b)
	return nil
}

func (rgb rgbMul) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rgb.r = r.Value()
	rgb.g = r.Value()
	rgb.b = r.Value()
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

	r, err := rgb.r.Float64(img)
	if err != nil {
		return img, err
	}
	g, err := rgb.g.Float64(img)
	if err != nil {
		return img, err
	}
	b, err := rgb.b.Float64(img)
	if err != nil {
		return img, err
	}

	core.RGBMultiply(img, r, g, b)

	return img, nil
}

type whiteBalanceSpot struct {
	x, y pipeline.Value
	r    pipeline.Value
}

func (whiteBalanceSpot) Name() string { return "white-balance-spot" }
func (whiteBalanceSpot) Inline() bool { return true }

func (wb whiteBalanceSpot) Encode(w pipeline.Writer) error {
	w.Value(wb.x)
	w.Value(wb.y)
	w.Value(wb.r)
	return nil
}

func (wb whiteBalanceSpot) Decode(r pipeline.Reader) (pipeline.Element, error) {
	wb.x = r.Value()
	wb.y = r.Value()
	wb.r = r.Value()
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

	x, err := wb.x.Int(img)
	if err != nil {
		return img, err
	}
	y, err := wb.y.Int(img)
	if err != nil {
		return img, err
	}
	radius, err := wb.r.Int(img)
	if err != nil {
		return img, err
	}

	r, g, b := core.WhiteBalanceCalc(img, x, y, radius)

	core.RGBMultiply(img, r, g, b)

	return img, nil
}
