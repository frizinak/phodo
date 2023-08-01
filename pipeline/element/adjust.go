package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Contrast(n float64) pipeline.Element   { return contrast{n: pipeline.PlainNumber(n)} }
func Brightness(n float64) pipeline.Element { return brightness{n: pipeline.PlainNumber(n)} }
func Gamma(n float64) pipeline.Element      { return gamma{n: pipeline.PlainNumber(n)} }
func Saturation(n float64) pipeline.Element { return saturation{n: pipeline.PlainNumber(n)} }
func Black(n float64) pipeline.Element      { return black{n: pipeline.PlainNumber(n)} }

type contrast struct {
	n pipeline.Value
}

func (c contrast) Name() string { return "contrast" }
func (c contrast) Inline() bool { return true }

func (c contrast) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<factor>)", c.Name()),
			"Adjusts the contrast by the given <factor>.",
		},
	}
}

func (c contrast) Encode(w pipeline.Writer) error {
	w.Value(c.n)
	return nil
}

func (c contrast) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.n = r.Value()
	return c, nil
}

func (c contrast) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	n, err := c.n.Float64(img)
	if err != nil {
		return img, err
	}

	core.Contrast(img, n)

	return img, nil
}

type brightness struct {
	n pipeline.Value
}

func (b brightness) Name() string { return "brightness" }
func (b brightness) Inline() bool { return true }

func (b brightness) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<factor>)", b.Name()),
			"Adjusts the brightness by the given <factor>.",
		},
	}
}

func (b brightness) Encode(w pipeline.Writer) error {
	w.Value(b.n)
	return nil
}

func (b brightness) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.n = r.Value()
	return b, nil
}

func (b brightness) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	n, err := b.n.Float64(img)
	if err != nil {
		return img, err
	}

	core.Brightness(img, n)

	return img, nil
}

type gamma struct {
	n pipeline.Value
}

func (g gamma) Name() string { return "gamma" }
func (g gamma) Inline() bool { return true }

func (g gamma) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<factor>)", g.Name()),
			"Adjusts the gamma by the given <factor>.",
		},
	}
}

func (g gamma) Encode(w pipeline.Writer) error {
	w.Value(g.n)
	return nil
}

func (g gamma) Decode(r pipeline.Reader) (pipeline.Element, error) {
	g.n = r.Value()
	return g, nil
}

func (g gamma) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(g)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(g.Name())
	}

	n, err := g.n.Float64(img)
	if err != nil {
		return img, err
	}

	core.Gamma(img, n)

	return img, nil
}

type saturation struct {
	n pipeline.Value
}

func (s saturation) Name() string { return "saturation" }
func (s saturation) Inline() bool { return true }

func (s saturation) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<factor>)", s.Name()),
			"Adjusts the saturation by the given <factor>",
		},
	}
}

func (s saturation) Encode(w pipeline.Writer) error {
	w.Value(s.n)
	return nil
}

func (s saturation) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.n = r.Value()
	return s, nil
}

func (s saturation) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(s.Name())
	}

	n, err := s.n.Float64(img)
	if err != nil {
		return img, err
	}

	core.Saturation(img, n)

	return img, nil
}

type black struct {
	n pipeline.Value
}

func (b black) Name() string { return "black" }
func (b black) Inline() bool { return true }

func (b black) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<factor>)", b.Name()),
			"Adjusts the black point by the given <factor>.",
		},
	}
}

func (b black) Encode(w pipeline.Writer) error {
	w.Value(b.n)
	return nil
}

func (b black) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.n = r.Value()
	return b, nil
}

func (b black) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	n, err := b.n.Float64(img)
	if err != nil {
		return img, err
	}

	core.Black(img, n)

	return img, nil
}
