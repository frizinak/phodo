package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func CLUT(e pipeline.Element, strength float64, iterations int) pipeline.Element {
	return clut{e: e, strength: pipeline.PlainNumber(strength), iterations: pipeline.PlainNumber(iterations)}
}

type clut struct {
	e          pipeline.Element
	iterations pipeline.Value
	strength   pipeline.Value
}

func (clut) Name() string { return "clut" }

func (c clut) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<element> [strength] [iterations])", c.Name()),
			"Hald CLUT. Executes the given <element> and uses it as a color",
		},
		{
			"",
			"lookup table for the input image. <strength> [0-1] determines how",
		},
		{
			"",
			"much of the original color is interpolated with the clut color.",
		},
	}
}

func (c clut) Encode(w pipeline.Writer) error {
	err := w.Element(c.e)
	w.Value(c.strength)
	w.Value(c.iterations)
	return err
}

func (c clut) Decode(r pipeline.Reader) (interface{}, error) {
	c.e = r.Element()
	c.strength = r.ValueDefault(pipeline.PlainNumber(1))
	c.iterations = r.ValueDefault(pipeline.PlainNumber(1))

	return c, nil
}

func (c clut) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	clut, err := c.e.Do(ctx, img)
	if err != nil {
		return img, err
	}

	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	strength, err := c.strength.Float64(img)
	if err != nil {
		return img, err
	}

	iterations, err := c.iterations.Int(img)
	if err != nil {
		return img, err
	}

	return img, core.CLUT(img, clut, strength, iterations)
}
