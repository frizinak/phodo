package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

const (
	InterpolationNearest   = "nearest"
	InterpolationTrilinear = "trilinear"
)

var clutInterpolations = []string{
	InterpolationNearest,
	InterpolationTrilinear,
}

func CLUT(e pipeline.Element, strength float64, interpolation string) pipeline.Element {
	return clut{e: e, strength: pipeline.PlainNumber(strength), interp: pipeline.PlainString(interpolation)}
}

type clut struct {
	e        pipeline.Element
	strength pipeline.Value
	interp   pipeline.Value
}

func (c clut) Name() string { return "clut" }

func (c clut) Help() [][2]string {
	help := [][2]string{
		{
			fmt.Sprintf("%s(<element> [strength] [interpolation])", c.Name()),
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
		{
			"",
			"<interpolation> can be one of:",
		},
	}

	l := make([]string, 0, len(clutInterpolations))
	for _, o := range clutInterpolations {
		l = append(l, o)
	}
	for _, t := range l {
		help = append(help, [2]string{"", " - " + t})
	}

	return help
}

func (c clut) Encode(w pipeline.Writer) error {
	err := w.Element(c.e)
	w.Value(c.strength)
	w.Value(c.interp)
	return err
}

func (c clut) Decode(r pipeline.Reader) (interface{}, error) {
	c.e = r.Element()
	c.strength = r.ValueDefault(pipeline.PlainNumber(1))
	c.interp = r.ValueDefault(pipeline.PlainString(InterpolationTrilinear))

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

	interp, err := c.interp.String(img)
	if err != nil {
		return img, err
	}

	interpolate := interp != InterpolationNearest

	return img, core.CLUT(img, clut, strength, interpolate)
}
