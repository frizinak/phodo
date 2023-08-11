package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Clipping(threshold float64, clr pipeline.ComplexValue) pipeline.Element {
	return clip{threshold: pipeline.PlainNumber(threshold), clr: clr}
}

type clip struct {
	channel   bool
	threshold pipeline.Value
	clr       pipeline.ComplexValue
}

type defaultColor struct {
	clrRGB16
}

func (c clip) Name() string {
	if c.channel {
		return "clipping-channel"
	}
	return "clipping"
}
func (clip) Inline() bool { return true }

func (c clip) Encode(w pipeline.Writer) error {
	w.Value(c.threshold)
	return w.ComplexValue(c.clr)
}

func (c clip) Decode(r pipeline.Reader) (interface{}, error) {
	c.threshold = r.Value()

	c.clr = r.ComplexValueDefault(defaultColor{RGB16(0, 0, 0)})

	return c, nil
}

func (c clip) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<threshold> [color])", c.Name()),
			"Visualises clipping pixels. i.e.: Pixels whose values are close",
		},
		{
			"",
			"to their min/max values, <threshold> determines how close the values",
		},
		{
			"",
			"have to be, as a percentage of the value range, to be visualised.",
		},
		{
			"",
			"Visualise clipping shadows: <threshold> <= 50%",
		},
		{
			"",
			"Visualise blown highlights: <threshold> > 50%",
		},
	}
}

func (c clip) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	th, err := c.threshold.Float64(img)
	if err != nil {
		return img, err
	}

	clrVal := c.clr
	if clrVal == nil {
		clrVal = defaultColor{RGB16(0, 0, 0)}
	}

	_clr, err := clrVal.Value(img)
	if err != nil {
		return img, err
	}

	clr, ok := _clr.(core.Color)
	if !ok {
		return img, fmt.Errorf("element of type '%T' is not a Color", _clr)
	}

	if _, ok := _clr.(defaultColor); ok {
		clr = core.SimpleColor{0, 0, 0}
		if th <= 0.5 {
			clr = core.SimpleColor{1<<16 - 1, 1<<16 - 1, 1<<16 - 1}
		}
	}

	core.DrawClipping(clr, img, th, c.channel)

	return img, nil
}
