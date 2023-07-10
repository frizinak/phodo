package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Contrast(n float64) pipeline.Element   { return contrast{n: n} }
func Brightness(n float64) pipeline.Element { return brightness{n: n} }
func Gamma(n float64) pipeline.Element      { return gamma{n: n} }
func Saturation(n float64) pipeline.Element { return saturation{n: n} }
func Black(n float64) pipeline.Element      { return black{n: n} }

type contrast struct {
	n float64
}

func (c contrast) Name() string { return "contrast" }
func (c contrast) Inline() bool { return true }

func (c contrast) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", c.Name()),
			"TODO",
		},
	}
}

func (c contrast) Encode(w pipeline.Writer) error {
	w.Float(c.n)
	return nil
}

func (c contrast) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.n = r.Float(0)
	return c, nil
}

func (c contrast) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	core.Contrast(img, c.n)

	return img, nil
}

type brightness struct {
	n float64
}

func (b brightness) Name() string { return "brightness" }
func (b brightness) Inline() bool { return true }

func (b brightness) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", b.Name()),
			"TODO",
		},
	}
}

func (b brightness) Encode(w pipeline.Writer) error {
	w.Float(b.n)
	return nil
}

func (b brightness) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.n = r.Float(0)
	return b, nil
}

func (b brightness) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	core.Brightness(img, b.n)

	return img, nil
}

type gamma struct {
	n float64
}

func (g gamma) Name() string { return "gamma" }
func (g gamma) Inline() bool { return true }

func (g gamma) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", g.Name()),
			"TODO",
		},
	}
}

func (g gamma) Encode(w pipeline.Writer) error {
	w.Float(g.n)
	return nil
}

func (g gamma) Decode(r pipeline.Reader) (pipeline.Element, error) {
	g.n = r.Float(0)
	return g, nil
}

func (g gamma) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(g)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(g.Name())
	}

	core.Gamma(img, g.n)

	return img, nil
}

type saturation struct {
	n float64
}

func (s saturation) Name() string { return "saturation" }
func (s saturation) Inline() bool { return true }

func (s saturation) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", s.Name()),
			"TODO",
		},
	}
}

func (s saturation) Encode(w pipeline.Writer) error {
	w.Float(s.n)
	return nil
}

func (s saturation) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.n = r.Float(0)
	return s, nil
}

func (s saturation) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(s.Name())
	}

	core.Saturation(img, s.n)

	return img, nil
}

type black struct {
	n float64
}

func (b black) Name() string { return "black" }
func (b black) Inline() bool { return true }

func (b black) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", b.Name()),
			"TODO",
		},
	}
}

func (b black) Encode(w pipeline.Writer) error {
	w.Float(b.n)
	return nil
}

func (b black) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.n = r.Float(0)
	return b, nil
}

func (b black) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	core.Black(img, b.n)

	return img, nil
}
