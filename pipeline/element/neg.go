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

func InvertFilm(r, g, b float64) pipeline.Element {
	return invertFilm{
		r: pipeline.PlainNumber(r),
		g: pipeline.PlainNumber(g),
		b: pipeline.PlainNumber(b),
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

type invertFilm struct {
	r, g, b pipeline.Value
}

func (i invertFilm) Name() string {
	return "invert-film"
}
func (invertFilm) Inline() bool { return true }

func (i invertFilm) Encode(w pipeline.Writer) error {
	w.Value(i.r)
	w.Value(i.g)
	w.Value(i.b)
	return nil
}

func (i invertFilm) Decode(r pipeline.Reader) (interface{}, error) {
	i.r = r.Value()
	i.g = r.Value()
	i.b = r.Value()
	return i, nil
}

func (i invertFilm) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<r> <g> <b>)", i.Name()),
			"Adjusts rgb components of the current image by raising them",
		},
		{
			"",
			"to exponents defined by the given rgb parameters. Green being the",
		},
		{
			"",
			"reference exponent and the red and blue parameters as multipliers",
		},
		{
			"",
			"to this reference exponent.",
		},
		{
			"",
			"A good starting point would be 1.18, 1.5 and 0.92 respectively.",
		},
	}
}

func (i invertFilm) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(i)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(i.Name())
	}

	r, err := i.r.Float64(img)
	if err != nil {
		return img, err
	}
	g, err := i.g.Float64(img)
	if err != nil {
		return img, err
	}
	b, err := i.b.Float64(img)
	if err != nil {
		return img, err
	}

	core.InvertFilm(img, r, g, b)

	return img, nil
}
