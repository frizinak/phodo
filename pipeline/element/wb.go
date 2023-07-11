package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func RGBAdd(r, g, b int) pipeline.Element     { return rgbAdd{r, g, b} }
func RGBMul(r, g, b float64) pipeline.Element { return rgbMul{r, g, b} }

type rgbAdd struct{ r, g, b int }

func (rgbAdd) Name() string { return "rgb-add" }
func (rgbAdd) Inline() bool { return true }

func (r rgbAdd) Encode(w pipeline.Writer) error {
	w.Int(r.r)
	w.Int(r.g)
	w.Int(r.b)
	return nil
}

func (rgb rgbAdd) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rgb.r = r.Int()
	rgb.g = r.Int()
	rgb.b = r.Int()
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

	core.RGBAdd(img, rgb.r, rgb.g, rgb.b)

	return img, nil
}

type rgbMul struct{ r, g, b float64 }

func (rgbMul) Name() string { return "rgb-multiply" }
func (rgbMul) Inline() bool { return true }

func (r rgbMul) Encode(w pipeline.Writer) error {
	w.Float(r.r)
	w.Float(r.g)
	w.Float(r.b)
	return nil
}

func (rgb rgbMul) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rgb.r = r.Float()
	rgb.g = r.Float()
	rgb.b = r.Float()
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

	core.RGBMultiply(img, rgb.r, rgb.g, rgb.b)

	return img, nil
}
