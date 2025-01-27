package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Contrast(n float64) pipeline.Element   { return contrast{n: pipeline.PlainNumber(n)} }
func ContrastY(n float64) pipeline.Element  { return contrastY{n: pipeline.PlainNumber(n)} }
func Brightness(n float64) pipeline.Element { return brightness{n: pipeline.PlainNumber(n)} }
func Gamma(n float64) pipeline.Element      { return gamma{n: pipeline.PlainNumber(n)} }
func Saturation(n float64) pipeline.Element { return saturation{n: pipeline.PlainNumber(n)} }
func Black(n float64) pipeline.Element      { return black{n: pipeline.PlainNumber(n)} }
func Eq(ns ...float64) pipeline.Element {
	l := make([]pipeline.Value, len(ns))
	for i := range l {
		l[i] = pipeline.PlainNumber(ns[i])
	}
	return eq{ns: l}
}

func AutoGamma(mid float64) pipeline.Element {
	return autogamma{
		mid: pipeline.PlainNumber(mid),
	}
}

func AutoGammaRegion(x, y, w, h float64, mid float64) pipeline.Element {
	return autogamma{
		x:   pipeline.PlainNumber(x),
		y:   pipeline.PlainNumber(y),
		w:   pipeline.PlainNumber(w),
		h:   pipeline.PlainNumber(h),
		mid: pipeline.PlainNumber(mid),
	}
}

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

func (c contrast) Decode(r pipeline.Reader) (interface{}, error) {
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

type contrastY struct {
	n pipeline.Value
}

func (c contrastY) Name() string { return "contrast-luminance" }
func (c contrastY) Inline() bool { return true }

func (c contrastY) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<factor>)", c.Name()),
			"Adjusts the luminance contrast by the given <factor>.",
		},
	}
}

func (c contrastY) Encode(w pipeline.Writer) error {
	w.Value(c.n)
	return nil
}

func (c contrastY) Decode(r pipeline.Reader) (interface{}, error) {
	c.n = r.Value()
	return c, nil
}

func (c contrastY) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	n, err := c.n.Float64(img)
	if err != nil {
		return img, err
	}

	core.ContrastY(img, n)

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

func (b brightness) Decode(r pipeline.Reader) (interface{}, error) {
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

func (g gamma) Decode(r pipeline.Reader) (interface{}, error) {
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

func (s saturation) Decode(r pipeline.Reader) (interface{}, error) {
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

func (b black) Decode(r pipeline.Reader) (interface{}, error) {
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

type eq struct {
	ns []pipeline.Value
}

func (eq eq) Name() string { return "eq" }
func (eq eq) Inline() bool { return true }

func (eq eq) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([factor1] [factor2] ...[factorN])", eq.Name()),
			"Adjust image levels by the given factors.",
		},
	}
}

func (eq eq) Encode(w pipeline.Writer) error {
	for _, v := range eq.ns {
		w.Value(v)
	}
	return nil
}

func (eq eq) Decode(r pipeline.Reader) (interface{}, error) {
	l := make([]pipeline.Value, r.Len())
	for i := range l {
		l[i] = r.Value()
	}
	eq.ns = l
	return eq, nil
}

func (eq eq) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(eq)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(eq.Name())
	}

	l := make([]float64, len(eq.ns))
	for i, v := range eq.ns {
		n, err := v.Float64(img)
		if err != nil {
			return img, err
		}
		l[i] = n
	}

	core.Eq(img, l...)

	return img, nil
}

type autogamma struct {
	x, y pipeline.Value
	w, h pipeline.Value
	mid  pipeline.Value
}

func (g autogamma) Name() string { return "autogamma" }
func (g autogamma) Inline() bool { return true }

func (g autogamma) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([mid-point])", g.Name()),
			"",
		},
		{
			fmt.Sprintf("%s(<x> <y> <w> <h> [mid-point])", g.Name()),
			"Center the histogram within the given region around the given",
		},
		{
			"",
			"mid-point (default: 0.5)",
		},
	}
}

func (g autogamma) Encode(w pipeline.Writer) error {
	if g.x != nil {
		w.Value(g.x)
		w.Value(g.y)
		w.Value(g.w)
		w.Value(g.h)
	}

	if g.mid != nil {
		w.Value(g.mid)
	}

	return nil
}

func (g autogamma) Decode(r pipeline.Reader) (interface{}, error) {
	if r.Len() > 1 {
		g.x = r.Value()
		g.y = r.Value()
		g.w = r.Value()
		g.h = r.Value()
	}

	g.mid = r.ValueDefault(pipeline.PlainNumber(0.5))
	return g, nil
}

func (g autogamma) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(g)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(g.Name())
	}

	rect := img.Rect
	if g.x != nil {
		x, err := g.x.Int(img)
		if err != nil {
			return img, err
		}
		y, err := g.y.Int(img)
		if err != nil {
			return img, err
		}
		w, err := g.w.Int(img)
		if err != nil {
			return img, err
		}
		h, err := g.h.Int(img)
		if err != nil {
			return img, err
		}

		rect.Min.X += x
		rect.Min.Y += y
		rect.Max.X = rect.Min.X + w
		rect.Max.Y = rect.Min.Y + h
	}

	mid, err := g.mid.Float64(img)
	if err != nil {
		return img, err
	}

	g1, g2 := core.AutoGamma(img, rect, mid)

	ctx.Print(g, fmt.Sprintf("gamma(%.4f) gamma(%.4f)", g1, g2))

	return img, nil
}
