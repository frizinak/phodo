package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func CLUT(e pipeline.Element) pipeline.Element { return clut{e: e} }

type clut struct {
	e      pipeline.Element
	amount pipeline.Number
}

func (clut) Name() string { return "clut" }

func (c clut) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<element>, <amount>)", c.Name()),
			"Hald CLUT. Executes the given <element> and uses it as a color",
		},
		{
			"",
			"lookup table for the input image. <amount> [0-1] determines how",
		},
		{
			"",
			"much of the original color is interpolated with the clut color.",
		},
		{
			"",
			"An amount of 1 results in the best performance as no values need",
		},
		{
			"",
			"to be interpolated.",
		},
	}
}

func (c clut) Encode(w pipeline.Writer) error {
	err := w.Element(c.e)
	w.Number(c.amount)
	return err
}

func (c clut) Decode(r pipeline.Reader) (pipeline.Element, error) {
	var err error

	c.e, err = r.Element()
	c.amount = r.NumberDefault(1)

	return c, err
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

	amount, err := c.amount.Execute(img)
	if err != nil {
		return img, err
	}

	return img, core.CLUT(img, clut, amount)
}
