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
	n pipeline.Number
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
	w.Number(c.n)
	return nil
}

func (c contrast) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.n = r.Number()
	return c, nil
}

func (c contrast) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	n, err := c.n.Execute(img)
	if err != nil {
		return img, err
	}

	core.Contrast(img, n)

	return img, nil
}

type brightness struct {
	n pipeline.Number
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
	w.Number(b.n)
	return nil
}

func (b brightness) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.n = r.Number()
	return b, nil
}

func (b brightness) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	n, err := b.n.Execute(img)
	if err != nil {
		return img, err
	}

	core.Brightness(img, n)

	return img, nil
}

type gamma struct {
	n pipeline.Number
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
	w.Number(g.n)
	return nil
}

func (g gamma) Decode(r pipeline.Reader) (pipeline.Element, error) {
	g.n = r.Number()
	return g, nil
}

func (g gamma) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(g)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(g.Name())
	}

	n, err := g.n.Execute(img)
	if err != nil {
		return img, err
	}

	core.Gamma(img, n)

	return img, nil
}

type saturation struct {
	n pipeline.Number
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
	w.Number(s.n)
	return nil
}

func (s saturation) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.n = r.Number()
	return s, nil
}

func (s saturation) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(s.Name())
	}

	n, err := s.n.Execute(img)
	if err != nil {
		return img, err
	}

	core.Saturation(img, n)

	return img, nil
}

type black struct {
	n pipeline.Number
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
	w.Number(b.n)
	return nil
}

func (b black) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.n = r.Number()
	return b, nil
}

func (b black) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	n, err := b.n.Execute(img)
	if err != nil {
		return img, err
	}

	core.Black(img, n)

	return img, nil
}
