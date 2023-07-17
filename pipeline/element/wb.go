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
			fmt.Sprintf("%s()", r.Name()),
			"TODO",
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
			fmt.Sprintf("%s()", r.Name()),
			"TODO",
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
