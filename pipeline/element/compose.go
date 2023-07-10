package element

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func NewPos(e pipeline.Element, coords image.Point) Pos {
	return Pos{e, coords}
}

func NewPosBlackTransparent(e pipeline.Element, coords image.Point) PosBlackTransparent {
	return PosBlackTransparent{Pos{e, coords}}
}

func Compose(in ...Positionable) pipeline.Element {
	return compose{in}
}

type PosBlackTransparent struct{ Pos }

func (PosBlackTransparent) Transparent(r, g, b uint16) bool {
	return r == 0 && g == 0 && b == 0
}

func (PosBlackTransparent) Name() string { return "pos-alpha" }

func (p PosBlackTransparent) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", p.Name()),
			"TODO",
		},
	}
}

func (p PosBlackTransparent) Decode(r pipeline.Reader) (pipeline.Element, error) {
	pos, err := p.Pos.Decode(r)
	if err == nil {
		p.Pos = pos.(Pos)
	}
	return p, err
}

type Pos struct {
	el pipeline.Element
	p  image.Point
}

func (p Pos) Point() image.Point        { return p.p }
func (p Pos) Element() pipeline.Element { return p.el }

func (Pos) Name() string { return "pos" }
func (Pos) Inline() bool { return true }

func (p Pos) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", p.Name()),
			"TODO",
		},
	}
}

func (p Pos) Encode(w pipeline.Writer) error {
	w.Int(p.p.X)
	w.Int(p.p.Y)
	return w.Element(p.el)
}

func (p Pos) Decode(r pipeline.Reader) (pipeline.Element, error) {
	p.p.X = r.Int(0)
	p.p.Y = r.Int(1)
	var err error
	p.el, err = r.Element(2)
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
			fmt.Sprintf("%s()", c.Name()),
			"TODO",
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
	Point() image.Point
	Element() pipeline.Element
}

type Transparent interface {
	Transparent(r, g, b uint16) bool
}

func (c compose) Decode(r pipeline.Reader) (pipeline.Element, error) {
	l := r.Len()
	c.items = make([]Positionable, l)
	for i := 0; i < l; i++ {
		el, err := r.Element(i)
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

func (c compose) do(p Positionable, src, dst *img48.Img) {
	d := p.Point()
	trans, _ := p.(Transparent)
	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		so_ := (y - src.Rect.Min.Y) * src.Stride
		do_ := (y + d.Y - src.Rect.Min.Y) * dst.Stride
		for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
			so := so_ + (x-src.Rect.Min.X)*3
			do := do_ + (x+d.X-src.Rect.Min.X)*3
			p := src.Pix[so : so+3 : so+3]
			if do < len(dst.Pix) && (trans == nil ||
				!trans.Transparent(p[0], p[1], p[2])) {
				copy(dst.Pix[do:do+3:do+3], p)
			}
		}
	}
}

func (c compose) Do(ctx pipeline.Context, dst *img48.Img) (*img48.Img, error) {
	imgs, err := c.doall(ctx, dst)
	if err != nil {
		return dst, err
	}

	ctx.Mark(c)

	// TODO make do without a dst image
	for i, src := range imgs {
		c.do(c.items[i], src, dst)
	}
	return dst, nil
}
