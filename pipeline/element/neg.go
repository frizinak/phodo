package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Invert(r, g, b float64) pipeline.Element {
	return invert{
		pipeline.PlainNumber(r),
		pipeline.PlainNumber(g),
		pipeline.PlainNumber(b),
	}
}

type invert struct {
	r, g, b pipeline.Value
}

func (i invert) Name() string { return "invert" }
func (invert) Inline() bool   { return true }

func (i invert) Encode(w pipeline.Writer) error {
	w.Value(i.r)
	w.Value(i.g)
	w.Value(i.b)
	return nil
}

func (i invert) Decode(r pipeline.Reader) (interface{}, error) {
	i.r = r.ValueDefault(pipeline.PlainNumber(1))
	i.g = r.ValueDefault(pipeline.PlainNumber(1))
	i.b = r.ValueDefault(pipeline.PlainNumber(1))
	return i, nil
}

func (i invert) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<r> <g> <b>)", i.Name()),
			"Inverts each channel by the given [0-1] amount.",
		},
	}
}

func (i invert) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(i)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(i.Name())
	}

	r, err := i.r.Float64(img)
	if err != nil {
		return nil, err
	}
	g, err := i.g.Float64(img)
	if err != nil {
		return nil, err
	}
	b, err := i.b.Float64(img)
	if err != nil {
		return nil, err
	}

	core.Invert(img, r, g, b)

	return img, nil
}
