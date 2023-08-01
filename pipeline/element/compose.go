package element

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func NewPos(x, y int, e pipeline.Element) Pos {
	return Pos{e, Point{X: pipeline.PlainNumber(x), Y: pipeline.PlainNumber(y)}}
}

func NewPosTransparent(x, y int, e pipeline.Element, trans Color) PosTransparent {
	return PosTransparent{Pos: NewPos(x, y, e), clr: trans, _clr: trans.Color()}
}

func Compose(in ...Positionable) pipeline.Element {
	return compose{in}
}

type PosTransparent struct {
	Pos
	clr  Color
	_clr [3]uint16
}

func (p PosTransparent) Transparent(r, g, b uint16) bool {
	return p._clr[0] == r && p._clr[1] == g && p._clr[2] == b
}

func (PosTransparent) Name() string { return "pos-alpha" }

func (p PosTransparent) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <color> <element>)", p.Name()),
			"Same as pos() but don't copy pixels of the specified color.",
		},
	}
}

func (p PosTransparent) Encode(w pipeline.Writer) error {
	w.Number(p.p.X)
	w.Number(p.p.Y)
	if err := w.Element(p.clr); err != nil {
		return err
	}

	return w.Element(p.el)
}

func (p PosTransparent) Decode(r pipeline.Reader) (pipeline.Element, error) {
	p.p.X = r.Number()
	p.p.Y = r.Number()

	clr, err := r.Element()
	if err != nil {
		return p, err
	}

	var ok bool
	p.clr, ok = clr.(Color)
	if !ok {
		return p, fmt.Errorf("element of type '%T' is not a Color", clr)
	}
	p._clr = p.clr.Color()

	p.el, err = r.Element()
	if err != nil {
		return p, err
	}

	return p, err
}

type Pos struct {
	el pipeline.Element
	p  Point
}

type Point struct {
	X, Y pipeline.Number
}

func (p Point) Execute(img *img48.Img) (image.Point, error) {
	var pt image.Point
	var err error
	pt.X, err = p.X.Int(img)
	if err != nil {
		return pt, err
	}
	pt.Y, err = p.Y.Int(img)
	return pt, err
}

func (p Pos) Point() Point              { return p.p }
func (p Pos) Element() pipeline.Element { return p.el }

func (Pos) Name() string { return "pos" }
func (Pos) Inline() bool { return true }

func (p Pos) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <element>)", p.Name()),
			"Assigns an x and y coordinate to the given element.",
		},
	}
}

func (p Pos) Encode(w pipeline.Writer) error {
	w.Number(p.p.X)
	w.Number(p.p.Y)
	return w.Element(p.el)
}

func (p Pos) Decode(r pipeline.Reader) (pipeline.Element, error) {
	p.p.X = r.Number()
	p.p.Y = r.Number()
	var err error
	p.el, err = r.Element()
	return p, err
}

func (p Pos) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	return p.el.Do(ctx, img)
}

type compose struct {
	items []Positionable
}

func (c compose) Name() string { return "compose" }

func (c compose) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([pos-element1] [pos-element2] ...[pos-elementN])", c.Name()),
			"Draws the given pos-elements on the input image. See pos() below.",
		},
	}
}

func (c compose) Encode(w pipeline.Writer) error {
	for _, e := range c.items {
		enc, ok := e.(pipeline.Element)
		if !ok {
			return fmt.Errorf("compose subelement of type '%T' is not a pipeline Element", e)
		}
		if err := w.Element(enc); err != nil {
			return err
		}
	}

	return nil
}

type Positionable interface {
	Point() Point
	Element() pipeline.Element
}

type Transparent interface {
	Transparent(r, g, b uint16) bool
}

func (c compose) Decode(r pipeline.Reader) (pipeline.Element, error) {
	l := r.Len()
	c.items = make([]Positionable, l)
	for i := 0; i < l; i++ {
		el, err := r.Element()
		if err != nil {
			return nil, err
		}
		var ok bool
		c.items[i], ok = el.(Positionable)
		if !ok {
			return c, fmt.Errorf("wrong type '%T' as argument %d to compose", el, i+1)
		}
	}

	return c, nil
}

func (c compose) doall(ctx pipeline.Context, img *img48.Img) ([]*img48.Img, error) {
	is := make([]*img48.Img, len(c.items))
	for i, e := range c.items {
		im, err := e.Element().Do(ctx, img)
		if err != nil {
			return is, err
		}
		is[i] = im
	}
	return is, nil
}

func (c compose) do(p Positionable, src, dst *img48.Img) error {
	var t func(r, g, b uint16) bool
	trans, ok := p.(Transparent)
	if ok {
		t = trans.Transparent
	}

	pnt, err := p.Point().Execute(dst)
	if err != nil {
		return err
	}
	core.Draw(src, dst, pnt, t)
	return nil
}

func (c compose) Do(ctx pipeline.Context, dst *img48.Img) (*img48.Img, error) {
	imgs, err := c.doall(ctx, dst)
	if err != nil {
		return dst, err
	}

	ctx.Mark(c)

	if dst == nil {
		return dst, pipeline.NewErrNeedImageInput(c.Name())
	}

	for i, src := range imgs {
		if err := c.do(c.items[i], src, dst); err != nil {
			return dst, err
		}

	}

	return dst, nil
}
